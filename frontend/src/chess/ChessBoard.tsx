import React, { useContext, useEffect, useRef, useState } from "react";
import { PieceColour, PieceVariant } from "./ChessLogic";
import { GameContext, gameContext } from "./GameContext";

const colourToString = new Map<PieceColour, string>()

colourToString.set(PieceColour.White, 'white')
colourToString.set(PieceColour.Black, 'black')

const variantToString = new Map<PieceVariant, string>()

variantToString.set(PieceVariant.Pawn, 'pawn')
variantToString.set(PieceVariant.Knight, 'knight')
variantToString.set(PieceVariant.Bishop, 'bishop')
variantToString.set(PieceVariant.Rook, 'rook')
variantToString.set(PieceVariant.Queen, 'queen')
variantToString.set(PieceVariant.King, 'king')

enum ClickAction {
  clear,
  showMoves,
  makeMove,
  choosePromotion,
}

async function clickHandler(
  event: React.MouseEvent,
  rect: DOMRect | null,
  flip: boolean,
  promotionActive: boolean,
  promotionSquare: number,
  moves: number[],
  setMoves: React.Dispatch<React.SetStateAction<number[]>>,
  captures: number[],
  setCaptures: React.Dispatch<React.SetStateAction<number[]>>,
  selectedPiece: number | null,
  setSelectedPiece: React.Dispatch<React.SetStateAction<number | null>>,
  promotionNextMove: boolean,
  setPromotionNextMove: React.Dispatch<React.SetStateAction<boolean>>,
  setPromotionActive: React.Dispatch<React.SetStateAction<boolean>>,
  setPromotionSquare: React.Dispatch<React.SetStateAction<number>>,
  waiting: boolean,
  setWaiting: React.Dispatch<React.SetStateAction<boolean>>,
) {
  console.log(`${event.clientX}, ${event.clientY}`)
  // Calculate board position
  if (rect === null) {
    throw new Error("Bounding rect for board is not defined")
  }
  const game = useContext(GameContext)
  if (!game) {
    throw new Error("clickHandler must be called from within a GameContext")
  }

  const boardXPosition = Math.floor((event.clientX - rect.left) / (rect.width / 8))
  const boardYPosition = Math.floor((event.clientY - rect.top) / (rect.height / 8))
  let position = boardYPosition * 8 + boardXPosition

  if (flip) {
    position = 63 - position
  }

  // If waiting do nothing
  if (waiting) {
    return
  }

  // Set Waiting
  setWaiting(true)

  // Get clickAction state
  let clickAction = ClickAction.clear
  if (game?.matchData.activeMove != (game as gameContext)?.matchData.stateHistory.length - 1) {
    setWaiting(false)
    return
  } else if (game.matchData.gameOverStatus != 0) {
    console.log("Game has finished")
    setWaiting(false)
    return
  } else if (promotionActive && [0, 8, 16, 24].includes(Math.abs(position - promotionSquare))) {
    clickAction = ClickAction.choosePromotion
  } else if ([...moves, ...captures].includes(position)) {
    clickAction = ClickAction.makeMove
    // @TODO
    // null == null 
  } else if (game?.matchData.activeState.board[position][0] == game?.playerColour && position != selectedPiece) {
    clickAction = ClickAction.showMoves
  }

  // If showing promotion, check promotion selection, if promoting submit move and reset to bare

  // If showing moves, check for move click, if move clicked, submit and reset to bare

  // If clicked on piece, select piece

  // If clicked on board, reset to bare

  switch (clickAction) {
  case ClickAction.clear:
    setToBare(
      setSelectedPiece,
      setMoves,
      setCaptures,
      setPromotionNextMove,
      setPromotionActive,
      setPromotionSquare
    )
    break;

  case ClickAction.makeMove:
  case ClickAction.choosePromotion:
  {
    if (selectedPiece === null) {
      throw new Error("Posting move with no piece")
    }

    // Render promotion component
    if (promotionNextMove) {
      setPromotionActive(true)
      setPromotionSquare(position)
      setPromotionNextMove(false)
      break;
    }

    let promotionString = ""
    if (clickAction == ClickAction.choosePromotion) {
      const promotionIndex = [0, 8, 16, 24].indexOf(Math.abs(position - promotionSquare))
      promotionString = "qnrb"[promotionIndex]
      position = promotionSquare
    }
    // Send current FEN, piece, move, new FEN
    wsPostMove(position, selectedPiece, promotionString)

    // Clear cache, clear moves
    setToBare(
      setSelectedPiece,
      setMoves,
      setCaptures,
      setPromotionNextMove,
      setPromotionActive,
      setPromotionSquare
    )
    break;
  }

  case ClickAction.showMoves:
  // Fetch moves
  {
    const data = await fetchPossibleMoves(position)

    // Set Moves
    setMoves(data["moves"] || [])
    setCaptures(data["captures"] || [])
    setPromotionNextMove(data["triggerPromotion"])

    // Show Moves
    setSelectedPiece(position)
    break;
  }
  }

  setWaiting(false)
}

function setToBare(
  setSelectedPiece: React.Dispatch<React.SetStateAction<number | null>>,
  setMoves: React.Dispatch<React.SetStateAction<number[]>>,
  setCaptures: React.Dispatch<React.SetStateAction<number[]>>,
  setPromotionNextMove: React.Dispatch<React.SetStateAction<boolean>>,
  setPromotionActive: React.Dispatch<React.SetStateAction<boolean>>,
  setPromotionSquare: React.Dispatch<React.SetStateAction<number>>,
) {
  setSelectedPiece(null)
  setMoves([])
  setCaptures([])
  setPromotionNextMove(false)
  setPromotionActive(false)
  setPromotionSquare(0)
}

async function fetchPossibleMoves(position: number) {
  const game = useContext(GameContext)
  if (!game) {
    throw new Error("fetchPossibleMoves must be called from within a GameContext")
  }
  try {
    const mostRecentMove = game.matchData.stateHistory.at(-1)
    if (!mostRecentMove) {
      return
    }
    const response = await fetch(import.meta.env.VITE_API_FETCH_MOVES_URL, {
      "method": "POST",
      "body": JSON.stringify({
        "fen": mostRecentMove["FEN"],
        "piece": position,
      })
    })

    if (!response.ok) {
      throw new Error(`Response status: ${response.status}`)
    }

    const data = await response.json()
    return data
  }

  catch (error: unknown) {
    if (error instanceof Error) {
      console.error(error.message)
    } else {
      console.error(error)
    }
  }

  return {}

}

function wsPostMove(position: number, piece: number, promotion: string) {
  const game = useContext(GameContext)
  if (!game) {
    throw new Error("wsPostMove must be called from within a GameContext")
  }
  game?.webSocket?.send(JSON.stringify({
    "messageType": "postMove",
    "body": {
      "piece": piece,
      "move": position,
      "promotionString": promotion,
    }
  }))
}

function getRowAndColFromBoardIndex(idx: number, flip: boolean): [number, number] {
  const row = Math.floor(idx / 8)
  const col = idx % 8

  // Checks if the board needs to be flipped and transforms the row and col
  if (flip) {
    return [Math.abs(7 - row), Math.abs(7 - col)]
  }
  return [row, col]
}

function PiecesComponent({ flip }: { flip: boolean }) {
  const game = useContext(GameContext)
  if (!game) {
    throw new Error("PiecesComponent must be called from within a GameContext")
  }

  return (
    game.matchData.activeState.board.map((square, idx) => {
      const [colour, variant] = square
      if (colour === null || variant === null) {
        return <React.Fragment key={idx} />
      }
      
      const [row, col] = getRowAndColFromBoardIndex(idx, flip)
      return (
        <div key={idx} className={`${colourToString.get(colour)}-${variantToString.get(variant)}`} style={{ transform: `translate(${col * 50}px, ${row * 50}px)` }} />
      )
    })
  )
}

function MovesComponent({ moves, flip }: { moves: number[], flip: boolean }) {
  return (
    (moves).map((move, idx) => {
      const [row, col] = getRowAndColFromBoardIndex(move, flip)
      return (
        <div key={idx} className='potential-move' style={{ transform: `translate(${col * 50}px, ${row * 50}px)` }} />
      )
    }))
}

function CapturesComponent({ captures, flip }: { captures: number[], flip: boolean }) {
  return (
    captures.map((move, idx) => {
      const [row, col] = getRowAndColFromBoardIndex(move, flip)
      return (
        <div key={idx} className='potential-capture' style={{ transform: `translate(${col * 50}px, ${row * 50}px)` }} />
      )
    }))
}

function LastMoveComponent({ flip }: { flip: boolean }) {
  const game = useContext(GameContext)
  if (!game) {
    throw new Error("LastMoveComponent must be called from within a GameContext")
  }
  return (
    game.matchData.activeState.lastMove.map((move, idx) => {
      if (game.matchData.activeMove == 0) {
        return <React.Fragment key={idx} />
      }
      const [row, col] = getRowAndColFromBoardIndex(move, flip)
      return (
        <div key={idx} className='last-move' style={{ transform: `translate(${col * 50}px, ${row * 50}px)` }} />
      )
    }))
}

function PromotionComponent({ promotionSquare, promotionActive, flip }: { promotionSquare: number, promotionActive: boolean, flip: boolean }) {
  if (!promotionActive) {
    return <></>
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [_row, col] = getRowAndColFromBoardIndex(promotionSquare, flip)

  let promotionColour = colourToString.get(PieceColour.Black)

  if (promotionSquare <= 7) {
    promotionColour = colourToString.get(PieceColour.White)
  }

  return (
    <>
      <div className={`${promotionColour}-queen promotion`} style={{ transform: `translate(${col * 50}px, ${0 * 50}px)` }} />
      <div className={`${promotionColour}-knight promotion`} style={{ transform: `translate(${col * 50}px, ${1 * 50}px)` }} />
      <div className={`${promotionColour}-rook promotion`} style={{ transform: `translate(${col * 50}px, ${2 * 50}px)` }} />
      <div className={`${promotionColour}-bishop promotion`} style={{ transform: `translate(${col * 50}px, ${3 * 50}px)` }} />
    </>
  )
}

function GameOverComponent() {
  const game = useContext(GameContext)
  if (!game) {
    throw new Error("LastMoveComponent must be called from within a GameContext")
  }

  if (game.matchData.gameOverStatus == 0) {
    return <></>
  }

  const gameOverStatusCodes = ["Ongoing", "Stalemate", "Checkmate", "Threefold Repetition", "Insufficient Material", "White Flagged", "Black Flagged", "Draw", "White Resigned", "Black Resigned", "Game Aborted"]
  const gameOverText = gameOverStatusCodes[game?.matchData.gameOverStatus || 0]

  return <div style={{ transform: `translate(${0}px, ${180}px)`, color: "black" }}>{gameOverText}</div>
}

export function ChessBoard() {
  const boardRef = useRef<HTMLDivElement | null>(null)
  const [rect, setRect] = useState<DOMRect | null>(null)
  
  const [waiting, setWaiting] = useState(false)
  
  const [selectedPiece, setSelectedPiece] = useState<number | null>(null)
  const [moves, setMoves] = useState<number[]>([])
  const [captures, setCaptures] = useState<number[]>([])
  
  const [promotionNextMove, setPromotionNextMove] = useState(false)
  const [promotionActive, setPromotionActive] = useState(false)
  const [promotionSquare, setPromotionSquare] = useState(0)
  
  const game = useContext(GameContext)
  if (!game) {
    throw new Error('ChessBoard must be used within a GameContext Provider');
  }

  const [flip, setFlip] = useState(false)

  useEffect(() => {
    setFlip(game.playerColour == PieceColour.Black)
  }, [game.playerColour])
  
  
  useEffect(() => {
    const updateRect = () => {
      if (boardRef.current) {
        const boundingRect = boardRef.current.getBoundingClientRect()
        setRect(boundingRect)
      }
    }
  
    updateRect()
  
    window.addEventListener('resize', updateRect)
  
    return () => {
      window.removeEventListener('resize', updateRect)
    }
  }, [])
  
  // Clear board when changing active move
  useEffect(() => {
    setToBare(
      setSelectedPiece,
      setMoves,
      setCaptures,
      setPromotionNextMove,
      setPromotionActive,
      setPromotionSquare
    )
  }, [game.matchData.activeMove])
  
  return (
    <div className='chessboard' onClick={(event) => clickHandler(
      event,
      rect,
      flip,
      promotionActive,
      promotionSquare,
      moves,
      setMoves,
      captures,
      setCaptures,
      selectedPiece,
      setSelectedPiece,
      promotionNextMove,
      setPromotionNextMove,
      setPromotionActive,
      setPromotionSquare,
      waiting,
      setWaiting
    )} ref={boardRef}>
      <LastMoveComponent flip={flip}/>
      <PiecesComponent flip={flip}/>
      <MovesComponent moves={moves} flip={flip}/>
      <CapturesComponent captures={captures} flip={flip}/>
      <PromotionComponent promotionSquare={promotionSquare} promotionActive={promotionActive} flip={flip}/>
      <GameOverComponent />
    </div>
  )
}