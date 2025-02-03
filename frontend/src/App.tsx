import { useEffect, useRef, useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import { parseGameStateFromFEN, PieceVariant, PieceColour } from './Chess.tsx'

function App() {
  const [count, setCount] = useState(0)
  const [responseText, setResponseText] = useState("")

  const [fen, setFen] = useState("")
  const [piece, setPiece] = useState("")
  const [move, setMove] = useState("")

  console.log(import.meta.env.VITE_API_URL)

  return (
    <>
      <div>
        <a href="https://vite.dev" target="_blank">
          <img src={viteLogo} className="logo" alt="Vite logo" />
        </a>
        <a href="https://react.dev" target="_blank">
          <img src={reactLogo} className="logo react" alt="React logo" />
        </a>
      </div>
      <h1>Vite + React</h1>
      <div className="card">
        <button onClick={() => setCount((count) => count + 1)}>
          count is {count}
        </button>
        <p>
          Edit <code>src/App.tsx</code> and save to test HMR
        </p>
      </div>
      <div>
        <label htmlFor="fen">FEN</label>
        <input type="text" id="fen" name="fen" value={fen} onChange={(event) => setFen(event.target.value)}/>
        <br />

        <label htmlFor="piece">Piece</label>
        <input type="text" id="piece" name="piece" value={piece} onChange={(event) => setPiece(event.target.value)}/>
        <br />

        <label htmlFor="move">Move</label>
        <input type="move" id="move" name="move" value={move} onChange={(event) => setMove(event.target.value)}/>
        <br />

        <input type="submit" onClick={() => {
          setResponseText("Waiting")
          fetch(import.meta.env.VITE_API_URL, {
            "method": "POST",
            "body": JSON.stringify({
              "fen": fen,
              "piece": parseInt(piece),
              "move": parseInt(move),
            })
          }).then(
            (response) => {
              response.text().then(
                (value) => setResponseText(value),
                () => setResponseText("Could Not Read")
              )
            },
            () => setResponseText("Failed")
          )
        }}/>
      </div>
      <div>
        {responseText}
      </div>
      <p className="read-the-docs">
        Click on the Vite and React logos to learn more
      </p>
      <ChessBoard/>
    </>
  )
}

const variantToString = new Map<PieceVariant, string>()

variantToString.set(PieceVariant.Pawn, 'pawn')
variantToString.set(PieceVariant.Knight, 'knight')
variantToString.set(PieceVariant.Bishop, 'bishop')
variantToString.set(PieceVariant.Rook, 'rook')
variantToString.set(PieceVariant.Queen, 'queen')
variantToString.set(PieceVariant.King, 'king')

const colourToString = new Map<PieceColour, string>()

colourToString.set(PieceColour.White, 'white')
colourToString.set(PieceColour.Black, 'black')

function ChessBoard() {
  const boardRef = useRef<HTMLDivElement | null>(null)
  const [rect, setRect] = useState<DOMRect | null>(null)
  const [waiting, setWaiting] = useState(false)
  const [selectedPiece, setSelectedPiece] = useState<number | null>(null)
  const [moves, setMoves] = useState<number[]>([])
  const [captures, setCaptures] = useState<number[]>([])
  const [lastMove, setLastMove] = useState<number[]>([])
  const [promotion, setPromotion] = useState(false)
  const [gameState, setGameState] = useState(parseGameStateFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"))


  // var moves = [4, 12, 20]
  // var captures = [1, 8, 9]
  // var lastMove = [50, 58]

  var moveCache = new Map<number, number[]>()

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

  useEffect(() => {
    console.log(moves)
  }, [moves])

  useEffect(() => {
    console.log(captures)
  }, [captures])

  async function clickHandler(event: React.MouseEvent) {
    if (rect === null) {
      throw new Error("Bounding rect for board is not defined")
    }
    var boardXPosition = Math.floor((event.clientX - rect.left) / (rect.width / 8))
    var boardYPosition = Math.floor((event.clientY - rect.top) / (rect.height / 8))
    var position = boardYPosition * 8 + boardXPosition
    console.log(boardXPosition)
    console.log(boardYPosition)

    // If waiting do nothing
    if (waiting) { 
      return
    }

    // Set Waiting
    setWaiting(true)

    if (selectedPiece === null) {

      if (gameState.board[position][0] == null || gameState.board[position][1] == null) {
        setWaiting(false)
        return
      }

      // Check cache
      if (!moveCache.get(position)) {
        // Fetch moves
        fetch(import.meta.env.VITE_API_FETCH_MOVES_URL, {
          "method": "POST",
          "body": JSON.stringify({
            "fen": gameState.fen,
            "piece": position,
          })}).then((response) => {
            response.json().then((data) => {
              // Add to cache
              setMoves(data["moves"] || [])
              setCaptures(data["captures"] || [])
              setPromotion(data["promotion"])
              moveCache.set(position, data["moves"] || [])
              console.log(captures)
            })
          })
      }
      
      // Set Moves
      // setMoves(data["moves"])

      // Show Moves
      setSelectedPiece(position)

      // Clear Waiting
      setWaiting(false)
    } else if (!moves.includes(position) && !captures.includes(position)) {
      // If not clicked on move
      setSelectedPiece(null)

      // Clear Moves
      setMoves([])
      setCaptures([])
      setPromotion(false)

      // Clear Waiting
      setWaiting(false)
    }

    else if (selectedPiece !== null) {

      // Send current FEN, piece, move, new FEN
      try {
        var newBoard = [...gameState.board]
        newBoard[position] = newBoard[selectedPiece]
        newBoard[selectedPiece] = [null, null]
        var response = await fetch(import.meta.env.VITE_API_MAKE_MOVE_URL, {
          "method": "POST",
          "body": JSON.stringify({
            "currentFEN": gameState.fen,
            "piece": selectedPiece,
            "move": position,
          })})
          if (!response.ok) {
            throw new Error(`Response status: ${response.status}`)
          }

          var data = await response.json()

          // If accepted update board
          if (data["isValid"]) {
            setGameState(parseGameStateFromFEN(data["newFEN"]))
            setLastMove(data["lastMove"])
          } 

          // Clear cache, clear moves
          moveCache.clear()
          setMoves([])
          setCaptures([])
          setPromotion(false)

      } catch (error: any) {
        console.error(error.message)
      }
      setSelectedPiece(null)

    }

    // Clear Waiting
    setWaiting(false)
  }

  const PiecesComponent = gameState.board.map((square, idx) => {
    const [colour, variant] = square
    if (colour === null || variant === null) {
      return
    }

    var row = Math.floor(idx / 8)
    var col = idx % 8
    return (
      <div className={`${colourToString.get(colour)}-${variantToString.get(variant)}`} style={{transform: `translate(${col * 50}px, ${row * 50}px)`}}/>
    )})

  const MovesComponent = (moves).map(move => {
    var row = Math.floor(move / 8)
    var col = move % 8
    return (
    <div className='potential-move' style={{transform: `translate(${col * 50}px, ${row * 50}px)`}}/>
  )})

  const CapturesComponent = captures.map(move => {
    var row = Math.floor(move / 8)
    var col = move % 8
    return (
    <div className='potential-capture' style={{transform: `translate(${col * 50}px, ${row * 50}px)`}}/>
  )})

  const LastMoveComponent = lastMove.map(move => {
    var row = Math.floor(move / 8)
    var col = move % 8
    return (
    <div className='last-move' style={{transform: `translate(${col * 50}px, ${row * 50}px)`}}/>
  )})

  const PromotionComponent = (promotionSquare: number) => {
    return (
      <div></div>
    )
  }

  return (
    <div className='chessboard' onClick={clickHandler} ref={boardRef}>
      {LastMoveComponent}
      {PiecesComponent}
      {MovesComponent}
      {CapturesComponent}
    </div>
  )
}

export default App
