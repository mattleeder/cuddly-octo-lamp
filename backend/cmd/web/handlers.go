package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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

type playerCodeEnum int

const (
	WhitePieces = iota
	BlackPieces
	Spectator
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
	IsValid                  bool               `json:"isValid"`
	NewFEN                   string             `json:"newFEN"`
	LastMove                 [2]int             `json:"lastMove"`
	MoveWillTriggerPromotion bool               `json:"triggerPromotion"`
	GameOverStatus           gameOverStatusCode `json:"gameOverStatus"`
}

type joinQueueRequest struct {
	Time      int    `json:"time"`
	Increment int    `json:"increment"`
	Action    string `json:"action"`
}

func Write(w http.ResponseWriter, cookie http.Cookie) error {
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))

	if len(cookie.String()) > 4096 {
		return ErrValueTooLong
	}

	http.SetCookie(w, &cookie)

	return nil
}

func Read(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	value, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", ErrInvalidValue
	}

	return string(value), nil
}

func WriteSigned(w http.ResponseWriter, cookie http.Cookie, secretKey []byte) error {
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(cookie.Name))
	mac.Write([]byte(cookie.Value))
	signature := mac.Sum(nil)

	cookie.Value = string(signature) + cookie.Value

	return Write(w, cookie)
}

func ReadSigned(r *http.Request, secretKey []byte, name string) (string, error) {
	signedValue, err := Read(r, name)
	if err != nil {
		return "", err
	}

	if len(signedValue) < sha256.Size {
		return "", ErrInvalidValue
	}

	signature := signedValue[:sha256.Size]
	value := signedValue[sha256.Size:]

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	expectedMAC := mac.Sum(nil)

	if !hmac.Equal([]byte(signature), expectedMAC) {
		return "", ErrInvalidValue
	}

	return value, nil
}

func generateNewPlayerId() int64 {
	return rand.Int63()
}

func chessMoveValidationHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	start := time.Now()

	var chessMove userMoveData
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")

	err := json.NewDecoder(r.Body).Decode(&chessMove)
	if err != nil {
		http.Error(w, "Could not decode request", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Received body: %+v\n", chessMove)

	var validMove = IsMoveValid(chessMove.Fen, chessMove.Piece, chessMove.Move)

	fmt.Fprintln(w, validMove)

	elapsed := time.Since(start)
	fmt.Printf("Took %s\n", elapsed)
}

func getChessMovesHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	start := time.Now()
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	var chessMoveData getChessMoveData
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")

	err := json.NewDecoder(r.Body).Decode(&chessMoveData)
	if err != nil {
		http.Error(w, "Could not decode request", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Received body: %+v\n", chessMoveData)

	var currentGameState = boardFromFEN(chessMoveData.Fen)
	var moves, captures, triggerPromotion, _ = getValidMovesForPiece(chessMoveData.Piece, currentGameState)

	var data = getChessMoveDataJSON{Moves: moves, Captures: captures, TriggerPromotion: triggerPromotion}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonStr)

	elapsed := time.Since(start)
	fmt.Printf("Took %s\n", elapsed)
}

func postChessMoveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	start := time.Now()

	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	var chessMove postChessMove
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")

	err := json.NewDecoder(r.Body).Decode(&chessMove)
	if err != nil {
		http.Error(w, "Could not decode request", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Received body: %+v\n", chessMove)

	var validMove = IsMoveValid(chessMove.CurrentFEN, chessMove.Piece, chessMove.Move)

	var data postChessMoveReply
	if !validMove {
		data = postChessMoveReply{IsValid: false, NewFEN: "", LastMove: [2]int{0, 0}}
	} else {
		// Need to generate the new FEN
		newFEN, gameOverStatus := getFENAfterMove(chessMove.CurrentFEN, chessMove.Piece, chessMove.Move, chessMove.PromotionString)
		data = postChessMoveReply{IsValid: true, NewFEN: newFEN, LastMove: [2]int{chessMove.Piece, chessMove.Move}, GameOverStatus: gameOverStatus}
	}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonStr)
	fmt.Printf("Sending body: %+v\n", data)

	elapsed := time.Since(start)
	fmt.Printf("Took %s\n", elapsed)
}

func joinQueueHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	defer checkQueue()
	start := time.Now()

	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	var joinQueue joinQueueRequest
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	fmt.Printf("%v\n", r.Body)

	err := json.NewDecoder(r.Body).Decode(&joinQueue)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not decode join queue request: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Received body: %+v\n", joinQueue)
	fmt.Println(r.Cookies())

	// Read cookie
	playerid, err := ReadSigned(r, app.secretKey, "playerid")

	fmt.Println(err)

	// Generate new cookie if it does not exists
	if errors.Is(err, http.ErrNoCookie) && joinQueue.Action == "join" {
		playerid = strconv.FormatInt(generateNewPlayerId(), 10)

		cookie := http.Cookie{
			Name:     "playerid",
			Value:    playerid,
			Domain:   "localhost",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   true,
		}

		err = WriteSigned(w, cookie, app.secretKey)
		if err != nil {
			http.Error(w, "Unable to decode or write cookie", http.StatusInternalServerError)
			return
		}
	} else if errors.Is(err, http.ErrNoCookie) && joinQueue.Action == "leave" {
		http.Error(w, "Must provide playerid to leave", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Could not read cookie", http.StatusInternalServerError)
	}

	fmt.Printf("Player ID: %v\n", playerid)
	var playerIDasInt int64
	playerIDasInt, err = strconv.ParseInt(playerid, 10, 64)
	if err != nil {
		http.Error(w, "Bad PlayerID", http.StatusInternalServerError)
		return
	}

	if joinQueue.Action == "join" {
		// err = addPlayerToQueue(playerIDasInt, joinQueue.Time, joinQueue.Increment)
		addPlayerToWaitingPool(playerIDasInt)
	} else {
		// err = removePlayerFromQueue(playerIDasInt, joinQueue.Time, joinQueue.Increment)
		removePlayerFromWaitingPool(playerIDasInt)
	}

	elapsed := time.Since(start)
	fmt.Printf("Took %s\n\n", elapsed)
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
		http.Error(w, "Could not read cookies", http.StatusBadRequest)
	}

	var playerIDasInt int64

	playerIDasInt, err = strconv.ParseInt(playerid, 10, 64)
	if err != nil {
		http.Error(w, "Bad PlayerID", http.StatusInternalServerError)
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

	defer checkLiveMatches()

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

type MatchStateResponse struct {
	PlayerCode playerCodeEnum `json:"playerCode"`
	MatchState MatchStateData `json:"matchStateData"`
}

func getMatchStateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	var couldBePlayer = true
	var err error
	var playerid string
	var matchID int64

	playerid, err = ReadSigned(r, app.secretKey, "playerid")
	if err != nil {
		couldBePlayer = false
	}

	var playerIDasInt int64
	if couldBePlayer {
		playerIDasInt, err = strconv.ParseInt(playerid, 10, 64)
		if err != nil {
			couldBePlayer = false
		}
	}

	matchID, err = strconv.ParseInt(r.PathValue("matchID"), 10, 64)
	if err != nil {
		http.Error(w, "Could not parse match room", http.StatusInternalServerError)
		return
	}

	var matchStateData *MatchStateData
	matchStateData, err = getLiveMatchStateFromInt64(matchID)
	if err != nil {
		http.Error(w, "Could not get match data", http.StatusInternalServerError)
	}

	var playerCode playerCodeEnum

	if !couldBePlayer {
		playerCode = Spectator
	} else if playerIDasInt == matchStateData.white_player_id {
		playerCode = WhitePieces
	} else {
		playerCode = BlackPieces
	}

	response := MatchStateResponse{PlayerCode: playerCode, MatchState: *matchStateData}

	var jsonStr []byte

	jsonStr, err = json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonStr)
}
