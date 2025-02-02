package main

import (
	"fmt"
	"time"
	"unicode"
)

/*
Gos basic types

bool

string

int  int8  int16  int32  int64
uint uint8 uint16 uint32 uint64 uintptr

byte // alias for uint8

rune // alias for int32
     // represents a Unicode code point

float32 float64

complex64 complex128
*/

/*

type Piece struct {
	position			int
    moves				[]int
	attacks				[]int
	attacksSameAsMoves 	bool
	isPinned			bool
}

Up			-8
Down		+8
Left		-1
Right		+1

Up Left		-9
Up Right	-7
Down Left	+7
Down Right	+9
*/

type pieceColour int

const (
	White = iota
	Black
)

var pieceColourName = map[pieceColour]string{
	White: "white",
	Black: "black",
}

type pieceVariant int

const (
	Pawn = iota
	Knight
	Bishop
	Rook
	Queen
	King
)

var pieceVariantName = map[pieceVariant]string{
	Pawn:   "pawn",
	Knight: "kinght",
	Bishop: "bishop",
	Rook:   "rook",
	Queen:  "queen",
	King:   "king",
}

type pieceType struct {
	position               int
	colour                 pieceColour
	variant                pieceVariant
	moves, attacks         []int
	moveRange, attackRange int
	movesEqualsAttacks     bool
}

type square struct {
	piece          *pieceType
	whiteAttacking bool
	blackAttacking bool
}

type gameState struct {
	board                   [64]square
	turn                    pieceColour
	blackKingPosition       int
	whiteKingPosition       int
	blackCanKingSideCastle  bool
	blackCanQueenSideCastle bool
	whiteCanKingSideCastle  bool
	whiteCanQueenSideCastle bool
	enPassantTargetSquare   int
	enPassantAvailable      bool
	halfMoveClock           int
	fullMoveNumber          int
}

func filter[T any](arr []T, fn func(T) bool) []T {
	result := []T{}
	for _, v := range arr {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

func lambdaMapGet[T comparable, U any](m map[T]U) func(T) U {
	return func(key T) U {
		return m[key]
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func getRow(position int) int {
	return position / 8
}

func getCol(position int) int {
	return position % 8
}

func defaultSquare() square {
	return square{nil, false, false}
}

func createPawn(position int, colour pieceColour) pieceType {
	var moves, attacks []int
	if colour == White {
		moves = []int{-8}
		attacks = []int{-7, -9}
	} else {
		moves = []int{8}
		attacks = []int{7, 9}
	}
	return pieceType{
		position:           position,
		colour:             colour,
		variant:            Pawn,
		moves:              moves,
		attacks:            attacks,
		moveRange:          1,
		attackRange:        1,
		movesEqualsAttacks: false,
	}
}

func createKnight(position int, colour pieceColour) pieceType {
	return pieceType{
		position:           position,
		colour:             colour,
		variant:            Knight,
		moves:              []int{-15, -6, 10, 17, 15, 6, -10, -17},
		moveRange:          1,
		movesEqualsAttacks: true,
	}
}

func createBishop(position int, colour pieceColour) pieceType {
	return pieceType{
		position:           position,
		colour:             colour,
		variant:            Bishop,
		moves:              []int{-9, -7, 7, 9},
		moveRange:          7,
		movesEqualsAttacks: true,
	}
}

func createRook(position int, colour pieceColour) pieceType {
	return pieceType{
		position:           position,
		colour:             colour,
		variant:            Rook,
		moves:              []int{-8, 8, -1, 1},
		moveRange:          7,
		movesEqualsAttacks: true,
	}
}

func createQueen(position int, colour pieceColour) pieceType {
	return pieceType{
		position:           position,
		colour:             colour,
		variant:            Queen,
		moves:              []int{-9, -8, -7, 1, 9, 8, 7, -1},
		moveRange:          7,
		movesEqualsAttacks: true,
	}
}

func createKing(position int, colour pieceColour) pieceType {
	return pieceType{
		position:           position,
		colour:             colour,
		variant:            King,
		moves:              []int{-9, -8, -7, 1, 9, 8, 7, -1},
		moveRange:          1,
		movesEqualsAttacks: true,
	}
}

func createPiece(position int, colour pieceColour, variant pieceVariant) *pieceType {
	var newPiece pieceType
	var p *pieceType = &newPiece
	switch variant {
	case Pawn:
		newPiece = createPawn(position, colour)
	case Knight:
		newPiece = createKnight(position, colour)
	case Bishop:
		newPiece = createBishop(position, colour)
	case Rook:
		newPiece = createRook(position, colour)
	case Queen:
		newPiece = createQueen(position, colour)
	case King:
		newPiece = createKing(position, colour)
	default:
		fmt.Println("Unknown piece variant")
	}

	return p
}

var boardEdges = make(map[int]bool)

func getBoardEdges() map[int]bool {
	for i := 0; i < 8; i++ {
		boardEdges[i] = true
		boardEdges[i+56] = true
	}
	for i := 0; i < 8; i++ {
		boardEdges[i*8] = true
		boardEdges[(i*8)+7] = true
	}
	return boardEdges
}

func hasMoveCrossedEdge(piecePosition int, movePosition int, pieceVariant pieceVariant) bool {
	var pieceRow = getRow(piecePosition)
	var pieceCol = getCol(piecePosition)
	var moveRow = getRow(movePosition)
	var moveCol = getCol(movePosition)
	var rowChange = abs(pieceRow - moveRow)
	var colChange = abs(pieceCol - moveCol)

	switch pieceVariant {
	case Pawn:
		/*Col change <= 1*/
		if colChange > 1 {
			return true
		}
	case Knight:
		/*Sum of row and colchange == 3*/
		if rowChange+colChange != 3 {
			return true
		}
	case Bishop:
		/*Row change must equal col change*/
		if rowChange != colChange {
			return true
		}
	case Rook:
		/*Only Row or Col must change*/
		if rowChange > 0 && colChange > 0 {
			return true
		}
	case Queen:
		/*Only Row or Col must change || Row change must equal col change*/
		if (rowChange != colChange) && (rowChange > 0 && colChange > 0) {
			return true
		}
	case King:
		/*Max col or row change should be 1*/
		if rowChange > 1 || colChange > 1 {
			return true
		}
	default:
		fmt.Println("Uknown piece type")

	}

	return false
}

func isSquareInBoard(square int) bool {
	return square >= 0 && square < 64
}

func isSquareUnderAttack(board [64]square, position int, defendingColour pieceColour) bool {
	var attackingSquare int
	var targetPiece *pieceType

	// @TODO
	/*Check Pawns*/
	var possibleAttackingPawnPositions [2]int
	if defendingColour == White {
		possibleAttackingPawnPositions = [2]int{-7, -9}
	} else {
		possibleAttackingPawnPositions = [2]int{7, 9}
	}
	for i := range possibleAttackingPawnPositions {
		attackingSquare = position + possibleAttackingPawnPositions[i]

		/*Check if square outside board*/
		if !isSquareInBoard(attackingSquare) {
			continue
		}

		/*Check if move goes over edge*/
		if hasMoveCrossedEdge(position, attackingSquare, Pawn) {
			continue
		}

		targetPiece = board[attackingSquare].piece
		if targetPiece != nil && targetPiece.colour != defendingColour && targetPiece.variant == Pawn {
			return true
		}
	}

	/*Check Knights*/
	var dummyKnight = createKnight(position, defendingColour)
	for _, attackDirection := range dummyKnight.moves {
		attackingSquare = position + attackDirection
		if !isSquareInBoard(attackingSquare) {
			continue
		}
		if hasMoveCrossedEdge(position, attackingSquare, Knight) {
			continue
		}
		targetPiece = board[attackingSquare].piece
		if targetPiece != nil && targetPiece.variant == Knight && targetPiece.colour != defendingColour {
			return true
		}
	}

	/*Check Rooks*/
	var dummyRook = createRook(position, defendingColour)
	for _, attackDirection := range dummyRook.moves {
		for attackRange := 1; attackRange <= dummyRook.moveRange; attackRange++ {
			attackingSquare = position + (attackDirection * attackRange)
			if !isSquareInBoard(attackingSquare) {
				break
			}
			if hasMoveCrossedEdge(position, attackingSquare, Rook) {
				break
			}
			targetPiece = board[attackingSquare].piece
			if targetPiece == nil {
				continue
			}
			if (targetPiece.variant == Rook || targetPiece.variant == Queen) && targetPiece.colour != defendingColour {
				return true
			}
			break
		}
	}

	/*Check Bishops*/
	var dummyBishop = createBishop(position, defendingColour)
	for _, attackDirection := range dummyBishop.moves {
		for attackRange := 1; attackRange <= dummyBishop.moveRange; attackRange++ {
			attackingSquare = position + (attackDirection * attackRange)
			if !isSquareInBoard(attackingSquare) {
				continue
			}
			if hasMoveCrossedEdge(position, attackingSquare, Bishop) {
				break
			}
			targetPiece = board[attackingSquare].piece
			if targetPiece == nil {
				continue
			}
			if (targetPiece.variant == Bishop || targetPiece.variant == Queen) && targetPiece.colour != defendingColour {
				return true
			}
			break
		}
	}

	return false
}

func getMovesandCapturesForPiece(board [64]square, piecePosition int, currentGameState gameState) (moves []int, captures []int) {
	/*
		Pawns must check for: promotion, double move, en passant.
		Kings must check for: castling, square attacked
		Promotions must check for pins?

		All piece must check for: checks, blocks, pins
	*/

	var currentSquare int
	var targetPiece *pieceType
	var moveDirection int
	var enpassantActive = currentGameState.enPassantAvailable
	var enpassantSquare = currentGameState.enPassantTargetSquare
	var piece = board[piecePosition].piece

	var shortCastleAvailable bool
	var longCastleAvailable bool
	var friendlyKingPosition int

	if piece.colour == White {
		shortCastleAvailable = currentGameState.whiteCanKingSideCastle
		longCastleAvailable = currentGameState.whiteCanQueenSideCastle
		friendlyKingPosition = currentGameState.whiteKingPosition

	} else {
		shortCastleAvailable = currentGameState.blackCanKingSideCastle
		longCastleAvailable = currentGameState.blackCanQueenSideCastle
		friendlyKingPosition = currentGameState.blackKingPosition
	}

	if piece == nil {
		return []int{}, []int{}
	}

	/*Check basic moves for non-kings*/
	if piece.variant != King {
		for i := range piece.moves {
			moveDirection = piece.moves[i]
			for moveRange := 1; moveRange <= piece.moveRange; moveRange++ {
				currentSquare = piece.position + (moveDirection * moveRange)

				/*Check if square outside board*/
				if !isSquareInBoard(currentSquare) {
					break
				}

				/*Check if move goes over edge*/
				if hasMoveCrossedEdge(piecePosition, currentSquare, piece.variant) {
					break
				}

				/*Check if piece in square*/
				targetPiece = board[currentSquare].piece
				if targetPiece != nil {

					/*Piece cannot capture*/
					if !piece.movesEqualsAttacks || targetPiece.colour == piece.colour || targetPiece.variant == King {
						break
					}

					/*Add currentSquare to captures*/
					captures = append(captures, currentSquare)

					break
				}

				/*Add currentSquare to moves*/
				moves = append(moves, currentSquare)

			}
		}
	}

	/*Check basic attacks*/
	if !piece.movesEqualsAttacks {
		for i := range piece.attacks {
			attackDirection := piece.attacks[i]
			for attackRange := 1; attackRange <= piece.moveRange; attackRange++ {
				currentSquare = piece.position + (attackDirection * attackRange)

				/*Check if square outside board*/
				if !isSquareInBoard(currentSquare) {
					break
				}

				/*Check if move goes over edge*/
				if hasMoveCrossedEdge(piecePosition, currentSquare, piece.variant) {
					break
				}

				/*Check if piece in square*/
				targetPiece = board[currentSquare].piece
				if targetPiece == nil {
					continue
				}

				/*Piece cannot capture*/
				if targetPiece.colour == piece.colour || targetPiece.variant == King {
					break
				}

				/*Add currentSquare to captures*/
				captures = append(captures, currentSquare)
				break
			}
		}
	}

	/*Pawn checks: double move, en passant, promotion*/
	if piece.variant == Pawn {

		/*Double move*/
		if (piece.colour == Black && piece.position/8 == 1 && board[piece.position+8].piece == nil) || (piece.colour == White && piece.position/8 == 6 && board[piece.position-8].piece == nil) {
			currentSquare = piece.position + (piece.moves[0] * 2)

			targetPiece = board[currentSquare].piece
			/*If no piece then append move*/
			if targetPiece == nil {
				moves = append(moves, currentSquare)
			}
		}

		/*En passant*/
		if enpassantActive {
			if piece.colour == Black && (piece.position+7 == enpassantSquare || piece.position+9 == enpassantSquare) {
				/*Check if move goes over edge*/
				if !hasMoveCrossedEdge(piecePosition, enpassantSquare, Pawn) {
					captures = append(captures, enpassantSquare)
				}
			}
			if piece.colour == White && (piece.position-7 == enpassantSquare || piece.position-9 == enpassantSquare) {
				/*Check if move goes over edge*/
				if !hasMoveCrossedEdge(piecePosition, enpassantSquare, Pawn) {
					captures = append(captures, enpassantSquare)
				}
			}
		}
	}

	/*King Checks: square attacked, castling*/
	if piece.variant == King {
		for _, moveDirection := range piece.moves {
			currentSquare = piece.position + moveDirection

			if !isSquareInBoard(currentSquare) {
				continue
			}

			/*Check if move goes over edge*/
			if hasMoveCrossedEdge(piecePosition, currentSquare, King) {
				continue
			}

			targetPiece = board[currentSquare].piece
			if targetPiece == nil || targetPiece.colour != piece.colour {
				if isSquareUnderAttack(board, currentSquare, piece.colour) {
					continue
				}
				moves = append(moves, currentSquare)
			}
		}

		if shortCastleAvailable {
			if board[piece.position+1].piece == nil && board[piece.position+2].piece == nil {
				if !isSquareUnderAttack(board, piece.position+1, piece.colour) && !isSquareUnderAttack(board, piece.position+2, piece.colour) {
					moves = append(moves, piece.position+2)
				}
			}
		}

		if longCastleAvailable {
			if board[piece.position-1].piece == nil && board[piece.position-2].piece == nil && board[piece.position-3].piece == nil {
				if !isSquareUnderAttack(board, piece.position-1, piece.colour) && !isSquareUnderAttack(board, piece.position-2, piece.colour) && !isSquareUnderAttack(board, piece.position-3, piece.colour) {
					moves = append(moves, piece.position-2)
				}
			}
		}
	}

	/*
		Check for pins or checks
		If single check, must take piece or block, so keep track of squares along check direction
		If double check king must move
	*/
	var blockingSquares = make(map[int]bool)
	var intermediateBlockingSquares []int
	var checkCount int

	var dummyKnight pieceType
	var dummyRook pieceType
	var dummyBishop pieceType

	if piece.variant != King {

		/*Check for checks*/

		/*Check Pawns First*/
		var possibleAttackingPawnPositions [2]int
		if piece.colour == White {
			possibleAttackingPawnPositions = [2]int{-7, -9}
		} else {
			possibleAttackingPawnPositions = [2]int{7, 9}
		}
		for i := range possibleAttackingPawnPositions {
			currentSquare = friendlyKingPosition + possibleAttackingPawnPositions[i]

			/*Check if square outside board*/
			if !isSquareInBoard(currentSquare) {
				continue
			}

			/*Check if move goes over edge*/
			if hasMoveCrossedEdge(friendlyKingPosition, currentSquare, Pawn) {
				continue
			}

			targetPiece = board[currentSquare].piece
			if targetPiece != nil && targetPiece.colour != piece.colour && targetPiece.variant == Pawn {
				blockingSquares[currentSquare] = true
				if enpassantActive {
					blockingSquares[enpassantSquare] = true
				}
				checkCount += 1
			}
		}

		/*Check Knights*/
		dummyKnight = createKnight(friendlyKingPosition, White)
		for i := range dummyKnight.moves {
			currentSquare = dummyKnight.moves[i] + friendlyKingPosition

			/*Check if square outside board*/
			if !isSquareInBoard(currentSquare) {
				continue
			}

			/*Check if move goes over edge*/
			if hasMoveCrossedEdge(friendlyKingPosition, currentSquare, Knight) {
				continue
			}

			targetPiece = board[currentSquare].piece
			if targetPiece != nil && targetPiece.colour != piece.colour && targetPiece.variant == Knight {
				blockingSquares[currentSquare] = true
				checkCount += 1
			}

		}

		/*Check for pins alongside checks*/

		/*Check horizontal and verical*/
		dummyRook = createRook(friendlyKingPosition, White)
		for i := range dummyRook.moves {

			moveDirection = dummyRook.moves[i]
			intermediateBlockingSquares = []int{}

			for moveRange := 1; moveRange <= dummyRook.moveRange; moveRange++ {
				currentSquare = friendlyKingPosition + (moveDirection * moveRange)
				/*Iterate until outside board, or piece that is not current piece is found*/

				/*Check if square outside board*/
				if !isSquareInBoard(currentSquare) {
					break
				}

				/*Check if move goes over edge*/
				if hasMoveCrossedEdge(friendlyKingPosition, currentSquare, Rook) {
					continue
				}

				targetPiece = board[currentSquare].piece

				/*Iterate until */
				if targetPiece == nil {

					intermediateBlockingSquares = append(intermediateBlockingSquares, currentSquare)
					continue

				}

				/*Target Piece is Current Piece*/
				if targetPiece == piece {
					continue
				}

				/* Target Piece is attacking or pinning to king*/
				if (targetPiece.variant == Rook || targetPiece.variant == Queen) && targetPiece.colour != piece.colour {

					intermediateBlockingSquares = append(intermediateBlockingSquares, currentSquare)
					for _, v := range intermediateBlockingSquares {
						blockingSquares[v] = true
					}
					checkCount += 1
				}

				break
			}

		}

		/*Check diagonal*/
		dummyBishop = createBishop(friendlyKingPosition, White)
		for i := range dummyBishop.moves {

			moveDirection = dummyBishop.moves[i]
			intermediateBlockingSquares = []int{}

			for moveRange := 1; moveRange <= dummyBishop.moveRange; moveRange++ {
				currentSquare = friendlyKingPosition + (moveDirection * moveRange)
				/*Iterate until outside board, or piece that is not current piece is found*/

				/*Check if square outside board*/
				if !isSquareInBoard(currentSquare) {
					break
				}

				/*Check if move goes over edge*/
				if hasMoveCrossedEdge(friendlyKingPosition, currentSquare, Bishop) {
					continue
				}

				targetPiece = board[currentSquare].piece

				/*Iterate until */
				if targetPiece == nil {

					intermediateBlockingSquares = append(intermediateBlockingSquares, currentSquare)
					continue

				}

				/*Target Piece is Current Piece*/
				if targetPiece == piece {
					continue
				}

				/* Target Piece is attacking or pinning to king*/
				if (targetPiece.variant == Bishop || targetPiece.variant == Queen) && targetPiece.colour != piece.colour {

					intermediateBlockingSquares = append(intermediateBlockingSquares, currentSquare)
					for _, v := range intermediateBlockingSquares {
						blockingSquares[v] = true
					}
					checkCount += 1
				}

				break
			}

		}
	}

	/*King must move, checkCount only calculated for non-king pieces*/
	if checkCount >= 2 {
		return []int{}, []int{}
	}

	if checkCount == 1 {
		moves = filter(moves, lambdaMapGet(blockingSquares))
		captures = filter(captures, lambdaMapGet(blockingSquares))
	}

	return moves, captures
}

type setup struct {
	colour  pieceColour
	variant pieceVariant
}

var boardSetup = make(map[int]setup)

func createBoard() [64]square {

	boardSetup[0] = setup{Black, Rook}
	boardSetup[1] = setup{Black, Knight}
	boardSetup[2] = setup{Black, Bishop}
	boardSetup[3] = setup{Black, Queen}
	boardSetup[4] = setup{Black, King}
	boardSetup[5] = setup{Black, Bishop}
	boardSetup[6] = setup{Black, Knight}
	boardSetup[7] = setup{Black, Rook}

	for i := 8; i < 16; i++ {
		boardSetup[i] = setup{Black, Pawn}
	}

	for i := 48; i < 56; i++ {
		boardSetup[i] = setup{White, Pawn}
	}

	boardSetup[56] = setup{White, Rook}
	boardSetup[57] = setup{White, Knight}
	boardSetup[58] = setup{White, Bishop}
	boardSetup[59] = setup{White, Queen}
	boardSetup[60] = setup{White, King}
	boardSetup[61] = setup{White, Bishop}
	boardSetup[62] = setup{White, Knight}
	boardSetup[63] = setup{White, Rook}

	var board [64]square
	for i := 0; i < len(board); i++ {
		board[i] = defaultSquare()
	}

	for key, value := range boardSetup {
		board[key].piece = createPiece(key, value.colour, value.variant)
	}

	return board
}

func getValidMovesForPiece(piecePosition int, currentGameState gameState) (moves []int, captures []int) {
	moves, captures = getMovesandCapturesForPiece(currentGameState.board, piecePosition, currentGameState)
	return
}

func boardFromFEN(fen string) gameState {
	var colour pieceColour
	var variant pieceVariant
	var boardIndex = 0
	var board [64]square
	var runeToVariant = make(map[rune]pieceVariant)
	var parseState = 0

	var currentGameState gameState

	runeToVariant['p'] = Pawn
	runeToVariant['n'] = Knight
	runeToVariant['b'] = Bishop
	runeToVariant['r'] = Rook
	runeToVariant['q'] = Queen
	runeToVariant['k'] = King

	runeToVariant['P'] = Pawn
	runeToVariant['N'] = Knight
	runeToVariant['B'] = Bishop
	runeToVariant['R'] = Rook
	runeToVariant['Q'] = Queen
	runeToVariant['K'] = King

	var fileToInt = make(map[rune]int)
	fileToInt['a'] = 0
	fileToInt['b'] = 1
	fileToInt['c'] = 2
	fileToInt['d'] = 3
	fileToInt['e'] = 4
	fileToInt['f'] = 5
	fileToInt['g'] = 6
	fileToInt['h'] = 7

	for _, char := range fen {

		if char == ' ' {
			parseState += 1
			continue
		}

		switch parseState {
		case 0:
			// Parse Board

			if char == '/' {
				continue
			}

			if unicode.IsDigit(char) {
				for i := 0; i < int(char-'0'); i++ {
					board[boardIndex] = defaultSquare()
					boardIndex += 1
				}
				continue
			}

			if unicode.IsUpper(char) {
				colour = White
			} else {
				colour = Black
			}

			variant = runeToVariant[char]

			board[boardIndex] = defaultSquare()
			board[boardIndex].piece = createPiece(boardIndex, colour, variant)

			if variant == King {
				if colour == White {
					currentGameState.whiteKingPosition = boardIndex
				} else {
					currentGameState.blackKingPosition = boardIndex
				}
			}

			boardIndex += 1

		case 1:
			currentGameState.board = board
			// Parse Turn
			if char == 'w' {
				currentGameState.turn = White
			} else {
				currentGameState.turn = Black
			}

		case 2:
			// Parse Castling
			switch char {
			case 'K':
				currentGameState.whiteCanKingSideCastle = true
			case 'Q':
				currentGameState.whiteCanQueenSideCastle = true
			case 'k':
				currentGameState.blackCanKingSideCastle = true
			case 'q':
				currentGameState.blackCanQueenSideCastle = true
			}

		case 3:
			// Parse Enpassant
			if char == '/' {
				continue
			}

			currentGameState.enPassantAvailable = true

			if unicode.IsLetter(char) {
				currentGameState.enPassantTargetSquare += fileToInt[char]
			}

			if unicode.IsDigit(char) {
				currentGameState.enPassantTargetSquare += (8 - int(char-'0')) * 8
			}

		case 4:
			// Parse Halfmove Clock

		case 5:
			// Parse fullmove number
		}

	}

	return currentGameState
}

func IsMoveValid(fen string, piece int, move int) bool {
	var currentGameState = boardFromFEN(fen)
	var moves, captures = getValidMovesForPiece(piece, currentGameState)
	fmt.Printf("Moves for piece: %v\n", append(moves, captures...))

	for _, possibleMove := range append(moves, captures...) {
		if move == possibleMove {
			return true
		}
	}

	return false
}

func chessMain() {

	start := time.Now()

	var rowMap = make(map[int]string)
	rowMap[0] = "8"
	rowMap[1] = "7"
	rowMap[2] = "6"
	rowMap[3] = "5"
	rowMap[4] = "4"
	rowMap[5] = "3"
	rowMap[6] = "2"
	rowMap[7] = "1"
	var colMap = make(map[int]string)
	colMap[0] = "a"
	colMap[1] = "b"
	colMap[2] = "c"
	colMap[3] = "d"
	colMap[4] = "e"
	colMap[5] = "f"
	colMap[6] = "g"
	colMap[7] = "h"

	var moves, captures []int

	// var currentGameState = boardFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	// Enpassant to remove check
	// var currentGameState = boardFromFEN("rn3qnr/pb5p/3bp3/1p1p2k1/1Q3Pp1/2P1P3/PP4PP/RNB1KB1R b KQ f3 0 28")
	// White Castling, black no castling
	// var currentGameState = boardFromFEN("r3k2r/pb1p1ppp/nppb1n2/q3p3/3PP2N/2NBB3/PPPQ1PPP/R3K2R w KQ - 2 17")
	var currentGameState = boardFromFEN("r3k2r/pb1p1ppp/nppb1n2/4p3/3P2qN/2NBB3/PPPQ1PPP/R3K2R b KQ - 3 20")

	fmt.Println(currentGameState)

	var out []string
	var testPosition = 60

	moves, captures = getValidMovesForPiece(testPosition, currentGameState)

	for _, v := range append(moves, captures...) {
		row := v / 8
		col := v % 8

		out = append(out, colMap[col]+rowMap[row])
	}

	fmt.Println(out)

	row := testPosition / 8
	col := testPosition % 8
	fmt.Printf("Moves for piece: %v\n", colMap[col]+rowMap[row])

	for _, v := range out {
		fmt.Println(v)
	}

	// for i := 0; i < 16; i++ {
	// 	row := i / 8
	// 	col := i % 8

	// 	moves, captures = getValidMovesForPiece(i, currentGameState)

	// 	fmt.Printf("Moves for piece: %v\n", colMap[col]+rowMap[row])
	// 	fmt.Printf("Moves:\n")
	// 	for _, v := range moves {
	// 		row := v / 8
	// 		col := v % 8

	// 		fmt.Println(colMap[col] + rowMap[row])
	// 	}
	// 	fmt.Printf("Captures:\n")
	// 	for _, v := range captures {
	// 		row := v / 8
	// 		col := v % 8

	// 		fmt.Println(colMap[col] + rowMap[row])
	// 	}
	// 	fmt.Println()
	// }

	// for i := 63; i > 47; i-- {
	// 	row := i / 8
	// 	col := i % 8

	// 	moves, captures = getValidMovesForPiece(i, currentGameState)

	// 	fmt.Printf("Moves for piece: %v\n", colMap[col]+rowMap[row])
	// 	fmt.Printf("Moves:\n")
	// 	for _, v := range moves {
	// 		row := v / 8
	// 		col := v % 8

	// 		fmt.Println(colMap[col] + rowMap[row])
	// 	}
	// 	fmt.Printf("Captures:\n")
	// 	for _, v := range captures {
	// 		row := v / 8
	// 		col := v % 8

	// 		fmt.Println(colMap[col] + rowMap[row])
	// 	}
	// 	fmt.Println()
	// }

	elapsed := time.Since(start)
	fmt.Printf("Took %s\n", elapsed)
}
