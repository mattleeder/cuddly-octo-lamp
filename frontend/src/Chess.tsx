export enum PieceColour {
    White,
    Black,
}

export enum PieceVariant {
    Pawn,
    Knight,
    Bishop,
    Rook,
    Queen,
    King,
}

const charToPiece = new Map<string, [PieceColour, PieceVariant]>()

// White
charToPiece.set('P', [PieceColour.White, PieceVariant.Pawn])
charToPiece.set('N', [PieceColour.White, PieceVariant.Knight])
charToPiece.set('B', [PieceColour.White, PieceVariant.Bishop])
charToPiece.set('R', [PieceColour.White, PieceVariant.Rook])
charToPiece.set('Q', [PieceColour.White, PieceVariant.Queen])
charToPiece.set('K', [PieceColour.White, PieceVariant.King])

// Black
charToPiece.set('p', [PieceColour.Black, PieceVariant.Pawn])
charToPiece.set('n', [PieceColour.Black, PieceVariant.Knight])
charToPiece.set('b', [PieceColour.Black, PieceVariant.Bishop])
charToPiece.set('r', [PieceColour.Black, PieceVariant.Rook])
charToPiece.set('q', [PieceColour.Black, PieceVariant.Queen])
charToPiece.set('k', [PieceColour.Black, PieceVariant.King])

const variantToChar = new Map<PieceVariant, string>()

variantToChar.set(PieceVariant.Pawn, 'p')
variantToChar.set(PieceVariant.Knight, 'n')
variantToChar.set(PieceVariant.Bishop, 'b')
variantToChar.set(PieceVariant.Rook, 'r')
variantToChar.set(PieceVariant.Queen, 'q')
variantToChar.set(PieceVariant.King, 'k')


interface pieceType {
    position: number
    colour: PieceColour
    variant: PieceVariant
    moves: number[]
    attacks: number[]
    moveRange: number
    attackRange: number
}

function Chess() {

}


export function parseBoardFromFEN(fen: string): [PieceColour | null, PieceVariant | null][] {
    // Board, turn, castling, enpassant, halfmove, fullmove
    var args = fen.split(" ")
    var pieces: [PieceColour | null, PieceVariant | null][] = []

    for (let char of args[0]) {
        if (char == "/") {
            continue
        }

        // Digit
        if (char >= '1' && char <= '8') {
            for (var i = 0; i < parseInt(char); i++) {
                pieces.push([null, null])
            }
            continue
        }
        
        var piece = charToPiece.get(char)
        if (piece) {
            pieces.push(piece)
        }

    }

    return pieces
}

export function boardToFEN(board: [PieceColour | null, PieceVariant | null][]): string {
    var fen: string[] = []

    var rowCount = 0
    var emptyCount = 0

    for (let [colour, variant] of board) {
        
        rowCount += 1

        if (variant === null || colour === null) {
            emptyCount += 1
        } else {
            
            if (emptyCount > 0) {
                fen.push(emptyCount.toString())
                emptyCount = 0
            }
            
            var char = variantToChar.get(variant)
            
            if (char == undefined) {
                throw new Error("Unable to convert variant to char")
            }
            
            if (colour == PieceColour.White) {
                char = char?.toUpperCase()
            }
            
            fen.push(char)
        }

        if (rowCount >= 8) {
            if (emptyCount > 0) {
                fen.push(emptyCount.toString())
                emptyCount = 0
            }
            fen.push("/")
            rowCount = 0
        }
    }

    fen.push(" w KQkq - 0 1")

    return fen.join("")
}