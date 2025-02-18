package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
	takeback       = "takeback"
	takebackAccept = "takebackAccept"
	draw           = "draw"
	drawAccept     = "drawAccept"
	resign         = "resign"
	extraTime      = "extraTime"
	abort          = "abort"
	rematch        = "rematch"
	rematchAccept  = "rematchAccept"
)

// Bodies

type onConnectBody struct {
	MatchStateHistory    []MatchStateHistory `json:"matchStateHistory"`
	GameOverStatusCode   gameOverStatusCode  `json:"gameOverStatus"`
	ThreefoldRepetition  bool                `json:"threefoldRepetition"`
	WhitePlayerConnected bool                `json:"whitePlayerConnected"`
	BlackPlayerConnected bool                `json:"blackPlayerConnected"`
}

type onMoveBody struct {
	MatchStateHistory   []MatchStateHistory `json:"matchStateHistory"`
	GameOverStatusCode  gameOverStatusCode  `json:"gameOverStatus"`
	ThreefoldRepetition bool                `json:"threefoldRepetition"`
}

type onPlayerConnectionChangeBody struct {
	PlayerColour string `json:"playerColour"`
	IsConnected  bool   `json:"isConnected"`
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

type MatchStateHistory struct {
	FEN                                  string `json:"FEN"`
	LastMove                             [2]int `json:"lastMove"`
	AlgebraicNotation                    string `json:"algebraicNotation"`
	WhitePlayerTimeRemainingMilliseconds int64  `json:"whitePlayerTimeRemainingMilliseconds"`
	BlackPlayerTimeRemainingMilliseconds int64  `json:"blackPlayerTimeRemainingMilliseconds"`
}

type pingStatus struct {
	PlayerColour string `json:"playerColour"`
	IsConnected  bool   `json:"isConnected"`
}

type offerInfo struct {
	sender messageIdentifier
	event  eventType
}

const pingTimeout = 10 * time.Second

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

	blackPlayerTimeout <-chan time.Time

	offerActive *offerInfo
}

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

	currentGameState := onMoveResponse{
		MessageType: onMove,
		Body: onMoveBody{
			MatchStateHistory:   matchStateHistory,
			GameOverStatusCode:  Ongoing,
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

func (hub *MatchRoomHub) updateGameStateAfterFlag() (err error) {
	var gameState onMoveResponse
	err = json.Unmarshal(hub.currentGameState, &gameState)
	if err != nil {
		app.errorLog.Printf("Error unmarshalling JSON: %v\n", err)
		return err
	}

	var outcome int
	if hub.turn == playerTurn(WhiteTurn) {
		gameState.Body.GameOverStatusCode = WhiteFlagged
		outcome = 2
	} else {
		gameState.Body.GameOverStatusCode = BlackFlagged
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

func (hub *MatchRoomHub) updateGameStateAfterMove(message []byte) (err error) {

	// Parse Message
	var chessMove postMoveResponse
	err = json.Unmarshal(message[1:], &chessMove)
	if err != nil {
		return errors.New(fmt.Sprintf("Error unmarshalling JSON: %v\n", err))
	}

	// Validate Move
	var validMove = IsMoveValid(hub.current_fen, chessMove.Body.Piece, chessMove.Body.Move)
	if !validMove {
		return errors.New("Move is not valid")
	}

	// Calcuate new time remaining
	hub.updateTimeRemaining()

	// Calculate reply variables
	newFEN, gameOverStatus, algebraicNotation := getFENAfterMove(hub.current_fen, chessMove.Body.Piece, chessMove.Body.Move, chessMove.Body.PromotionString)
	var threefoldRepetition = false
	splitFEN := strings.Join(strings.Split(newFEN, " ")[:4], " ")
	hub.fenFreqMap[splitFEN] += 1
	if hub.fenFreqMap[splitFEN] >= 3 {
		threefoldRepetition = true
	}

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
	if gameOverStatus != Ongoing {
		go hub.updateMatchThenMoveToFinished(newFEN, chessMove.Body.Piece, chessMove.Body.Move, gameOverStatus, hub.turn, matchStateHistoryData)
	} else {
		go app.liveMatches.UpdateLiveMatch(hub.matchID, newFEN, chessMove.Body.Piece, chessMove.Body.Move, hub.whitePlayerTimeRemaining.Milliseconds(), hub.blackPlayerTimeRemaining.Milliseconds(), matchStateHistoryData, hub.timeOfLastMove)
	}

	return nil
}

func (hub *MatchRoomHub) getCurrentMatchStateForNewConnection() (jsonStr []byte, err error) {
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

	var response = onConnectResponse{
		MessageType: onConnect,
		Body: onConnectBody{
			MatchStateHistory:    gameState.Body.MatchStateHistory,
			GameOverStatusCode:   gameState.Body.GameOverStatusCode,
			ThreefoldRepetition:  gameState.Body.ThreefoldRepetition,
			WhitePlayerConnected: hub.whitePlayerConnected,
			BlackPlayerConnected: hub.blackPlayerConnected,
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
	if message[0] == byte(WhitePlayer) || message[0] == byte(BlackPlayer) {
		return postMove
	}
	app.errorLog.Printf("Unknown message type\n")
	return unknown
}

func (hub *MatchRoomHub) makeDraw(reason gameOverStatusCode) []byte {
	var data onMoveResponse
	err := json.Unmarshal(hub.currentGameState, &data)
	if err != nil {
		app.errorLog.Printf("Error unmarshaling hub.currentGameState: %s", err)
	}

	var jsonStr []byte
	data.Body.GameOverStatusCode = reason
	jsonStr, err = json.Marshal(data)
	if err != nil {
		app.errorLog.Printf("Error marshaling onMoveResponse: %s", err)
	}

	hub.currentGameState = jsonStr

	go app.liveMatches.MoveMatchToPastMatches(hub.matchID, 0)

	return jsonStr
}

// @TODO: implement this
func (hub *MatchRoomHub) takeBack() []byte {
	return nil
}

func (hub *MatchRoomHub) acceptEventOffer(event eventType) []byte {
	// @TODO implement
	switch event {
	case takeback:
		return hub.takeBack()

	case draw:
		return hub.makeDraw(Draw)
	}
}

func (hub *MatchRoomHub) endGame(reason gameOverStatusCode) {
	// Function should
	// End game
	// Send final update to players
	// Handle database changes
}

func (hub *MatchRoomHub) makeNewEventOffer(sender messageIdentifier, event eventType) []byte {
	hub.offerActive = &offerInfo{sender, event}
	var responseFrom string
	if sender == messageIdentifier(WhitePlayer) {
		responseFrom = "white"
	} else {
		responseFrom = "black"
	}

	response := opponentEventResponse{
		MessageType: opponentEvent,
		Body:        opponentEventBody{Sender: responseFrom, EventType: event},
	}

	jsonStr, err := json.Marshal(response)
	if err != nil {
		app.errorLog.Printf("Could not marshal opponentEventResponse: %s\n", err)
		return nil
	}
	return jsonStr
}

func (hub *MatchRoomHub) handlePlayerEvent(message []byte) []byte {
	var data playerEventResponse
	err := json.Unmarshal(message[1:], &data)
	if err != nil {
		app.errorLog.Printf("Could not unmarshal playerEvent: %s", err)
		return nil
	}

	if hub.offerActive == nil || hub.offerActive.event != data.Body.EventType {
		return hub.makeNewEventOffer(messageIdentifier(message[0]), data.Body.EventType)
	} else if hub.offerActive != nil && byte(hub.offerActive.sender) != message[0] {
		return hub.acceptEventOffer(data.Body.EventType)
	}
}

func (hub *MatchRoomHub) handleMessage(message []byte) (response []byte) {
	switch msgType := hub.getMessageType(message); msgType {
	case postMove:

		// Ignore messages from inactive player
		if message[0] != byte(hub.turn) {
			return nil
		}

		// Validate move and update
		err := hub.updateGameStateAfterMove(message)
		if err != nil {
			app.errorLog.Println(err)
			return nil
		}

		return hub.currentGameState

	case playerEvent:
		// @TODO: implement this fully
		return hub.handlePlayerEvent(message)

	default:
		app.errorLog.Printf("Could not understand message: %s\n", message)
		return nil
	}
}

func (hub *MatchRoomHub) pingStatusMessage(playerColour string, isConnected bool) ([]byte, error) {
	data := onPlayerConnectionChangeResponse{
		MessageType: connectionStatus,
		Body:        onPlayerConnectionChangeBody{PlayerColour: playerColour, IsConnected: isConnected},
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
		hub.whitePlayerTimeout = nil
		pingMessage, err := hub.pingStatusMessage("white", true)
		if err != nil {
			app.errorLog.Printf("Could not generate pingMessage: %s", err)
		}
		hub.sendMessageToAllClients(pingMessage)
	} else if client.playerIdentifier == messageIdentifier(BlackPlayer) {
		hub.blackPlayerConnected = true
		hub.blackPlayerTimeout = nil
		pingMessage, err := hub.pingStatusMessage("black", true)
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
		hub.whitePlayerTimeout = time.After(pingTimeout)
		pingMessage, err := hub.pingStatusMessage("white", false)
		if err != nil {
			app.errorLog.Printf("Could not generate pingMessage: %s", err)
		}
		hub.sendMessageToAllClients(pingMessage)
	} else if client.playerIdentifier == messageIdentifier(BlackPlayer) {
		hub.blackPlayerConnected = false
		hub.blackPlayerTimeout = time.After(pingTimeout)
		pingMessage, err := hub.pingStatusMessage("black", false)
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
			hub.setDisconnected(client)
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

		case <-hub.whitePlayerTimeout:
		case <-hub.blackPlayerTimeout:
			// @TODO, add logic
			continue

		case message := <-hub.broadcast:
			app.infoLog.Printf("WS Message: %s\n", message)

			response := hub.handleMessage(message)
			app.infoLog.Printf("Response: %s\n", response)
			if response != nil {
				hub.sendMessageToAllClients(response)
			}

		}
	}
}
