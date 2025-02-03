package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
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
	CurrentFEN string
	Piece      int
	Move       int
	NewFEN     string
}

type postChessMoveReply struct {
	IsValid                  bool   `json:"isValid"`
	NewFEN                   string `json:"newFEN"`
	LastMove                 [2]int `json:"lastMove"`
	MoveWillTriggerPromotion bool   `json:"triggerPromotion"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
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
	var moves, captures, triggerPromotion = getValidMovesForPiece(chessMoveData.Piece, currentGameState)

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
		newFEN := getFENAfterMove(chessMove.CurrentFEN, chessMove.Piece, chessMove.Move)
		data = postChessMoveReply{IsValid: true, NewFEN: newFEN, LastMove: [2]int{chessMove.Piece, chessMove.Move}}
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

func main() {
	http.HandleFunc("/", chessMoveValidationHandler)
	http.HandleFunc("/getMoves", getChessMovesHandler)
	http.HandleFunc("/makeMove", postChessMoveHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
