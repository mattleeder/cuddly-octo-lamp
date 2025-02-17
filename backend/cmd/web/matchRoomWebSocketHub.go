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

	turn byte // byte(0) is white, byte(1) is black

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

	var turn byte
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

	// Calculate time remaining for player whose turn it is
	splitFEN = strings.Split(matchState.CurrentFEN, " ")
	if splitFEN[1] == "w" {
		turn = byte(0)
		whitePlayerTimeRemaining = time.Duration(matchState.WhitePlayerTimeRemainingMilliseconds)*time.Millisecond - time.Since(timeOfLastMove)
		blackPlayerTimeRemaining = time.Duration(matchState.BlackPlayerTimeRemainingMilliseconds) * time.Millisecond
		flagTimer = time.After(whitePlayerTimeRemaining)
	} else {
		turn = byte(1)
		whitePlayerTimeRemaining = time.Duration(matchState.WhitePlayerTimeRemainingMilliseconds) * time.Millisecond
		blackPlayerTimeRemaining = time.Duration(matchState.BlackPlayerTimeRemainingMilliseconds)*time.Millisecond - time.Since(timeOfLastMove)
		flagTimer = time.After(blackPlayerTimeRemaining)
	}

	var isTimerActive = false
	if splitFEN[5] == "1" {
		isTimerActive = true
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
	if hub.turn == byte(0) {
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

func (hub *MatchRoomHub) updateMatchThenMoveToFinished(newFEN string, piece int, move int, gameOverStatus gameOverStatusCode, turn byte, matchStateHistoryData []byte) {
	app.liveMatches.UpdateLiveMatch(hub.matchID, newFEN, piece, move, hub.whitePlayerTimeRemaining.Milliseconds(), hub.blackPlayerTimeRemaining.Milliseconds(), matchStateHistoryData, hub.timeOfLastMove)
	var outcome int
	if gameOverStatus == Checkmate || gameOverStatus == WhiteFlagged || gameOverStatus == BlackFlagged {
		if turn == byte(1) {
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
	if message[0] == byte(0) && hub.isTimerActive {
		hub.whitePlayerTimeRemaining -= time.Since(hub.timeOfLastMove)
		hub.whitePlayerTimeRemaining += hub.increment
	} else if message[1] == byte(1) && hub.isTimerActive {
		hub.blackPlayerTimeRemaining -= time.Since(hub.timeOfLastMove)
		hub.blackPlayerTimeRemaining += hub.increment
	} else if message[1] == byte(1) {
		hub.isTimerActive = true
	}

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

	// Start new flag timer and update turn
	if message[0] == byte(0) && hub.isTimerActive {
		hub.flagTimer = time.After(hub.blackPlayerTimeRemaining)
		hub.turn = byte(1)
	} else if message[0] == byte(1) && hub.isTimerActive {
		hub.flagTimer = time.After(hub.whitePlayerTimeRemaining)
		hub.turn = byte(0)
	}

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
	if hub.turn == byte(0) && hub.isTimerActive {
		gameState[0].MatchStateHistory[len(gameState[0].MatchStateHistory)-1].WhitePlayerTimeRemainingMilliseconds -= time.Since(hub.timeOfLastMove).Milliseconds()
	} else if hub.turn == byte(1) && hub.isTimerActive {
		gameState[0].MatchStateHistory[len(gameState[0].MatchStateHistory)-1].BlackPlayerTimeRemainingMilliseconds -= time.Since(hub.timeOfLastMove).Milliseconds()
	}

	jsonStr, err = json.Marshal(gameState)
	if err != nil {
		app.errorLog.Printf("Error marshalling JSON: %v\n", err)
		return []byte{}, err
	}

	return jsonStr, nil
}

func (hub *MatchRoomHub) run() {
	defer app.infoLog.Println("Hub stopped")
	for {
		app.infoLog.Println("Hub running")
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

		case <-hub.flagTimer:
			err := hub.updateGameStateAfterFlag()
			if err != nil {
				app.errorLog.Println(err)
				continue
			}

			hub.sendMessageToAllClients(hub.currentGameState)

		case message := <-hub.broadcast:
			app.infoLog.Printf("WS Message: %v\n", message)

			// Ignore messages from inactive player
			if message[0] != hub.turn {
				continue
			}

			// Validate move and update
			err := hub.updateGameStateAfterMove(message)
			if err != nil {
				app.errorLog.Println(err)
				continue
			}

			hub.sendMessageToAllClients(hub.currentGameState)

		}
	}
}
