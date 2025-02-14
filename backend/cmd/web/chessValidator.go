package main

import (
	"fmt"
	"slices"
	"strconv"
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

type gameOverStatusCode int

const (
	Ongoing = iota
	Stalemate
	Checkmate
	ThreefoldRepetition
	InsufficientMaterial
)

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

func hasQueenCrossedEdgeThroughDiagonal(direction int, piecePosition int, movePosition int) bool {
	// Function is used to prevent Queen jumping 7 squares left or right
	// Not diagonal
	if !(abs(direction) == 7 || abs(direction) == 9) {
		return false
	}

	var pieceRow = getRow(piecePosition)
	var pieceCol = getCol(piecePosition)
	var moveRow = getRow(movePosition)
	var moveCol = getCol(movePosition)
	var rowChange = abs(pieceRow - moveRow)
	var colChange = abs(pieceCol - moveCol)

	return rowChange != colChange
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
			if targetPiece == nil || (targetPiece.variant == King && targetPiece.colour == defendingColour) {
				continue
			}
			if (targetPiece.variant == Rook || targetPiece.variant == Queen || (targetPiece.variant == King && attackRange == 1)) && targetPiece.colour != defendingColour {
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
			if targetPiece == nil || (targetPiece.variant == King && targetPiece.colour == defendingColour) {
				continue
			}
			if (targetPiece.variant == Bishop || targetPiece.variant == Queen || (targetPiece.variant == King && attackRange == 1)) && targetPiece.colour != defendingColour {
				return true
			}
			break
		}
	}

	return false
}

func squareOnEdgeOfBoard(square int) bool {
	return square <= 7 || square >= 56 || square%8 == 0 || square%7 == 0
}

func getMovesandCapturesForPiece(piecePosition int, currentGameState gameState) (moves []int, captures []int, triggerPromotion bool, friendlyKingInCheck bool) {
	/*
		Pawns must check for: promotion, double move, en passant.
		Kings must check for: castling, square attacked
		Promotions must check for pins?

		All piece must check for: checks, blocks, pins
	*/

	var currentSquare int
	var targetPiece *pieceType
	var moveDirection int
	var board = currentGameState.board
	var enpassantActive = currentGameState.enPassantAvailable
	var enpassantSquare = currentGameState.enPassantTargetSquare
	var piece = board[piecePosition].piece
	triggerPromotion = false

	if piece == nil {
		return []int{}, []int{}, triggerPromotion, false
	}

	// Not your turn
	if piece.colour != currentGameState.turn {
		return []int{}, []int{}, triggerPromotion, false
	}

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

				// For queen, to stop jumping 7 squares left / right
				if piece.variant == Queen && hasQueenCrossedEdgeThroughDiagonal(moveDirection, piecePosition, currentSquare) {
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

		if (8 <= piecePosition && piecePosition <= 15 && piece.colour == White) || (48 <= piecePosition && piecePosition <= 55 && piece.colour == Black) {
			triggerPromotion = true
		}

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

	// if piece.variant != King {
	if true {

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
				if (targetPiece.variant == Rook || targetPiece.variant == Queen || (targetPiece.variant == King && moveRange == 1)) && targetPiece.colour != piece.colour {

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
				if (targetPiece.variant == Bishop || targetPiece.variant == Queen || (targetPiece.variant == King && moveRange == 1)) && targetPiece.colour != piece.colour {

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
				if targetPiece != nil {
					captures = append(captures, currentSquare)
				} else {
					moves = append(moves, currentSquare)
				}
			}
		}

		if shortCastleAvailable && checkCount == 0 {
			if board[piece.position+1].piece == nil && board[piece.position+2].piece == nil {
				if !isSquareUnderAttack(board, piece.position+1, piece.colour) && !isSquareUnderAttack(board, piece.position+2, piece.colour) {
					moves = append(moves, piece.position+2)
				}
			}
		}

		if longCastleAvailable && checkCount == 0 {
			if board[piece.position-1].piece == nil && board[piece.position-2].piece == nil && board[piece.position-3].piece == nil {
				if !isSquareUnderAttack(board, piece.position-1, piece.colour) && !isSquareUnderAttack(board, piece.position-2, piece.colour) && !isSquareUnderAttack(board, piece.position-3, piece.colour) {
					moves = append(moves, piece.position-2)
				}
			}
		}
	}

	/*King must move, checkCount only calculated for non-king pieces*/
	if checkCount >= 2 {
		return []int{}, []int{}, false, true
	}

	if checkCount == 1 && piece.variant != King {
		moves = filter(moves, lambdaMapGet(blockingSquares))
		captures = filter(captures, lambdaMapGet(blockingSquares))
	}

	return moves, captures, triggerPromotion, checkCount > 0
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

func getValidMovesForPiece(piecePosition int, currentGameState gameState) (moves []int, captures []int, triggerPromotion bool, friendlyKingInCheck bool) {
	moves, captures, triggerPromotion, friendlyKingInCheck = getMovesandCapturesForPiece(piecePosition, currentGameState)
	fmt.Printf("Moves for piece: %v\n", append(moves, captures...))
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
			currentGameState.halfMoveClock *= 10
			val, err := strconv.Atoi(string(char))
			if err != nil {
				app.errorLog.Println(err)
			}
			currentGameState.halfMoveClock += val

		case 5:
			// Parse fullmove number
			currentGameState.fullMoveNumber *= 10
			val, err := strconv.Atoi(string(char))
			if err != nil {
				app.errorLog.Println(err)
			}
			currentGameState.fullMoveNumber += val
		}

	}

	return currentGameState
}

func IsMoveValid(fen string, piece int, move int) bool {
	var currentGameState = boardFromFEN(fen)
	var moves, captures, _, _ = getValidMovesForPiece(piece, currentGameState)
	fmt.Printf("Moves for piece: %v\n", append(moves, captures...))

	for _, possibleMove := range append(moves, captures...) {
		if move == possibleMove {
			return true
		}
	}

	return false
}

func canColourMove(currentGameState gameState, colour pieceColour) bool {
	var piece *pieceType
	for i := range currentGameState.board {
		piece = currentGameState.board[i].piece
		if piece != nil && piece.colour == colour {
			var moves, captures, _, _ = getValidMovesForPiece(i, currentGameState)
			if len(moves) > 0 || len(captures) > 0 {
				return true
			}
		}
	}

	return false
}

func gameHasSufficientMaterial(currentGameState gameState) bool {
	// Insufficient Scenarios
	// K v K
	// K & B vs K
	// K & N vs K
	// K & B vs K & B of same colour
	var piece *pieceType
	var freqDict = make(map[pieceVariant]int)
	var totalPieceCount int
	var darkSquareBishopPresent = false
	var lightSquareBishopPresent = false
	for i := range currentGameState.board {
		piece = currentGameState.board[i].piece
		if piece != nil {
			freqDict[piece.variant] += 1
			totalPieceCount += 1
			if piece.variant == Bishop {
				row := getRow(i)
				col := getCol(i)
				if (row+col)%2 == 0 {
					lightSquareBishopPresent = true
				} else {
					darkSquareBishopPresent = true
				}
			}
		}
	}

	if totalPieceCount == 2 {
		return false
	}

	if totalPieceCount == 3 && (freqDict[Knight] >= 1 || freqDict[Bishop] >= 1) {
		return false
	}

	if totalPieceCount-2 == freqDict[Bishop] && !(darkSquareBishopPresent && lightSquareBishopPresent) {
		return false
	}

	return true
}

func intToAlgebraicNotation(position int) string {
	var columns = []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	row := 8 - position/8
	col := position % 8

	return fmt.Sprintf("%s%v", columns[col], row)
}

func getFENAfterMove(currentFEN string, piece int, move int, promotionString string) (string, gameOverStatusCode, string) {
	var currentGameState = boardFromFEN(currentFEN)

	// Algebraic Notation
	// Add the piece type, if pawn add nothing but store the file
	// Are there any other pieces of the same type and colour that can move to this square?
	// If so add file and/or rank to distinguish
	// If capture add an "x", prepend pawn file if it is stored
	// Add the move
	// If promotion add an equals and then promotion

	var variantToString = make(map[pieceVariant]string)
	variantToString[Knight] = "n"
	variantToString[Bishop] = "b"
	variantToString[Rook] = "r"
	variantToString[Queen] = "q"

	var algebraicNotation = ""
	var pawnFile = ""
	var oldPositionAlgebraic = intToAlgebraicNotation(piece)
	if currentGameState.board[piece].piece.variant == Pawn {
		pawnFile = string(oldPositionAlgebraic[0])
	} else {
		algebraicNotation += variantToString[currentGameState.board[piece].piece.variant]
	}
	// Can other piece get there?
	var currentPieceColour = currentGameState.board[piece].piece.colour
	var currentPieceVariant = currentGameState.board[piece].piece.variant
	var rankDistinguishNeeded = false
	var fileDistinguishNeeded = false

	for i := 0; i < 64; i++ {
		if i == piece {
			continue
		}
		if currentGameState.board[i].piece == nil || currentGameState.board[i].piece.colour != currentPieceColour || currentGameState.board[i].piece.variant != currentPieceVariant {
			continue
		}

		moves, captures, _, _ := getMovesandCapturesForPiece(i, currentGameState)

		if slices.Contains(append(moves, captures...), move) {
			// Check if same file
			if (i-piece)%8 == 0 {
				rankDistinguishNeeded = true
			} else {
				fileDistinguishNeeded = true
			}
		}
	}

	if fileDistinguishNeeded {
		algebraicNotation += string(oldPositionAlgebraic[0])
	}

	if rankDistinguishNeeded {
		algebraicNotation += string(oldPositionAlgebraic[1])
	}

	// Capture? Check for piece present or pawn horizontal (to get enpassant)
	if currentGameState.board[move].piece != nil || (pawnFile != "" && (move-piece)%8 != 0) {
		algebraicNotation = pawnFile + algebraicNotation + "x"
	}
	algebraicNotation += intToAlgebraicNotation(move)
	// Promotion added later

	currentGameState.board[move].piece = currentGameState.board[piece].piece
	currentGameState.board[piece].piece = nil

	var newGameState = currentGameState
	newGameState.halfMoveClock += 1
	if newGameState.halfMoveClock%2 == 0 {
		newGameState.fullMoveNumber += 1
	}

	var runeToVariant = make(map[string]pieceVariant)
	runeToVariant["n"] = Knight
	runeToVariant["b"] = Bishop
	runeToVariant["r"] = Rook
	runeToVariant["q"] = Queen

	// Check for promotion
	if newGameState.board[move].piece.variant == Pawn && (move <= 7 || move >= 56) {
		var promotionColour pieceColour = Black
		if move <= 7 {
			promotionColour = White
		}
		promotionVariant, ok := runeToVariant[promotionString]
		// @TODO
		// Handle error properly
		if !ok {
			app.errorLog.Println("Could not understand promotion string")
		}

		newGameState.board[move].piece = createPiece(move, promotionColour, promotionVariant)

		algebraicNotation += "=" + promotionString
	}

	// Check for king move
	if newGameState.board[move].piece.variant == King {
		// Check for castling
		if abs(move-piece) == 2 {
			if newGameState.board[move+1].piece != nil && newGameState.board[move+1].piece.variant == Rook {
				newGameState.board[move-1].piece = newGameState.board[move+1].piece
				newGameState.board[move+1].piece = nil
				algebraicNotation = "O-O"
			} else if newGameState.board[move-2].piece != nil && newGameState.board[move-2].piece.variant == Rook {
				newGameState.board[move+1].piece = newGameState.board[move-2].piece
				newGameState.board[move-2].piece = nil
				algebraicNotation = "O-O-O"
			}
		}
		if newGameState.turn == White {
			newGameState.whiteCanKingSideCastle = false
			newGameState.whiteCanQueenSideCastle = false
		} else {
			newGameState.blackCanKingSideCastle = false
			newGameState.blackCanQueenSideCastle = false
		}
	}

	// Check for rook moves or captures
	if move == 0 || piece == 0 {
		newGameState.blackCanQueenSideCastle = false
	}

	if move == 7 || piece == 7 {
		newGameState.blackCanKingSideCastle = false
	}

	if move == 56 || piece == 56 {
		newGameState.whiteCanQueenSideCastle = false
	}

	if move == 63 || piece == 63 {
		newGameState.whiteCanKingSideCastle = false
	}

	// Check for enpassant capture
	if move == newGameState.enPassantTargetSquare && newGameState.board[move].piece.variant == Pawn && abs(move-piece) != 8 {
		if newGameState.board[move].piece.colour == White {
			newGameState.board[move+8].piece = nil
		} else {
			newGameState.board[move-8].piece = nil
		}
	}

	// Check for enpassant available
	if newGameState.board[move].piece.variant == Pawn && abs(move-piece) == 16 {
		newGameState.enPassantAvailable = true
		if newGameState.turn == White {
			newGameState.enPassantTargetSquare = move + 8
		} else {
			newGameState.enPassantTargetSquare = move - 8
		}
	} else {
		newGameState.enPassantAvailable = false
	}

	// Change Turn
	if newGameState.turn == White {
		newGameState.turn = Black
	} else {
		newGameState.turn = White
	}

	newFEN := gameStateToFEN(newGameState)

	sufficientMaterial := gameHasSufficientMaterial(currentGameState)

	if !sufficientMaterial {
		return newFEN, InsufficientMaterial, algebraicNotation
	}

	// Check for checkmate / stalemate
	var gameOverStatus gameOverStatusCode = Ongoing
	var enemyKing = newGameState.blackKingPosition
	var enemyKingColour = Black
	if currentGameState.board[move].piece.colour == Black {
		enemyKing = newGameState.whiteKingPosition
		enemyKingColour = White
	}

	var moves, captures, _, enemyKingInCheck = getValidMovesForPiece(enemyKing, newGameState)
	var enemyKingMoves = append(moves, captures...)

	// If king has no moves, check if colour can move
	if len(enemyKingMoves) == 0 {
		if !canColourMove(newGameState, pieceColour(enemyKingColour)) {
			if enemyKingInCheck {
				gameOverStatus = Checkmate
			} else {
				gameOverStatus = Stalemate
			}
		}
	}

	if gameOverStatus == Checkmate {
		algebraicNotation += "#"
	} else if enemyKingInCheck {
		algebraicNotation += "+"
	}

	return newFEN, gameOverStatus, algebraicNotation
}

func gameStateToFEN(newGameState gameState) string {
	var newFEN []rune

	var rowCount int
	var emptyCount int
	var colour pieceColour
	var variant pieceVariant

	var variantToRune = make(map[pieceVariant]rune)

	variantToRune[Pawn] = 'p'
	variantToRune[Knight] = 'n'
	variantToRune[Bishop] = 'b'
	variantToRune[Rook] = 'r'
	variantToRune[Queen] = 'q'
	variantToRune[King] = 'k'

	var intToFile = make(map[int]rune)

	intToFile[0] = 'a'
	intToFile[1] = 'b'
	intToFile[2] = 'c'
	intToFile[3] = 'd'
	intToFile[4] = 'e'
	intToFile[5] = 'f'
	intToFile[6] = 'g'
	intToFile[7] = 'h'

	var intToRune = make(map[int]rune)

	intToRune[1] = '1'
	intToRune[2] = '2'
	intToRune[3] = '3'
	intToRune[4] = '4'
	intToRune[5] = '5'
	intToRune[6] = '6'
	intToRune[7] = '7'
	intToRune[8] = '8'

	var char rune

	for _, value := range newGameState.board {
		rowCount += 1

		if value.piece == nil {
			emptyCount += 1
		} else {
			colour = value.piece.colour
			variant = value.piece.variant

			if emptyCount > 0 {
				newFEN = append(newFEN, intToRune[emptyCount])
				emptyCount = 0
			}

			char = variantToRune[variant]

			if colour == White {
				char = unicode.To(0, char)
			}

			newFEN = append(newFEN, char)
		}

		if rowCount >= 8 {
			if emptyCount > 0 {
				newFEN = append(newFEN, intToRune[emptyCount])
				emptyCount = 0
			}
			newFEN = append(newFEN, '/')
			rowCount = 0
		}

	}

	newFEN = newFEN[:len(newFEN)-1]
	newFEN = append(newFEN, ' ')

	// Turn
	if newGameState.turn == White {
		newFEN = append(newFEN, 'w')
	} else {
		newFEN = append(newFEN, 'b')
	}

	newFEN = append(newFEN, ' ')

	// Castling
	if newGameState.whiteCanKingSideCastle {
		newFEN = append(newFEN, 'K')
	}

	if newGameState.whiteCanQueenSideCastle {
		newFEN = append(newFEN, 'Q')
	}

	if newGameState.blackCanKingSideCastle {
		newFEN = append(newFEN, 'k')
	}

	if newGameState.blackCanQueenSideCastle {
		newFEN = append(newFEN, 'q')
	}

	// If no castling info added
	if newFEN[len(newFEN)-1] == ' ' {
		newFEN = append(newFEN, '-')
	}

	newFEN = append(newFEN, ' ')

	// En passant
	if newGameState.enPassantAvailable {
		file := newGameState.enPassantTargetSquare % 8
		rank := 8 - (newGameState.enPassantTargetSquare / 8)

		newFEN = append(newFEN, intToFile[file])
		newFEN = append(newFEN, intToRune[rank])
	} else {
		newFEN = append(newFEN, '-')
	}

	newFEN = append(newFEN, ' ')

	newFEN = append(newFEN, []rune(fmt.Sprint(newGameState.halfMoveClock))...)
	newFEN = append(newFEN, ' ')
	newFEN = append(newFEN, []rune(fmt.Sprint(newGameState.fullMoveNumber))...)

	return string(newFEN)
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

	moves, captures, _, _ = getValidMovesForPiece(testPosition, currentGameState)

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
