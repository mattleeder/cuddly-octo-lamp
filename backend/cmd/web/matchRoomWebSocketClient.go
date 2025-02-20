package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type messageIdentifier byte

const (
	WhitePlayer = byte(iota)
	BlackPlayer = byte(iota)
	Spectator   = byte(iota)
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 20 * time.Second

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

	// White, black or spectator
	playerIdentifier messageIdentifier

	// Buffered channel of outbound messages
	send chan []byte
}

type sendPlayerCodeResponse struct {
	MessageType hubMessageType     `json:"messageType"`
	Body        sendPlayerCodeBody `json:"body"`
}

type sendPlayerCodeBody struct {
	PlayerCode messageIdentifier `json:"playerCode"`
}

func pingClient(conn *websocket.Conn) {
	ticker := time.NewTicker(5 * time.Second) // Ping every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send a ping message to the client
			err := conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				app.errorLog.Println("Ping error:", err)
				return
			}
		}
	}
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
		app.infoLog.Println("Client closed")
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	// go pingClient(c.conn)
	for {
		_, message, err := c.conn.ReadMessage()
		app.infoLog.Println(message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				app.errorLog.Printf("error: %v", err)
			}
			break
		}

		// Append sender to message
		sender := []byte{byte(c.playerIdentifier)}
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

// serveWs handles websocket requests from the peer.
func serveMatchroomWs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	app.infoLog.Println("WS Request")

	matchID, err := strconv.ParseInt(r.PathValue("matchID"), 10, 64)
	if err != nil {
		app.errorLog.Println(err)
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
		app.serverError(w, err)
		return
	}

	var conn *websocket.Conn

	conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var client *MatchRoomHubClient

	client, err = matchRoomHubManager.registerClientToMatchRoomHub(conn, matchID, playerIDasInt)
	if err != nil {
		app.websocketError(conn, err)
		return
	}

	// Send player code
	var codeMessage = sendPlayerCodeResponse{
		MessageType: sendPlayerCode,
		Body:        sendPlayerCodeBody{PlayerCode: client.playerIdentifier},
	}
	var jsonStr []byte
	jsonStr, err = json.Marshal(codeMessage)
	if err != nil {
		app.websocketError(conn, err)
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
