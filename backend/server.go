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
	time      int
	increment int
	playerID  int
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
	value := signedValue[:sha256.Size]

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

	start := time.Now()

	var chessMove userMoveData
	w.Header().Set("Access-Control-Allow-Origin", "*")

	err := json.NewDecoder(r.Body).Decode(&chessMove)
	if err != nil {
		http.Error(w, "Could not decode request", http.StatusInternalServerError)
		return
	}

	fmt.Println(chessMove)
	fmt.Printf("Received body: %+v\n", chessMove)

	var validMove = IsMoveValid(chessMove.Fen, chessMove.Piece, chessMove.Move)

	fmt.Fprintln(w, validMove)

	elapsed := time.Since(start)
	fmt.Printf("Took %s\n", elapsed)
}

func getChessMovesHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var chessMoveData getChessMoveData
	w.Header().Set("Access-Control-Allow-Origin", "*")

	err := json.NewDecoder(r.Body).Decode(&chessMoveData)
	if err != nil {
		http.Error(w, "Could not decode request", http.StatusInternalServerError)
		return
	}

	fmt.Println(chessMoveData)
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

	start := time.Now()

	var chessMove postChessMove
	w.Header().Set("Access-Control-Allow-Origin", "*")

	err := json.NewDecoder(r.Body).Decode(&chessMove)
	if err != nil {
		http.Error(w, "Could not decode request", http.StatusInternalServerError)
		return
	}

	fmt.Println(chessMove)
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
	start := time.Now()

	var joinQueue joinQueueRequest
	w.Header().Set("Access-Control-Allow-Origin", "*")

	err := json.NewDecoder(r.Body).Decode(&joinQueue)
	if err != nil {
		http.Error(w, "Could not decode join queue request", http.StatusInternalServerError)
		return
	}

	playerid, err := ReadSigned(r, secretKey, "playerid")
	if err == http.ErrNoCookie {
		playerid = strconv.FormatInt(generateNewPlayerId(), 10)

		cookie := http.Cookie{
			Name:  "playerid",
			Value: playerid,
		}

		err = WriteSigned(w, cookie, secretKey)
		if err != nil {
			http.Error(w, "Unable to decode or write cookie", http.StatusInternalServerError)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "playerid",
		Value: "",
	})

	fmt.Println(joinQueue)
	fmt.Printf("Received body: %+v\n", joinQueue)

	elapsed := time.Since(start)
	fmt.Printf("Took %s\n", elapsed)
}

func main() {

	secretKey = []byte("}\xa4\xc3\x85D\x89\xb75\xf0\xe6\xcf\xcaZ\x00k\x88\xe4\x8f\xd0\xd6\x95\x0e\xa6\xf9\xc2;!\xa2\xc4[\xca\x91")

	fmt.Println(secretKey)

	os.Remove("./chess_site.db")

	db := initDatabase()

	defer db.Close()

	for i := 0; i < 5; i++ {
		insertNewLiveMatch()
	}

	rows, err := db.Query("select id, current_fen from live_matches")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	for rows.Next() {
		var id int
		var fen string
		err = rows.Scan(&id, &fen)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, fen)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", chessMoveValidationHandler)
	http.HandleFunc("/getMoves", getChessMovesHandler)
	http.HandleFunc("/makeMove", postChessMoveHandler)
	http.HandleFunc("/joinQueue", joinQueueHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
