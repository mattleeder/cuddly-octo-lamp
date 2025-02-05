package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

var secretKey []byte

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

	var joinQueue joinQueueRequest
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	err := json.NewDecoder(r.Body).Decode(&joinQueue)
	if err != nil {
		http.Error(w, "Could not decode join queue request", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Received body: %+v\n", joinQueue)
	fmt.Println(r.Cookies())

	// Read cookie
	playerid, err := ReadSigned(r, secretKey, "playerid")

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

		err = WriteSigned(w, cookie, secretKey)
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
		err = addPlayerToQueue(playerIDasInt, joinQueue.Time, joinQueue.Increment)
	} else {
		err = removePlayerFromQueue(playerIDasInt, joinQueue.Time, joinQueue.Increment)
	}

	if err != nil {
		http.Error(w, "Unable to add/remove player to/from queue", http.StatusInternalServerError)
		return
	}

	elapsed := time.Since(start)
	fmt.Printf("Took %s\n\n", elapsed)
}

func main() {

	secretKey = []byte("}\xa4\xc3\x85D\x89\xb75\xf0\xe6\xcf\xcaZ\x00k\x88\xe4\x8f\xd0\xd6\x95\x0e\xa6\xf9\xc2;!\xa2\xc4[\xca\x91")

	fmt.Println(secretKey)

	os.Remove("./chess_site.db")

	db := initDatabase()

	defer db.Close()

	http.HandleFunc("/", chessMoveValidationHandler)
	http.HandleFunc("/getMoves", getChessMovesHandler)
	http.HandleFunc("/makeMove", postChessMoveHandler)
	http.HandleFunc("/joinQueue", joinQueueHandler)

	log.Fatal(http.ListenAndServeTLS(":8080", "burrchess.crt", "burrchess.key", nil))
}
