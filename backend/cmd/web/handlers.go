package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrValueTooLong = errors.New("Cookie value too long")
	ErrInvalidValue = errors.New("Invalid cookie value")
)

type userMoveData struct {
	Fen   string
	Piece int
	Move  int
}

type getChessMoveData struct {
	Fen   string
	Piece int
}

type getChessMoveDataJSON struct {
	Moves            []int `json:"moves"`
	Captures         []int `json:"captures"`
	TriggerPromotion bool  `json:"triggerPromotion"`
}

type postChessMove struct {
	CurrentFEN      string `json:"currentFEN"`
	Piece           int    `json:"piece"`
	Move            int    `json:"move"`
	PromotionString string `json:"promotionString"`
}

type postChessMoveReply struct {
	MatchStateHistory   []MatchStateHistory `json:"pastMoves"`
	GameOverStatus      gameOverStatusCode  `json:"gameOverStatus"`
	ThreefoldRepetition bool                `json:"threefoldRepetition"`
}

type joinQueueRequest struct {
	TimeFormatInMilliseconds int64  `json:"timeFormatInMilliseconds"`
	IncrementInMilliseconds  int64  `json:"incrementInMilliseconds"`
	Action                   string `json:"action"`
}

func generateNewPlayerId() int64 {
	return rand.Int63()
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	app.clientError(w, http.StatusNotFound)
}

func getChessMovesHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() { app.perfLog.Printf("getChessMovesHandler took: %s\n", time.Since(start)) }()

	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Content-Type", "application/json")

	var chessMoveData getChessMoveData

	err := json.NewDecoder(r.Body).Decode(&chessMoveData)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.infoLog.Printf("Received body: %+v\n", chessMoveData)

	var currentGameState = boardFromFEN(chessMoveData.Fen)
	var moves, captures, triggerPromotion, _ = getValidMovesForPiece(chessMoveData.Piece, currentGameState)

	var data = getChessMoveDataJSON{Moves: moves, Captures: captures, TriggerPromotion: triggerPromotion}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Write(jsonStr)
}

func joinQueueHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() { app.perfLog.Printf("joinQueueHandler took: %s\n", time.Since(start)) }()

	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		app.clientError(w, http.StatusMethodNotAllowed)
	}

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	var joinQueue joinQueueRequest

	app.infoLog.Printf("%v\n", r.Body)

	err := json.NewDecoder(r.Body).Decode(&joinQueue)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.infoLog.Printf("Received body: %+v\n", joinQueue)
	app.infoLog.Println(r.Cookies())

	// Read cookie
	playerid, err := ReadSigned(r, app.secretKey, "playerid")

	app.infoLog.Println(err)

	// Generate new cookie if it does not exists
	if errors.Is(err, http.ErrNoCookie) && joinQueue.Action == "join" {
		playerid = strconv.FormatInt(generateNewPlayerId(), 10)

		cookie := http.Cookie{
			Name:     "playerid",
			Value:    playerid,
			Domain:   "localhost",
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
			Secure:   true,
		}

		err = WriteSigned(w, cookie, app.secretKey)
		if err != nil {
			app.serverError(w, err)
			return
		}
	} else if errors.Is(err, http.ErrNoCookie) && joinQueue.Action == "leave" {
		app.clientError(w, http.StatusBadRequest)
		return
	} else if err != nil {
		app.serverError(w, err)
	}

	app.infoLog.Printf("Player ID: %v\n", playerid)
	var playerIDasInt int64
	playerIDasInt, err = strconv.ParseInt(playerid, 10, 64)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if joinQueue.Action == "join" {
		addPlayerToWaitingPool(playerIDasInt, joinQueue.TimeFormatInMilliseconds, joinQueue.IncrementInMilliseconds)
	} else {
		// err = removePlayerFromQueue(playerIDasInt, joinQueue.Time, joinQueue.Increment)
		removePlayerFromWaitingPool(playerIDasInt, joinQueue.TimeFormatInMilliseconds, joinQueue.IncrementInMilliseconds)
	}

}

type Client struct {
	id      int64
	channel chan string
}

type Clients struct {
	mu      sync.Mutex
	clients map[int64]*Client
}

var clients = Clients{
	clients: make(map[int64]*Client),
}

func matchFoundSSEHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	playerid, err := ReadSigned(r, app.secretKey, "playerid")
	if err != nil {
		app.serverError(w, err)
	}

	var playerIDasInt int64

	playerIDasInt, err = strconv.ParseInt(playerid, 10, 64)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Do the channels get properly closed on a leave queue?
	// Do we send the proper message on a leave queue?
	clients.mu.Lock()
	_, ok := clients.clients[playerIDasInt]
	if !ok {
		clients.clients[playerIDasInt] = &Client{id: playerIDasInt, channel: make(chan string)}
	}
	clientChannel := clients.clients[playerIDasInt].channel
	clients.mu.Unlock()

	// Set appropriate headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	defer func() {
		clients.mu.Lock()
		delete(clients.clients, playerIDasInt)
		clients.mu.Unlock()
	}()

	defer app.liveMatches.EnQueueLogAll()

	for {
		select {
		case message, ok := <-clientChannel:
			if !ok {
				return
			}
			// Send the message to the client in SSE format
			fmt.Fprintf(w, "data: %s\n\n", message)
			// Flush the response to send the data to the client
			w.(http.Flusher).Flush()
		}
	}
}
