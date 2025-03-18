package main

import (
	"burrchess/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrValueTooLong = errors.New("cookie value too long")
	ErrInvalidValue = errors.New("invalid cookie value")
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

type getHighestEloMatchResponse struct {
	MatchID int64 `json:"matchID"`
}

type authData struct {
	Username string `json:"username"`
}

func generateNewPlayerId() int64 {
	return rand.Int63()
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000")
	app.clientError(w, http.StatusNotFound)
}

func getChessMovesHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() { app.perfLog.Printf("getChessMovesHandler took: %s\n", time.Since(start)) }()

	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

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

	var joinQueue joinQueueRequest

	app.infoLog.Printf("%v\n", r.Body)

	err := json.NewDecoder(r.Body).Decode(&joinQueue)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.infoLog.Printf("Received body: %+v\n", joinQueue)

	// Generate new playerID if it doesnt exist, this is for logged out players
	if !app.sessionManager.Exists(r.Context(), "playerID") && joinQueue.Action == "join" {
		var playerID = generateNewPlayerId()
		app.sessionManager.Put(r.Context(), "playerID", playerID)
	} else if !app.sessionManager.Exists(r.Context(), "playerID") && joinQueue.Action == "leave" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var playerID = app.sessionManager.GetInt64(r.Context(), "playerID")

	app.infoLog.Printf("Player ID: %v\n", playerID)

	if joinQueue.Action == "join" {
		addPlayerToWaitingPool(playerID, joinQueue.TimeFormatInMilliseconds, joinQueue.IncrementInMilliseconds)
	} else {
		// err = removePlayerFromQueue(playerIDasInt, joinQueue.Time, joinQueue.Increment)
		removePlayerFromWaitingPool(playerID, joinQueue.TimeFormatInMilliseconds, joinQueue.IncrementInMilliseconds)
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

	if !app.sessionManager.Exists(r.Context(), "playerID") {
		app.serverError(w, errors.New("no playerID in session"))
	}

	var playerID = app.sessionManager.GetInt64(r.Context(), "playerID")
	app.infoLog.Printf("playerID in session: %v", playerID)

	// Do the channels get properly closed on a leave queue?
	// Do we send the proper message on a leave queue?
	clients.mu.Lock()
	_, ok := clients.clients[playerID]
	if !ok {
		clients.clients[playerID] = &Client{id: playerID, channel: make(chan string)}
	}
	clientChannel := clients.clients[playerID].channel
	clients.mu.Unlock()

	// Set appropriate headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	defer func() {
		clients.mu.Lock()
		delete(clients.clients, playerID)
		clients.mu.Unlock()
	}()

	defer app.liveMatches.EnQueueLogAll()

	for {
		select {
		case message, ok := <-clientChannel:
			if !ok {
				return
			}
			app.infoLog.Printf("Sending: data: %s\n\n", message)
			// Send the message to the client in SSE format
			fmt.Fprintf(w, "data: %s\n\n", message)
			// Flush the response to send the data to the client
			w.(http.Flusher).Flush()
		}
	}
}

func getHighestEloMatchHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() { app.perfLog.Printf("getChessMovesHandler took: %s\n", time.Since(start)) }()

	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	matchID, err := app.liveMatches.GetHighestEloMatch()
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := getHighestEloMatchResponse{MatchID: matchID}
	jsonStr, err := json.Marshal(data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Write(jsonStr)
}

func registerUserHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() { app.perfLog.Printf("registerUserHandler took: %s\n", time.Since(start)) }()

	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	var newUser models.NewUserInfo

	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		app.serverError(w, err)
		return
	}

	newUserOptions := models.CreateNewUserOptions(newUser)

	var registerUserValidationErrors models.NewUserInfo

	err = app.users.InsertNew(newUser.Username, newUser.Password, &newUserOptions)
	if err != nil {
		app.errorLog.Printf("DB Error: %s\n", err.Error())
		if err.Error() == "constraint failed: UNIQUE constraint failed: users.username (2067)" {
			registerUserValidationErrors.Username = "Username already exists."
		}
		jsonStr, jsonErr := json.Marshal(registerUserValidationErrors)
		if jsonErr != nil {
			app.errorLog.Printf("Error marshalling json: %s\n", jsonErr.Error())
			app.serverError(w, err)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonStr)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() { app.perfLog.Printf("loginHandler took: %s\n", time.Since(start)) }()

	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	var loginInfo models.UserLoginInfo

	err := json.NewDecoder(r.Body).Decode(&loginInfo)
	if err != nil {
		app.serverError(w, err)
		return
	}

	playerID, authorized := app.users.Authenticate(loginInfo.Username, loginInfo.Password)
	if !authorized {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var authData = authData{
		Username: loginInfo.Username,
	}

	jsonStr, err := json.Marshal(authData)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "username", loginInfo.Username)
	app.sessionManager.Put(r.Context(), "playerID", playerID)
	w.Write(jsonStr)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() { app.perfLog.Printf("logoutHandler took: %s\n", time.Since(start)) }()

	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	if !app.sessionManager.Exists(r.Context(), "username") {
		app.errorLog.Printf("Not logged in\n")
		w.WriteHeader(http.StatusBadRequest)
	}

	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Destroy(r.Context())
	w.WriteHeader(http.StatusOK)
}

func validateSessionHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() { app.perfLog.Printf("validateSessionHandler took: %s\n", time.Since(start)) }()

	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	if !app.sessionManager.Exists(r.Context(), "username") {
		if !app.sessionManager.Exists(r.Context(), "playerID") {
			app.sessionManager.Put(r.Context(), "playerID", generateNewPlayerId())
		}
		w.WriteHeader(http.StatusUnauthorized)
	}

	var authData = authData{
		Username: app.sessionManager.GetString(r.Context(), "username"),
	}

	jsonStr, err := json.Marshal(authData)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Write(jsonStr)
}
