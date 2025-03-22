package main

import (
	"burrchess/internal/chess"
	"burrchess/internal/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// WEBSOCKET TO CLIENT TYPES

type hubMessageType string

const (
	onConnect        = "onConnect"
	onMove           = "onMove"
	connectionStatus = "connectionStatus"
	opponentEvent    = "opponentEvent"
	userMessage      = "userMessage"
	sendPlayerCode   = "sendPlayerCode"
)

type eventType string

const (
	takeback            = "takeback"
	draw                = "draw"
	resign              = "resign"
	extraTime           = "extraTime"
	abort               = "abort"
	rematch             = "rematch"
	disconnect          = "disconnect"
	decline             = "decline"
	threefoldRepetition = "threefoldRepetition"
)

// Bodies

type onConnectBody struct {
	MatchStateHistory        []MatchStateHistory      `json:"matchStateHistory"`
	GameOverStatusCode       chess.GameOverStatusCode `json:"gameOverStatus"`
	ThreefoldRepetition      bool                     `json:"threefoldRepetition"`
	WhitePlayerConnected     bool                     `json:"whitePlayerConnected"`
	BlackPlayerConnected     bool                     `json:"blackPlayerConnected"`
	MillisecondsUntilTimeout int64                    `json:"millisecondsUntilTimeout"`
	WhitePlayerUsername      sql.NullString           `json:"whitePlayerUsername"`
	BlackPlayerUsername      sql.NullString           `json:"blackPlayerUsername"`
}

type onMoveBody struct {
	MatchStateHistory   []MatchStateHistory      `json:"matchStateHistory"`
	GameOverStatusCode  chess.GameOverStatusCode `json:"gameOverStatus"`
	ThreefoldRepetition bool                     `json:"threefoldRepetition"`
}

type onPlayerConnectionChangeBody struct {
	PlayerColour             string `json:"playerColour"`
	IsConnected              bool   `json:"isConnected"`
	MillisecondsUntilTimeout int64  `json:"millisecondsUntilTimeout"`
}

type opponentEventBody struct {
	Sender    string    `json:"sender"`
	EventType eventType `json:"eventType"`
}

type onUserMessageBody struct {
	Sender         string `json:"sender"`
	MessageContent string `json:"messageContent"`
}

// Responses

type onConnectResponse struct {
	MessageType hubMessageType `json:"messageType"`
	Body        onConnectBody  `json:"body"`
}

type onMoveResponse struct {
	MessageType hubMessageType `json:"messageType"`
	Body        onMoveBody     `json:"body"`
}

type onPlayerConnectionChangeResponse struct {
	MessageType hubMessageType               `json:"messageType"`
	Body        onPlayerConnectionChangeBody `json:"body"`
}

type opponentEventResponse struct {
	MessageType hubMessageType    `json:"messageType"`
	Body        opponentEventBody `json:"body"`
}

type onUserMessageResponse struct {
	MessageType hubMessageType    `json:"messageType"`
	Body        onUserMessageBody `json:"body"`
}

// CLIENT TO WEBSOCKET TYPES

type clientMessageType string

const (
	postMove    = "postMove"
	playerEvent = "playerEvent"
	chatMessage = "userMessage"
	unknown     = "unknown"
)

// Bodies
type postMoveBody struct {
	Piece           int    `json:"piece"`
	Move            int    `json:"move"`
	PromotionString string `json:"promotionString"`
}

type playerEventBody struct {
	EventType eventType `json:"eventType"`
}

type userMessageBody struct {
	MessageContent string `json:"messageContent"`
}

// Responses
type postMoveResponse struct {
	MessageType clientMessageType `json:"messageType"`
	Body        postMoveBody      `json:"body"`
}

type playerEventResponse struct {
	MessageType clientMessageType `json:"messageType"`
	Body        playerEventBody   `json:"body"`
}

type userMessageResponse struct {
	MessageType clientMessageType `json:"messageType"`
	Body        userMessageBody   `json:"body"`
}

//////////////////////////////////////////////////////////////

type userJSON struct {
	MessageType string          `json:"messageType"`
	Body        json.RawMessage `json:"body"`
}

type MatchStateHistory struct {
	FEN                                  string `json:"FEN"`
	LastMove                             [2]int `json:"lastMove"`
	AlgebraicNotation                    string `json:"algebraicNotation"`
	WhitePlayerTimeRemainingMilliseconds int64  `json:"whitePlayerTimeRemainingMilliseconds"`
	BlackPlayerTimeRemainingMilliseconds int64  `json:"blackPlayerTimeRemainingMilliseconds"`
}

type offerInfo struct {
	sender messageIdentifier
	event  eventType
}

const pingTimeout = 20 * time.Second

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

	whitePlayerUsername sql.NullString

	blackPlayerUsername sql.NullString

	blackPlayerID int64

	whitePlayerTimeRemaining time.Duration

	blackPlayerTimeRemaining time.Duration

	isTimerActive bool

	turn playerTurn // byte(0) is white, byte(1) is black

	currentGameState []byte // onMoveResponse

	current_fen string

	moveHistory []MatchStateHistory

	timeOfLastMove time.Time

	flagTimer <-chan time.Time

	timeFormatInMilliseconds int64

	increment time.Duration

	fenFreqMap map[string]int

	whitePlayerConnected bool

	blackPlayerConnected bool

	whitePlayerTimeout <-chan time.Time

	whitePlayerTimeoutStarted time.Time

	blackPlayerTimeout <-chan time.Time

	blackPlayerTimeoutStarted time.Time

	whiteCanClaimTimeout bool

	blackCanClaimTimeout bool

	offerActive *offerInfo

	gameEnded bool

	threefoldRepetition bool

	averageElo float64

	whitePlayerElo int64

	blackPlayerElo int64

	matchStartTime int64

	taskQueueWaitGroup *sync.WaitGroup
}

type playerTurn byte

const (
	WhiteTurn = byte(iota)
	BlackTurn = byte(iota)
)

func newMatchRoomHub(matchID int64) (*MatchRoomHub, error) {
	// Build hub from data in db
	matchState, err := app.liveMatches.EnQueueReturnGetFromMatchID(matchID, nil, nil)

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

	currentGameState := onMoveResponse{
		MessageType: onMove,
		Body: onMoveBody{
			MatchStateHistory:   matchStateHistory,
			GameOverStatusCode:  chess.Ongoing,
			ThreefoldRepetition: threefoldRepetition,
		},
	}

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
		whitePlayerUsername:      matchState.WhitePlayerUsername,
		blackPlayerUsername:      matchState.BlackPlayerUsername,
		blackPlayerID:            matchState.BlackPlayerID,
		whitePlayerTimeRemaining: whitePlayerTimeRemaining,
		blackPlayerTimeRemaining: blackPlayerTimeRemaining,
		isTimerActive:            isTimerActive,
		turn:                     turn,
		currentGameState:         jsonStr,
		current_fen:              matchState.CurrentFEN,
		moveHistory:              currentGameState.Body.MatchStateHistory,
		timeOfLastMove:           timeOfLastMove,
		flagTimer:                flagTimer,
		timeFormatInMilliseconds: matchState.TimeFormatInMilliseconds,
		increment:                time.Duration(matchState.IncrementInMilliseconds) * time.Millisecond,
		fenFreqMap:               fenFreqMap,
		whitePlayerConnected:     false,
		blackPlayerConnected:     false,
		threefoldRepetition:      threefoldRepetition,
		averageElo:               matchState.AverageElo,
		whitePlayerElo:           matchState.WhitePlayerElo,
		blackPlayerElo:           matchState.BlackPlayerElo,
		matchStartTime:           matchState.MatchStartTime,
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

func (hub *MatchRoomHub) sendMessageToAllPlayers(message []byte) {
	for client := range hub.clients {
		if client.playerIdentifier != messageIdentifier(WhitePlayer) && client.playerIdentifier != messageIdentifier(BlackPlayer) {
			continue
		}
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(hub.clients, client)
		}
	}
}

func (hub *MatchRoomHub) sendMessageToOnePlayer(message []byte, colour messageIdentifier) {
	for client := range hub.clients {
		if client.playerIdentifier != colour {
			continue
		}
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(hub.clients, client)
		}
	}
}

func (hub *MatchRoomHub) sendMessageToAllSpectators(message []byte) {
	for client := range hub.clients {
		if client.playerIdentifier == messageIdentifier(WhitePlayer) || client.playerIdentifier == messageIdentifier(BlackPlayer) {
			continue
		}
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(hub.clients, client)
		}
	}
}

func getKFactor(elo int64) float64 {
	if elo < 2100 {
		return 32
	} else if elo <= 2400 {
		return 24
	} else {
		return 16
	}
}

func calculateEloChanges(playerOneElo int64, playerOnePoints float64, playerTwoElo int64, playerTwoPoints float64) (playerOneEloGain float64, playerTwoEloGain float64) {
	var playerOneExpectedPoints, playerTwoExpectedPoints float64

	playerOneExpectedPoints = (1) / (1 + math.Pow(10, (playerTwoExpectedPoints-playerOneExpectedPoints)/400))
	playerTwoExpectedPoints = (1) / (1 + math.Pow(10, (playerOneExpectedPoints-playerTwoExpectedPoints)/400))

	playerOneEloGain = getKFactor(playerOneElo) * (playerOnePoints - playerOneExpectedPoints)
	playerTwoEloGain = getKFactor(playerTwoElo) * (playerTwoPoints - playerTwoExpectedPoints)

	return playerOneEloGain, playerTwoEloGain
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
	} else if hub.isTimerActive || hub.turn == playerTurn(WhiteTurn) {
		hub.flagTimer = time.After(hub.whitePlayerTimeRemaining)
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

func (hub *MatchRoomHub) getOutcomeInt(gameOverStatus chess.GameOverStatusCode) int {
	if gameOverStatus == chess.Checkmate {
		if hub.turn == chess.Black {
			return 2
		} else {
			return 1
		}
	} else if gameOverStatus == chess.WhiteFlagged || gameOverStatus == chess.WhiteResigned {
		return 2
	} else if gameOverStatus == chess.BlackFlagged || gameOverStatus == chess.BlackResigned {
		return 1
	}
	return 0
}

func (hub *MatchRoomHub) updateGameStateAfterMove(message []byte) (err error) {

	// Parse Message
	var chessMove postMoveResponse
	err = json.Unmarshal(message[1:], &chessMove)
	if err != nil {
		return errors.New(fmt.Sprintf("error unmarshalling JSON: %v\n", err))
	}

	// Validate Move
	var validMove = chess.IsMoveValid(hub.current_fen, chessMove.Body.Piece, chessMove.Body.Move)
	if !validMove {
		return errors.New("move is not valid")
	}

	// Calcuate new time remaining
	hub.updateTimeRemaining()

	// Calculate reply variables
	newFEN, gameOverStatus, algebraicNotation := chess.GetFENAfterMove(hub.current_fen, chessMove.Body.Piece, chessMove.Body.Move, chessMove.Body.PromotionString)
	var threefoldRepetition = false
	splitFEN := strings.Join(strings.Split(newFEN, " ")[:4], " ")
	hub.fenFreqMap[splitFEN] += 1
	if hub.fenFreqMap[splitFEN] >= 3 {
		threefoldRepetition = true
	}
	hub.threefoldRepetition = threefoldRepetition

	// Construct Reply
	data := onMoveResponse{
		MessageType: onMove,
		Body: onMoveBody{
			MatchStateHistory: append(hub.moveHistory, MatchStateHistory{
				FEN:                                  newFEN,
				LastMove:                             [2]int{chessMove.Body.Piece, chessMove.Body.Move},
				AlgebraicNotation:                    algebraicNotation,
				WhitePlayerTimeRemainingMilliseconds: hub.whitePlayerTimeRemaining.Milliseconds(),
				BlackPlayerTimeRemainingMilliseconds: hub.blackPlayerTimeRemaining.Milliseconds(),
			}),
			GameOverStatusCode:  gameOverStatus,
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
	hub.moveHistory = data.Body.MatchStateHistory
	hub.timeOfLastMove = time.Now()

	// Ppdate turn and start new flag timer
	hub.changeTurn()

	var matchStateHistoryData []byte
	matchStateHistoryData, err = json.Marshal(data.Body.MatchStateHistory)
	if err != nil {
		app.errorLog.Printf("Error marshalling matchStateHistoryData: %s", err)
	}

	// Update database
	// @TODO: do we need a new waitGroup each time? Hub could just have one waitGroup that we add tasks to
	var wg sync.WaitGroup
	wg.Add(1)
	app.liveMatches.EnQueueUpdateLiveMatch(hub.matchID, newFEN, chessMove.Body.Piece, chessMove.Body.Move, hub.whitePlayerTimeRemaining.Milliseconds(), hub.blackPlayerTimeRemaining.Milliseconds(), matchStateHistoryData, hub.timeOfLastMove, hub.taskQueueWaitGroup, &wg)
	hub.taskQueueWaitGroup = &wg

	if gameOverStatus != chess.Ongoing {
		return hub.endGame(gameOverStatus)
	}

	return nil
}

func (hub *MatchRoomHub) getCurrentMatchStateForNewConnection(playerIdentifier messageIdentifier) (jsonStr []byte, err error) {
	var gameState onMoveResponse
	err = json.Unmarshal(hub.currentGameState, &gameState)
	if err != nil {
		app.errorLog.Printf("Error unmarshalling JSON: %v\n", err)
		return []byte{}, err
	}

	// Correct times
	if hub.turn == playerTurn(WhiteTurn) && hub.isTimerActive {
		gameState.Body.MatchStateHistory[len(gameState.Body.MatchStateHistory)-1].WhitePlayerTimeRemainingMilliseconds -= time.Since(hub.timeOfLastMove).Milliseconds()
	} else if hub.turn == playerTurn(BlackTurn) && hub.isTimerActive {
		gameState.Body.MatchStateHistory[len(gameState.Body.MatchStateHistory)-1].BlackPlayerTimeRemainingMilliseconds -= time.Since(hub.timeOfLastMove).Milliseconds()
	}

	var millisecondsUntilTimeout int64 = 0
	if playerIdentifier == chess.White && !hub.blackPlayerConnected {
		millisecondsUntilTimeout = pingTimeout.Milliseconds() - time.Since(hub.blackPlayerTimeoutStarted).Milliseconds()
	} else if playerIdentifier == chess.Black && !hub.whitePlayerConnected {
		millisecondsUntilTimeout = pingTimeout.Milliseconds() - time.Since(hub.whitePlayerTimeoutStarted).Milliseconds()
	}

	var response = onConnectResponse{
		MessageType: onConnect,
		Body: onConnectBody{
			MatchStateHistory:        gameState.Body.MatchStateHistory,
			GameOverStatusCode:       gameState.Body.GameOverStatusCode,
			ThreefoldRepetition:      gameState.Body.ThreefoldRepetition,
			WhitePlayerConnected:     hub.whitePlayerConnected,
			BlackPlayerConnected:     hub.blackPlayerConnected,
			MillisecondsUntilTimeout: millisecondsUntilTimeout,
			WhitePlayerUsername:      hub.whitePlayerUsername,
			BlackPlayerUsername:      hub.blackPlayerUsername,
		},
	}

	jsonStr, err = json.Marshal(response)
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

func (hub *MatchRoomHub) getMessageType(message []byte) clientMessageType {
	var msg userJSON
	err := json.Unmarshal(message[1:], &msg)
	if err != nil {
		app.errorLog.Printf("Unable to parse message into userJSON: %s\n", err)
		return unknown
	}

	switch msg.MessageType {
	case "postMove":
		if message[0] != byte(WhitePlayer) && message[0] != byte(BlackPlayer) {
			app.errorLog.Printf("Non-player trying to post move: %s\n", message)
			return unknown
		}
		return postMove

	case "playerEvent":
		if message[0] != byte(WhitePlayer) && message[0] != byte(BlackPlayer) {
			app.errorLog.Printf("Non-player trying to send playerEvent: %s\n", message)
			return unknown
		}
		return playerEvent
	}

	app.errorLog.Printf("Unknown message type\n")
	return unknown
}

// @TODO: implement this
func (hub *MatchRoomHub) takeBack() {

}

func (hub *MatchRoomHub) acceptEventOffer(event eventType) {
	// @TODO implement
	app.infoLog.Printf("Making accepting event of type %s\n", event)
	switch event {
	case takeback:
		hub.takeBack()

	case draw:
		hub.endGame(chess.Draw)
		hub.sendMessageToAllClients(hub.currentGameState)

	default:
		return
	}
}

func (hub *MatchRoomHub) endGame(reason chess.GameOverStatusCode) error {
	// Updates game state and updates DB. Does not send response
	app.infoLog.Println("Ending Match")
	hub.flagTimer = nil
	var gameState onMoveResponse
	err := json.Unmarshal(hub.currentGameState, &gameState)
	if err != nil {
		app.errorLog.Printf("Error unmarshalling JSON: %v\n", err)
		return err
	}

	gameState.Body.GameOverStatusCode = reason

	var jsonStr []byte

	jsonStr, err = json.Marshal(gameState)
	if err != nil {
		app.errorLog.Printf("Error marshalling JSON: %v\n", err)
		return err
	}

	hub.currentGameState = jsonStr
	var outcome = hub.getOutcomeInt(reason)
	var whitePlayerPoints, blackPlayerPoints float64

	if outcome == 1 {
		whitePlayerPoints = 1
		blackPlayerPoints = 0
	} else if outcome == 2 {
		whitePlayerPoints = 0
		blackPlayerPoints = 1
	} else {
		whitePlayerPoints = 0.5
		blackPlayerPoints = 0.5
	}

	whitePlayerEloGain, blackPlayerEloGain := calculateEloChanges(hub.whitePlayerElo, whitePlayerPoints, hub.blackPlayerElo, blackPlayerPoints)
	app.infoLog.Printf("whitePlayerElo: %v, whitePlayerEloGain: %v\n", hub.whitePlayerElo, whitePlayerEloGain)
	whitePlayerNewElo := int64(math.Max(float64(hub.whitePlayerElo)+math.Round(whitePlayerEloGain), 0))
	blackPlayerNewElo := int64(math.Max(float64(hub.blackPlayerElo)+math.Round(blackPlayerEloGain), 0))
	go app.userRatings.UpdateRatingFromPlayerID(hub.whitePlayerID, models.GetRatingTypeFromTimeFormat(hub.timeFormatInMilliseconds), whitePlayerNewElo)
	go app.userRatings.UpdateRatingFromPlayerID(hub.blackPlayerID, models.GetRatingTypeFromTimeFormat(hub.timeFormatInMilliseconds), blackPlayerNewElo)

	hub.gameEnded = true
	app.liveMatches.EnQueueMoveMatchToPastMatches(hub.matchID, outcome, reason, whitePlayerEloGain, blackPlayerEloGain, hub.taskQueueWaitGroup, nil)
	return nil
}

func (hub *MatchRoomHub) makeNewEventOffer(sender messageIdentifier, event eventType) {
	app.infoLog.Printf("Making new event from %v of type %s\n", sender, event)
	hub.offerActive = &offerInfo{sender, event}
	var responseFrom string
	var receiver messageIdentifier
	if sender == messageIdentifier(WhitePlayer) {
		responseFrom = "white"
		receiver = messageIdentifier(BlackPlayer)
	} else {
		responseFrom = "black"
		receiver = messageIdentifier(WhitePlayer)
	}

	response := opponentEventResponse{
		MessageType: opponentEvent,
		Body:        opponentEventBody{Sender: responseFrom, EventType: event},
	}

	jsonStr, err := json.Marshal(response)
	if err != nil {
		app.errorLog.Printf("Could not marshal opponentEventResponse: %s\n", err)
		return
	}

	hub.sendMessageToOnePlayer(jsonStr, receiver)
}

func isOneSidedEvent(event eventType) bool {
	return event == extraTime || event == resign || event == abort || event == disconnect
}

func (hub *MatchRoomHub) oneSidedEvent(sender messageIdentifier, event eventType) {
	switch event {
	case extraTime:
		return
	case resign:
		if sender == chess.White {
			hub.endGame(chess.WhiteResigned)
		} else {
			hub.endGame(chess.BlackResigned)
		}
		hub.sendMessageToAllClients(hub.currentGameState)
	case disconnect:
		if sender == chess.White && hub.whiteCanClaimTimeout {
			hub.endGame(chess.BlackDisconnected)
		} else if sender == chess.Black && hub.blackCanClaimTimeout {
			hub.endGame(chess.WhiteDisconnected)
		}
		hub.sendMessageToAllClients(hub.currentGameState)
	}
}

func (hub *MatchRoomHub) handlePlayerEvent(message []byte) {
	var data playerEventResponse
	err := json.Unmarshal(message[1:], &data)
	if err != nil {
		app.errorLog.Printf("Could not unmarshal playerEvent: %s", err)
		return
	}

	if data.Body.EventType == threefoldRepetition {
		if hub.threefoldRepetition {
			hub.endGame(chess.ThreefoldRepetition)
			hub.sendMessageToAllClients(hub.currentGameState)
		}
		return
	}

	if isOneSidedEvent(data.Body.EventType) {
		hub.oneSidedEvent(messageIdentifier(message[0]), data.Body.EventType)
	} else if hub.offerActive == nil || hub.offerActive.event != data.Body.EventType {
		hub.makeNewEventOffer(messageIdentifier(message[0]), data.Body.EventType)
	} else if hub.offerActive != nil && byte(hub.offerActive.sender) != message[0] {
		hub.acceptEventOffer(data.Body.EventType)
	}

}

func (hub *MatchRoomHub) handleMessage(message []byte) {
	switch msgType := hub.getMessageType(message); msgType {
	case postMove:

		if hub.gameEnded {
			return
		}

		// Ignore messages from inactive player
		if message[0] != byte(hub.turn) {
			return
		}

		// Validate move and update
		err := hub.updateGameStateAfterMove(message)
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		hub.sendMessageToAllClients(hub.currentGameState)
		return

	case playerEvent:

		if hub.gameEnded {
			return
		}
		// @TODO: implement this fully
		hub.handlePlayerEvent(message)

	default:
		app.errorLog.Printf("Could not understand message: %s\n", message)
		return
	}
}

func (hub *MatchRoomHub) pingStatusMessage(playerColour string, isConnected bool, millisecondsUntilTimeout int64) ([]byte, error) {
	data := onPlayerConnectionChangeResponse{
		MessageType: connectionStatus,
		Body:        onPlayerConnectionChangeBody{PlayerColour: playerColour, IsConnected: isConnected, MillisecondsUntilTimeout: millisecondsUntilTimeout},
	}
	jsonStr, err := json.Marshal(data)
	if err != nil {
		app.errorLog.Printf("Unable to marshal pingStatus: %s", err)
	}
	return jsonStr, nil
}

func (hub *MatchRoomHub) setConnected(client *MatchRoomHubClient) {
	// Sets connections status of players and sends message to all clients
	if client.playerIdentifier == messageIdentifier(WhitePlayer) {
		hub.whitePlayerConnected = true
		hub.blackCanClaimTimeout = false
		hub.whitePlayerTimeout = nil
		pingMessage, err := hub.pingStatusMessage("white", true, 0)
		if err != nil {
			app.errorLog.Printf("Could not generate pingMessage: %s", err)
		}
		hub.sendMessageToAllClients(pingMessage)
	} else if client.playerIdentifier == messageIdentifier(BlackPlayer) {
		hub.blackPlayerConnected = true
		hub.whiteCanClaimTimeout = false
		hub.blackPlayerTimeout = nil
		pingMessage, err := hub.pingStatusMessage("black", true, 0)
		if err != nil {
			app.errorLog.Printf("Could not generate pingMessage: %s", err)
		}
		hub.sendMessageToAllClients(pingMessage)
	}
}

func (hub *MatchRoomHub) setDisconnected(client *MatchRoomHubClient) {
	// Sets connections status of players and sends message to all clients
	if client.playerIdentifier == messageIdentifier(WhitePlayer) {
		hub.whitePlayerConnected = false
		if !hub.gameEnded {
			hub.whitePlayerTimeout = time.After(pingTimeout)
			hub.whitePlayerTimeoutStarted = time.Now()
		}
		pingMessage, err := hub.pingStatusMessage("white", false, pingTimeout.Milliseconds())
		if err != nil {
			app.errorLog.Printf("Could not generate pingMessage: %s", err)
		}
		hub.sendMessageToAllClients(pingMessage)
	} else if client.playerIdentifier == messageIdentifier(BlackPlayer) {
		hub.blackPlayerConnected = false
		if !hub.gameEnded {
			hub.blackPlayerTimeout = time.After(pingTimeout)
			hub.blackPlayerTimeoutStarted = time.Now()
		}
		pingMessage, err := hub.pingStatusMessage("black", false, pingTimeout.Milliseconds())
		if err != nil {
			app.errorLog.Printf("Could not generate pingMessage: %s", err)
		}
		hub.sendMessageToAllClients(pingMessage)
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
			hub.setConnected(client)
			jsonStr, err := hub.getCurrentMatchStateForNewConnection(client.playerIdentifier)
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
			hub.setDisconnected(client)
			if !hub.hasActiveClients() {
				matchRoomHubManager.unregisterHub(hub.matchID)
				return
			}

		case <-hub.flagTimer:
			var gameOverStatus chess.GameOverStatusCode
			if hub.turn == playerTurn(WhiteTurn) {
				gameOverStatus = chess.WhiteFlagged
			} else {
				gameOverStatus = chess.BlackFlagged
			}
			err := hub.endGame(gameOverStatus)
			if err != nil {
				app.errorLog.Println(err)
				continue
			}

			hub.sendMessageToAllClients(hub.currentGameState)

		case <-hub.whitePlayerTimeout:
			hub.blackCanClaimTimeout = true

		case <-hub.blackPlayerTimeout:
			hub.whiteCanClaimTimeout = true

		case message := <-hub.broadcast:
			app.infoLog.Printf("WS Message: %s\n", message)

			hub.handleMessage(message)

		}
	}
}
