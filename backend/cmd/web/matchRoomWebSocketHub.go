package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Hub Functions
// Create New
// Run
// GetMessageType
// HandleMessage
//

type MatchStateHistory struct {
	FEN                                  string `json:"FEN"`
	LastMove                             [2]int `json:"lastMove"`
	AlgebraicNotation                    string `json:"algebraicNotation"`
	WhitePlayerTimeRemainingMilliseconds int64  `json:"whitePlayerTimeRemainingMilliseconds"`
	BlackPlayerTimeRemainingMilliseconds int64  `json:"blackPlayerTimeRemainingMilliseconds"`
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type MatchRoomHub struct {
	matchID int64
	// Registered clients.
	clients map[*MatchRoomHubClient]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *MatchRoomHubClient

	// Unregister requests from clients.
	unregister chan *MatchRoomHubClient

	whitePlayerID int64

	blackPlayerID int64

	whitePlayerTimeRemaining time.Duration

	blackPlayerTimeRemaining time.Duration

	isTimerActive bool

	turn playerTurn // byte(0) is white, byte(1) is black

	currentGameState []byte

	current_fen string

	moveHistory []MatchStateHistory

	timeOfLastMove time.Time

	flagTimer <-chan time.Time

	timeFormatInMilliseconds int64

	increment time.Duration

	fenFreqMap map[string]int
}

type wsChessMove struct {
	Piece           int    `json:"piece"`
	Move            int    `json:"move"`
	PromotionString string `json:"promotionString"`
}

type msgType int

const (
	PlayerMove = iota
	PlayerMessage
	DrawOffer
	DrawClaim
	SpectatorMessage
	Unknown
)

type playerTurn byte

const (
	WhiteTurn = byte(iota)
	BlackTurn = byte(iota)
)

func newMatchRoomHub(matchID int64) (*MatchRoomHub, error) {
	// Build hub from data in db
	matchState, err := app.liveMatches.GetFromMatchID(matchID)

	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	var matchStateHistory []MatchStateHistory

	err = json.Unmarshal(matchState.GameHistoryJSONString, &matchStateHistory)
	if err != nil {
		app.errorLog.Printf("Error unmarshalling matchStateHistory %v\n", err)
		return nil, err
	}

	var turn playerTurn
	var fenFreqMap = make(map[string]int)
	var splitFEN []string
	var fen string
	var threefoldRepetition = false

	// FEN Freq Map
	for _, val := range matchStateHistory {
		splitFEN = strings.Split(val.FEN, " ")
		fen = strings.Join(splitFEN[:4], " ")
		fenFreqMap[fen] += 1
		if fenFreqMap[fen] >= 3 {
			threefoldRepetition = true
		}
	}

	currentGameState := [1]postChessMoveReply{{
		MatchStateHistory:   matchStateHistory,
		GameOverStatus:      Ongoing,
		ThreefoldRepetition: threefoldRepetition,
	}}

	timeOfLastMove := time.UnixMilli(matchState.UnixMsTimeOfLastMove)
	var whitePlayerTimeRemaining, blackPlayerTimeRemaining time.Duration
	var flagTimer <-chan time.Time

	splitFEN = strings.Split(matchState.CurrentFEN, " ")

	// Calculate time remaining for player whose turn it is
	var isTimerActive = false
	if splitFEN[5] != "1" {
		isTimerActive = true
	}

	whitePlayerTimeRemaining = time.Duration(matchState.WhitePlayerTimeRemainingMilliseconds) * time.Millisecond
	blackPlayerTimeRemaining = time.Duration(matchState.BlackPlayerTimeRemainingMilliseconds) * time.Millisecond

	if splitFEN[1] == "w" {
		turn = playerTurn(WhiteTurn)
	} else {
		turn = playerTurn(BlackTurn)
	}

	if turn == playerTurn(WhiteTurn) && isTimerActive {
		whitePlayerTimeRemaining = time.Duration(matchState.WhitePlayerTimeRemainingMilliseconds)*time.Millisecond - time.Since(timeOfLastMove)
		flagTimer = time.After(whitePlayerTimeRemaining)
	} else if turn == playerTurn(BlackTurn) && isTimerActive {
		blackPlayerTimeRemaining = time.Duration(matchState.BlackPlayerTimeRemainingMilliseconds)*time.Millisecond - time.Since(timeOfLastMove)
		flagTimer = time.After(blackPlayerTimeRemaining)
	}

	jsonStr, err := json.Marshal(currentGameState)
	if err != nil {
		app.errorLog.Printf("Error marshalling JSON: %v\n", err)
		return nil, err
	}

	match := &MatchRoomHub{
		matchID:                  matchID,
		broadcast:                make(chan []byte),
		register:                 make(chan *MatchRoomHubClient),
		unregister:               make(chan *MatchRoomHubClient),
		clients:                  make(map[*MatchRoomHubClient]bool),
		whitePlayerID:            matchState.WhitePlayerID,
		blackPlayerID:            matchState.BlackPlayerID,
		whitePlayerTimeRemaining: whitePlayerTimeRemaining,
		blackPlayerTimeRemaining: blackPlayerTimeRemaining,
		isTimerActive:            isTimerActive,
		turn:                     turn,
		currentGameState:         jsonStr,
		current_fen:              matchState.CurrentFEN,
		moveHistory:              currentGameState[0].MatchStateHistory,
		timeOfLastMove:           timeOfLastMove,
		flagTimer:                flagTimer,
		timeFormatInMilliseconds: matchState.TimeFormatInMilliseconds,
		increment:                time.Duration(matchState.IncrementInMilliseconds) * time.Millisecond,
		fenFreqMap:               fenFreqMap,
	}

	return match, nil
}

func (hub *MatchRoomHub) sendMessageToAllClients(message []byte) {
	for client := range hub.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(hub.clients, client)
		}
	}
}

func (hub *MatchRoomHub) updateGameStateAfterFlag() (err error) {
	var gameState []postChessMoveReply
	err = json.Unmarshal(hub.currentGameState, &gameState)
	if err != nil {
		app.errorLog.Printf("Error unmarshalling JSON: %v\n", err)
		return err
	}

	var outcome int
	if hub.turn == playerTurn(WhiteTurn) {
		gameState[0].GameOverStatus = WhiteFlagged
		outcome = 2
	} else {
		gameState[0].GameOverStatus = BlackFlagged
		outcome = 1
	}

	var jsonStr []byte

	jsonStr, err = json.Marshal(gameState)
	if err != nil {
		app.errorLog.Printf("Error marshalling JSON: %v\n", err)
		return err
	}

	hub.currentGameState = jsonStr

	// Game is over after flag
	go app.liveMatches.MoveMatchToPastMatches(hub.matchID, outcome)

	return nil
}

func (hub *MatchRoomHub) updateMatchThenMoveToFinished(newFEN string, piece int, move int, gameOverStatus gameOverStatusCode, turn playerTurn, matchStateHistoryData []byte) {
	app.liveMatches.UpdateLiveMatch(hub.matchID, newFEN, piece, move, hub.whitePlayerTimeRemaining.Milliseconds(), hub.blackPlayerTimeRemaining.Milliseconds(), matchStateHistoryData, hub.timeOfLastMove)
	var outcome int
	if gameOverStatus == Checkmate || gameOverStatus == WhiteFlagged || gameOverStatus == BlackFlagged {
		if turn == playerTurn(BlackTurn) {
			// White wins as its blacks turn
			outcome = 1
		} else {
			// Black wins
			outcome = 2
		}
	} else {
		outcome = 0
	}
	app.liveMatches.MoveMatchToPastMatches(hub.matchID, outcome)
	app.pastMatches.LogAll()
}

func (hub *MatchRoomHub) changeTurn() {
	// Swap turn and activate timer if changing back to whites turn
	if hub.turn == playerTurn(WhiteTurn) {
		hub.turn = playerTurn(BlackTurn)
	} else {
		hub.turn = playerTurn(WhiteTurn)
	}

	if hub.isTimerActive && hub.turn == playerTurn(BlackTurn) {
		hub.flagTimer = time.After(hub.blackPlayerTimeRemaining)
	} else if hub.isTimerActive {
		hub.flagTimer = time.After(hub.whitePlayerTimeRemaining)
	} else if hub.turn == playerTurn(WhiteTurn) {
		hub.isTimerActive = true
	}
}

func (hub *MatchRoomHub) updateTimeRemaining() {
	if !hub.isTimerActive {
		return
	}

	if hub.turn == playerTurn(WhiteTurn) {
		hub.whitePlayerTimeRemaining -= time.Since(hub.timeOfLastMove)
		hub.whitePlayerTimeRemaining += hub.increment
	} else if byte(hub.turn) == BlackTurn {
		hub.blackPlayerTimeRemaining -= time.Since(hub.timeOfLastMove)
		hub.blackPlayerTimeRemaining += hub.increment
	}
}

func (hub *MatchRoomHub) updateGameStateAfterMove(message []byte) (err error) {

	// Parse Message
	var chessMove wsChessMove
	err = json.Unmarshal(message[1:], &chessMove)
	if err != nil {
		return errors.New(fmt.Sprintf("Error unmarshalling JSON: %v\n", err))
	}

	// Validate Move
	var validMove = IsMoveValid(hub.current_fen, chessMove.Piece, chessMove.Move)
	if !validMove {
		return errors.New("Move is not valid")
	}

	// Calcuate new time remaining
	hub.updateTimeRemaining()

	// Calculate reply variables
	newFEN, gameOverStatus, algebraicNotation := getFENAfterMove(hub.current_fen, chessMove.Piece, chessMove.Move, chessMove.PromotionString)
	var threefoldRepetition = false
	splitFEN := strings.Join(strings.Split(newFEN, " ")[:4], " ")
	hub.fenFreqMap[splitFEN] += 1
	if hub.fenFreqMap[splitFEN] >= 3 {
		threefoldRepetition = true
	}

	// Construct Reply
	data := []postChessMoveReply{
		{
			MatchStateHistory: append(hub.moveHistory, MatchStateHistory{
				FEN:                                  newFEN,
				LastMove:                             [2]int{chessMove.Piece, chessMove.Move},
				AlgebraicNotation:                    algebraicNotation,
				WhitePlayerTimeRemainingMilliseconds: hub.whitePlayerTimeRemaining.Milliseconds(),
				BlackPlayerTimeRemainingMilliseconds: hub.blackPlayerTimeRemaining.Milliseconds(),
			}),
			GameOverStatus:      gameOverStatus,
			ThreefoldRepetition: threefoldRepetition,
		},
	}

	var jsonStr []byte

	jsonStr, err = json.Marshal(data)
	if err != nil {
		return errors.New(fmt.Sprintf("Error marshalling JSON: %v\n", err))
	}

	// Update game state
	hub.current_fen = newFEN
	hub.currentGameState = jsonStr
	hub.moveHistory = data[0].MatchStateHistory
	hub.timeOfLastMove = time.Now()

	// Ppdate turn and start new flag timer
	hub.changeTurn()

	var matchStateHistoryData []byte
	matchStateHistoryData, err = json.Marshal(data[0].MatchStateHistory)
	if err != nil {
		app.errorLog.Printf("Error marshalling matchStateHistoryData: %s", err)
	}

	// Update database
	if gameOverStatus != Ongoing {
		go hub.updateMatchThenMoveToFinished(newFEN, chessMove.Piece, chessMove.Move, gameOverStatus, hub.turn, matchStateHistoryData)
	} else {
		go app.liveMatches.UpdateLiveMatch(hub.matchID, newFEN, chessMove.Piece, chessMove.Move, hub.whitePlayerTimeRemaining.Milliseconds(), hub.blackPlayerTimeRemaining.Milliseconds(), matchStateHistoryData, hub.timeOfLastMove)
	}

	return nil
}

func (hub *MatchRoomHub) getCurrentMatchStateForNewConnection() (jsonStr []byte, err error) {
	var gameState []postChessMoveReply
	err = json.Unmarshal(hub.currentGameState, &gameState)
	if err != nil {
		app.errorLog.Printf("Error unmarshalling JSON: %v\n", err)
		return []byte{}, err
	}

	// Correct times
	if hub.turn == playerTurn(WhiteTurn) && hub.isTimerActive {
		gameState[0].MatchStateHistory[len(gameState[0].MatchStateHistory)-1].WhitePlayerTimeRemainingMilliseconds -= time.Since(hub.timeOfLastMove).Milliseconds()
	} else if hub.turn == playerTurn(BlackTurn) && hub.isTimerActive {
		gameState[0].MatchStateHistory[len(gameState[0].MatchStateHistory)-1].BlackPlayerTimeRemainingMilliseconds -= time.Since(hub.timeOfLastMove).Milliseconds()
	}

	jsonStr, err = json.Marshal(gameState)
	if err != nil {
		app.errorLog.Printf("Error marshalling JSON: %v\n", err)
		return []byte{}, err
	}

	return jsonStr, nil
}

func (hub *MatchRoomHub) hasActiveClients() bool {
	for _, val := range hub.clients {
		if val {
			return true
		}
	}
	return false
}

func (hub *MatchRoomHub) getMessageType(message []byte) msgType {
	if message[0] == WhiteTurn || message[0] == BlackTurn {
		return PlayerMove
	}
	app.errorLog.Printf("Unknown message type\n")
	return Unknown
}

func (hub *MatchRoomHub) handleMessage(message []byte) (response []byte) {
	switch msgType := hub.getMessageType(message); msgType {
	case PlayerMove:

		// Ignore messages from inactive player
		if message[0] != byte(hub.turn) {
			app.infoLog.Printf("Not your turn\n")
			return nil
		}

		// Validate move and update
		err := hub.updateGameStateAfterMove(message)
		if err != nil {
			app.errorLog.Println(err)
			return nil
		}

		return hub.currentGameState

	default:
		app.errorLog.Printf("Could not understand message: %s\n", message)
		return nil
	}
}

func (hub *MatchRoomHub) run() {
	app.infoLog.Println("Hub running")
	defer app.infoLog.Println("Hub stopped")
	for {
		select {
		// Clients get currentGameState on register
		case client := <-hub.register:
			hub.clients[client] = true
			jsonStr, err := hub.getCurrentMatchStateForNewConnection()
			if err != nil {
				app.errorLog.Printf("Could not get json for new connection: %v\n", err)
				continue
			}
			client.send <- jsonStr

		case client := <-hub.unregister:
			if _, ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				close(client.send)
			}
			if !hub.hasActiveClients() {
				matchRoomHubManager.unregisterHub(hub.matchID)
				return
			}

		case <-hub.flagTimer:
			err := hub.updateGameStateAfterFlag()
			if err != nil {
				app.errorLog.Println(err)
				continue
			}

			hub.sendMessageToAllClients(hub.currentGameState)

		case message := <-hub.broadcast:
			app.infoLog.Printf("WS Message: %s\n", message)
			app.infoLog.Printf("WS Message: %v\n", message)

			response := hub.handleMessage(message)
			app.infoLog.Printf("Response: %s\n", response)
			if response != nil {
				hub.sendMessageToAllClients(response)
			}

		}
	}
}
