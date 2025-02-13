// Need hub and client
// Clients connect, if they are not players everything they send is ignored
// If clients are players, only listen to whose turn it is
// When a move is submitted, validate it and if it is valid broadcast the new match data to all clients
// If it is invalid, reject it

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Hub manager, opens new websockets for games in progress
type MatchRoomHubManager struct {

	// Registered hubs
	hubs map[int64]*MatchRoomHub
}

func newMatchRoomHubManager() *MatchRoomHubManager {
	return &MatchRoomHubManager{
		hubs: make(map[int64]*MatchRoomHub),
	}
}

func (hubManager *MatchRoomHubManager) registerClientToMatchRoomHub(conn *websocket.Conn, matchID int64, playerID *int64) (*MatchRoomHubClient, error) {
	val, ok := hubManager.hubs[matchID]

	// If hub not running, run it
	if !ok {
		newHub, err := newMatchRoomHub(matchID)
		if err != nil {
			app.errorLog.Println(err)
			return nil, err
		}
		hubManager.hubs[matchID] = newHub
		go hubManager.hubs[matchID].run()
		val = hubManager.hubs[matchID]
	}

	var playerCode playerCodeEnum = Spectator

	if playerID == nil {
		// Do nothing
	} else if *playerID == val.whitePlayerID {
		playerCode = WhitePieces
	} else if *playerID == val.blackPlayerID {
		playerCode = BlackPieces
	}

	return &MatchRoomHubClient{hub: val, conn: conn, playerCode: playerCode, send: make(chan []byte, 256)}, nil
}

type MatchStateHistory struct {
	FEN                                  string `json:"FEN"`
	LastMove                             [2]int `json:"lastMove"`
	AlgebraicNotation                    string `json:"algebraicNotation"`
	WhitePlayerTimeRemainingMilliseconds int64  `json:"whitePlayerTimeRemainingMilliseconds"`
	BlackPlayerTimeRemainingMilliseconds int64  `json:"blackPlayerTimeRemainingMilliseconds"`
}

var matchRoomHubManager = newMatchRoomHubManager()

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

	hasWhitePlayerMadeFirstMove bool

	hasBlackPlayerMadeFirstMove bool

	turn byte // byte(0) is white, byte(1) is black

	currentGameState []byte

	current_fen string

	moveHistory []MatchStateHistory

	timeOfLastMove time.Time

	flagTimer <-chan time.Time
}

func newMatchRoomHub(matchID int64) (*MatchRoomHub, error) {
	matchState, err := app.liveMatches.GetFromMatchID(matchID)

	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	var last_move_move, last_move_piece int
	if !matchState.LastMoveMove.Valid {
		last_move_move = 0
	}

	if !matchState.LastMovePiece.Valid {
		last_move_piece = 0
	}

	var turn byte

	currentGameState := [1]postChessMoveReply{{
		MatchStateHistory: []MatchStateHistory{{
			FEN:                                  matchState.CurrentFEN,
			LastMove:                             [2]int{last_move_move, last_move_piece},
			AlgebraicNotation:                    intToAlgebraicNotation(last_move_move),
			WhitePlayerTimeRemainingMilliseconds: time.Duration(3 * time.Minute).Milliseconds(),
			BlackPlayerTimeRemainingMilliseconds: time.Duration(3 * time.Minute).Milliseconds(),
		}},
		GameOverStatus: Ongoing,
	}}

	if strings.Split(matchState.CurrentFEN, " ")[1] == "w" {
		turn = byte(0)
	} else {
		turn = byte(1)
	}

	jsonStr, err := json.Marshal(currentGameState)
	if err != nil {
		app.errorLog.Printf("Error marshalling JSON: %v\n", err)
		return nil, err
	}

	return &MatchRoomHub{
		matchID:                     matchID,
		broadcast:                   make(chan []byte),
		register:                    make(chan *MatchRoomHubClient),
		unregister:                  make(chan *MatchRoomHubClient),
		clients:                     make(map[*MatchRoomHubClient]bool),
		whitePlayerID:               matchState.WhitePlayerID,
		blackPlayerID:               matchState.BlackPlayerID,
		whitePlayerTimeRemaining:    time.Duration(3 * time.Minute),
		blackPlayerTimeRemaining:    time.Duration(3 * time.Minute),
		hasWhitePlayerMadeFirstMove: false,
		hasBlackPlayerMadeFirstMove: false,
		turn:                        turn,
		currentGameState:            jsonStr,
		current_fen:                 matchState.CurrentFEN,
		moveHistory:                 currentGameState[0].MatchStateHistory,
		timeOfLastMove:              time.Now(),
	}, nil
}

type wsChessMove struct {
	Piece           int    `json:"piece"`
	Move            int    `json:"move"`
	PromotionString string `json:"promotionString"`
}

func (hub *MatchRoomHub) run() {
	defer app.infoLog.Println("Hub stopped")
	for {
		app.infoLog.Println("Hub running")
		select {
		// Clients get currentGameState on register
		case client := <-hub.register:
			hub.clients[client] = true
			client.send <- hub.currentGameState
		case client := <-hub.unregister:
			if _, ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				close(client.send)
			}
		case <-hub.flagTimer:
			var gameState []postChessMoveReply
			err := json.Unmarshal(hub.currentGameState, &gameState)
			if err != nil {
				app.errorLog.Printf("Error unmarshalling JSON: %v\n", err)
				continue
			}

			if hub.turn == byte(0) {
				gameState[0].GameOverStatus = WhiteFlagged
			} else {
				gameState[0].GameOverStatus = BlackFlagged
			}

			jsonStr, err := json.Marshal(gameState)
			if err != nil {
				app.errorLog.Printf("Error marshalling JSON: %v\n", err)
				continue
			}

			hub.currentGameState = jsonStr

			for client := range hub.clients {
				select {
				case client.send <- jsonStr:
				default:
					close(client.send)
					delete(hub.clients, client)
				}
			}
		case message := <-hub.broadcast:
			app.infoLog.Println("WS Message")
			app.infoLog.Println(message)
			app.infoLog.Println(hub.turn)
			if message[0] != hub.turn {
				continue
			}

			var chessMove wsChessMove
			// Parse and validate move
			err := json.Unmarshal(message[1:], &chessMove)
			if err != nil {
				app.errorLog.Printf("Error unmarshalling JSON: %v\n", err)
				continue
			}

			var validMove = IsMoveValid(hub.current_fen, chessMove.Piece, chessMove.Move)
			if !validMove {
				continue
			}

			if message[0] == byte(0) {
				if hub.hasWhitePlayerMadeFirstMove {
					hub.whitePlayerTimeRemaining -= time.Since(hub.timeOfLastMove)
					hub.whitePlayerTimeRemaining += time.Duration(2 * time.Second)

				} else {
					hub.hasWhitePlayerMadeFirstMove = true
				}
			} else {
				if hub.hasBlackPlayerMadeFirstMove {
					hub.blackPlayerTimeRemaining -= time.Since(hub.timeOfLastMove)
					hub.blackPlayerTimeRemaining += time.Duration(2 * time.Second)
				} else {
					hub.hasBlackPlayerMadeFirstMove = true
				}
			}

			newFEN, gameOverStatus, algebraicNotation := getFENAfterMove(hub.current_fen, chessMove.Piece, chessMove.Move, chessMove.PromotionString)
			// Need to put move into db
			data := []postChessMoveReply{
				{
					MatchStateHistory: append(hub.moveHistory, MatchStateHistory{
						FEN:                                  newFEN,
						LastMove:                             [2]int{chessMove.Piece, chessMove.Move},
						AlgebraicNotation:                    algebraicNotation,
						WhitePlayerTimeRemainingMilliseconds: hub.whitePlayerTimeRemaining.Milliseconds(),
						BlackPlayerTimeRemainingMilliseconds: hub.blackPlayerTimeRemaining.Milliseconds(),
					}),
					GameOverStatus: gameOverStatus,
				},
			}

			jsonStr, err := json.Marshal(data)
			if err != nil {
				app.errorLog.Printf("Error marshalling JSON: %v\n", err)
				continue
			}

			hub.current_fen = newFEN
			hub.currentGameState = jsonStr
			hub.moveHistory = append(hub.moveHistory, MatchStateHistory{
				FEN:                                  newFEN,
				LastMove:                             [2]int{chessMove.Piece, chessMove.Move},
				AlgebraicNotation:                    intToAlgebraicNotation(chessMove.Move),
				WhitePlayerTimeRemainingMilliseconds: hub.whitePlayerTimeRemaining.Milliseconds(),
				BlackPlayerTimeRemainingMilliseconds: hub.blackPlayerTimeRemaining.Milliseconds(),
			})
			go app.liveMatches.UpdateFENForLiveMatch(hub.matchID, newFEN, chessMove.Piece, chessMove.Move)

			hub.timeOfLastMove = time.Now()

			if message[0] == byte(0) && hub.hasBlackPlayerMadeFirstMove {
				hub.flagTimer = time.After(hub.blackPlayerTimeRemaining)
			} else if message[0] == byte(1) && hub.hasWhitePlayerMadeFirstMove {
				hub.flagTimer = time.After(hub.whitePlayerTimeRemaining)
			}

			for client := range hub.clients {
				select {
				case client.send <- jsonStr:
				default:
					close(client.send)
					delete(hub.clients, client)
				}
			}

			// Change turn
			if hub.turn == byte(0) {
				hub.turn = byte(1)
			} else {
				hub.turn = byte(0)
			}

		}
	}
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:5173"
	},
}

type MatchRoomHubClient struct {
	hub *MatchRoomHub

	// The websocket connection
	conn *websocket.Conn

	playerCode playerCodeEnum

	// Buffered channel of outbound messages
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *MatchRoomHubClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		app.infoLog.Println("Checking for client sent messages")
		_, message, err := c.conn.ReadMessage()
		app.infoLog.Println(message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Append sender to message
		sender := []byte{byte(c.playerCode)}
		message = append(sender, message...)
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *MatchRoomHubClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type sendPlayerCode struct {
	PlayerCode playerCodeEnum `json:"playerCode"`
}

// serveWs handles websocket requests from the peer.
func serveMatchroomWs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	app.infoLog.Println("WS Request")

	matchID, err := strconv.ParseInt(r.PathValue("matchID"), 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not find match", http.StatusInternalServerError)
		return
	}

	var playerid string
	var rawID int64
	var playerIDasInt *int64

	playerid, err = ReadSigned(r, app.secretKey, "playerid")
	if err == nil {
		rawID, err = strconv.ParseInt(playerid, 10, 64)
		playerIDasInt = &rawID
	}

	if err != nil {
		log.Printf("Error occured whilst parsing playerID: %v\n", err)
		playerIDasInt = nil
	}

	var conn *websocket.Conn

	conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	var client *MatchRoomHubClient

	client, err = matchRoomHubManager.registerClientToMatchRoomHub(conn, matchID, playerIDasInt)
	if err != nil {
		app.errorLog.Println(err)
		conn.WriteMessage(websocket.CloseMessage, []byte{})
		conn.Close()
		return
	}

	// Send player code
	var codeMessage [1]sendPlayerCode = [1]sendPlayerCode{{PlayerCode: client.playerCode}}
	var jsonStr []byte
	jsonStr, err = json.Marshal(codeMessage)
	if err != nil {
		app.errorLog.Println(err)
		conn.WriteMessage(websocket.CloseMessage, []byte{})
		conn.Close()
		return
	}

	// Is this blocking? Should the goroutines come first? Should we use a buffered channel?
	client.send <- jsonStr
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
