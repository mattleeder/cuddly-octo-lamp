import { useEffect, useRef, useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import { parseGameStateFromFEN, PieceVariant, PieceColour } from './Chess.tsx'
import {
  BrowserRouter as Router,
  Routes,
  Route,
  useParams,
  redirect,
  useNavigate,
} from "react-router-dom";

function App() {
  console.log(import.meta.env.VITE_API_URL)

  return (
    <Router>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/matchroom/:matchid" element={<ChessBoard />} />
      </Routes>
    </Router>
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

enum ClickAction {
  clear,
  showMoves,
  makeMove,
  choosePromotion,
}

function Home() {
  const [count, setCount] = useState(0)
  const [responseText, setResponseText] = useState("")

  const [fen, setFen] = useState("")
  const [piece, setPiece] = useState("")
  const [move, setMove] = useState("")
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
    <ChessBoard />
    <JoinQueue />
  </>
  )
}

function JoinQueue() {
  const [inQueue, setInQueue] = useState(false)
  const [waiting, setWaiting] = useState(false)
  const [eventSource, setEventSource] = useState<EventSource | null>(null)
  const navigate = useNavigate()

  useEffect(() => {
    return () => {
      leaveQueue()
    }
  }, [])

  enum ClickAction {
    leaveQueue,
    joinQueue,
  }

  async function joinQueue() {

    try {
      var response = await fetch(import.meta.env.VITE_API_JOIN_QUEUE_URL, {
        method: "POST",
        credentials: 'include',
        body: JSON.stringify({
          "time": 3,
          "increment": 0,
          "action": "join",
        })
      })

      if (!response.ok) {
        console.error(response.status)
      }

      // Joined, start listening for events
      var eventSource = new EventSource(import.meta.env.VITE_API_MATCH_LISTEN_URL, {
        withCredentials: true,
      })
      eventSource.onmessage = (event) => {
        console.log(`message: ${event.data}`)
        navigate("matchroom/" + event.data)
      }
      setEventSource(eventSource)
      return true

    } catch (error) {
      console.error(error)
    }

    // Did not join
    return false
  }

  async function leaveQueue() {

    try {
      var response = await fetch(import.meta.env.VITE_API_JOIN_QUEUE_URL, {
        method: "POST",
        credentials: 'include',
        body: JSON.stringify({
          "action": "leave"
        })
      })

      if (!response.ok) {
        console.error(response.status)
      }

      // Left
      eventSource?.close()
      return true

    } catch (error) {
      console.error(error)
    }

    // Did not leave
    return false
  }

  async function toggleQueue() {
    if (waiting) {
      return
    }

    setWaiting(true)
    var clickAction = inQueue ? ClickAction.leaveQueue : ClickAction.joinQueue

    if (clickAction == ClickAction.leaveQueue) {
      var result = await leaveQueue()
      setInQueue(!result)
    } else {
      var result = await joinQueue()
      setInQueue(result)
    }

    setWaiting(false)
  }

  var buttonText = inQueue ? "In Queue" : "Join Queue"
  return (
    <button onClick={toggleQueue}>{buttonText}</button>
  )
}

function ChessBoard() {
  const boardRef = useRef<HTMLDivElement | null>(null)
  const [rect, setRect] = useState<DOMRect | null>(null)
  const [waiting, setWaiting] = useState(false)
  const [selectedPiece, setSelectedPiece] = useState<number | null>(null)
  const [moves, setMoves] = useState<number[]>([])
  const [captures, setCaptures] = useState<number[]>([])
  const [lastMove, setLastMove] = useState<number[]>([])
  const [promotionNextMove, setPromotionNextMove] = useState(false)
  const [promotionActive, setPromotionActive] = useState(false)
  const [promotionSquare, setPromotionSquare] = useState(0)
  const [gameState, setGameState] = useState(parseGameStateFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"))
  const [gameOverStatus, setGameOverStatus] = useState(0)

  const { matchid } = useParams()
  const [webSocket, setWebSocket] = useState<WebSocket | null>(null)
  const [matchState, setMatchState] = useState(null)
  const [playerCode, setPlayerCode] = useState(2)


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
    // Connect to websocket for matchroom
    var ws = new WebSocket(import.meta.env.VITE_API_MATCHROOM_URL + matchid + '/ws')
    ws.onmessage = (event) => readMessage(event.data)
    setWebSocket(ws)
    return () => {
      // Disconnect
      ws?.close()
    }
  }, [])

  function readMessage(message: any) {
    console.log("FROM WEBSOCKET")
    console.log(message)
    for (let msg of message.split("\n")) {
      var parsedMsg = JSON.parse(msg)
      if (parsedMsg.hasOwnProperty("playerCode")) {
        setPlayerCode(parsedMsg["playerCode"])
      } else if (parsedMsg.hasOwnProperty("newFEN")) {
        setGameState(parseGameStateFromFEN(parsedMsg["newFEN"]))
        setLastMove(parsedMsg("lastMove"))
        setGameOverStatus(parsedMsg("gameOverStatus"))
      }
    }
  }

  function wsPostMove(position: number, piece: number, promotion: string) {
    webSocket?.send(JSON.stringify({
      "piece": piece,
      "move": position,
      "promotionString": promotion,
    }))
  }

  async function fetchPossibleMoves(position: number) {
    try {
      var response = await fetch(import.meta.env.VITE_API_FETCH_MOVES_URL, {
        "method": "POST",
        "body": JSON.stringify({
          "fen": gameState.fen,
          "piece": position,
      })})

      if (!response.ok) {
        throw new Error(`Response status: ${response.status}`)
      }

      var data = await response.json()
      return data
    }

    catch (error: any) {
      console.error(error.message)
    }

    return {}

  }

  async function postMove(position: number, piece: number, promotion: string) {
    wsPostMove(position, piece, promotion)
    try {
      var newBoard = [...gameState.board]
      newBoard[position] = newBoard[piece]
      newBoard[piece] = [null, null]

      var response = await fetch(import.meta.env.VITE_API_MAKE_MOVE_URL, {
        "method": "POST",
        "body": JSON.stringify({
          "currentFEN": gameState.fen,
          "piece": selectedPiece,
          "move": position,
          "promotionString": promotion,
        })})
        if (!response.ok) {
          throw new Error(`Response status: ${response.status}`)
        }

        var data = await response.json()

        return data

    } catch (error: any) {
      console.error(error.message)
    }
  }

  function setToBare() {
    setSelectedPiece(null)
    setMoves([])
    setCaptures([])
    setPromotionNextMove(false)
    setPromotionActive(false)
    setPromotionSquare(0)
  }

  async function clickHandler(event: React.MouseEvent) {
    // Calculate board position
    if (rect === null) {
      throw new Error("Bounding rect for board is not defined")
    }
    var boardXPosition = Math.floor((event.clientX - rect.left) / (rect.width / 8))
    var boardYPosition = Math.floor((event.clientY - rect.top) / (rect.height / 8))
    var position = boardYPosition * 8 + boardXPosition

    // If waiting do nothing
    if (waiting) { 
      return
    }

    // Set Waiting
    setWaiting(true)

    // Get clickAction state
    var clickAction = ClickAction.clear
    if (promotionActive && [0, 8, 16, 24].includes(Math.abs(position - promotionSquare))) {
      clickAction = ClickAction.choosePromotion
    } else if ([...moves, ...captures].includes(position)) {
      clickAction = ClickAction.makeMove
    }else if (gameState.board[position][0] != null && position != selectedPiece) {
      clickAction = ClickAction.showMoves
    }

    console.log(`Click Action: ${clickAction}`)

    // If showing promotion, check promotion selection, if promoting submit move and reset to bare

    // If showing moves, check for move click, if move clicked, submit and reset to bare

    // If clicked on piece, select piece

    // If clicked on board, reset to bare

    switch (clickAction) {
      case ClickAction.clear:
        setToBare()
        break;

      case ClickAction.makeMove:
      case ClickAction.choosePromotion:
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

        var promotionString = ""
        if (clickAction == ClickAction.choosePromotion) {
          var promotionIndex = [0, 8, 16, 24].indexOf(Math.abs(position - promotionSquare))
          console.log(promotionIndex)
          promotionString = "qnrb"[promotionIndex]
          position = promotionSquare
        }
        // Send current FEN, piece, move, new FEN
        console.log("Post move")
        var data = await postMove(position, selectedPiece, promotionString)
        console.log(data)

        // If accepted update board
        // if (data["isValid"]) {
        //   setGameState(parseGameStateFromFEN(data["newFEN"]))
        //   setLastMove(data["lastMove"])
        //   setGameOverStatus(data["gameOverStatus"])
        // } 

        // Clear cache, clear moves
        setToBare()
        break;

      case ClickAction.showMoves:
        // Fetch moves
        data = await fetchPossibleMoves(position)
        
        // Set Moves
        setMoves(data["moves"] || [])
        setCaptures(data["captures"] || [])
        setPromotionNextMove(data["triggerPromotion"])

        // Show Moves
        setSelectedPiece(position)
        break;
    }
    
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

  const PromotionComponent = ({ promotionSquare }: { promotionSquare: number }) => {
    if (!promotionActive) {
      return <></>
    }

    var row = Math.floor(promotionSquare / 8)
    var col = promotionSquare % 8

    var promotionColour = colourToString.get(PieceColour.Black)
    var promotionDirection = -1
    var verticalOffset = 350

    if (promotionSquare <= 7) {
      promotionColour = colourToString.get(PieceColour.White)
      promotionDirection = 1
      verticalOffset = 0
    }

    return (
      <>
      <div className={`${promotionColour}-queen promotion`} style={{transform: `translate(${col * 50}px, ${verticalOffset}px)`}}/>
      <div className={`${promotionColour}-knight promotion`} style={{transform: `translate(${col * 50}px, ${verticalOffset + promotionDirection * 50}px)`}}/>
      <div className={`${promotionColour}-rook promotion`} style={{transform: `translate(${col * 50}px, ${verticalOffset + promotionDirection * 50 * 2}px)`}}/>
      <div className={`${promotionColour}-bishop promotion`} style={{transform: `translate(${col * 50}px, ${verticalOffset + promotionDirection * 50 * 3}px)`}}/>
      </>
    )
  }

  const GameOverComponent = () => {
    if (gameOverStatus == 0) {
      return <></>
    }

    var gameOverStatusCodes = ["Ongoing", "Stalemate", "Checkmate", "Threefold Repetition", "Insufficient Material"]
    var gameOverText = gameOverStatusCodes[gameOverStatus]

    return <div style={{transform: `translate(${0}px, ${180}px)`, color: "black"}}>{gameOverText}</div>
  }

  return (
    <div className='chessboard' onClick={clickHandler} ref={boardRef}>
      {LastMoveComponent}
      {PiecesComponent}
      {MovesComponent}
      {CapturesComponent}
      <PromotionComponent promotionSquare={promotionSquare}/>
      <GameOverComponent />
    </div>
  )
}

interface matchState {
  
}

function MatchRoom() {
  // If match doesnt exist redirect to home
  const { matchid } = useParams()
  const [webSocket, setWebSocket] = useState<WebSocket | null>(null)
  const [matchState, setMatchState] = useState(null)

  console.log(`Matchroom: ${matchid}`)
  console.log(import.meta.env.VITE_API_MATCHROOM_URL + matchid + '/ws')

  useEffect(() => {
    // Connect to websocket for matchroom
    var ws = new WebSocket(import.meta.env.VITE_API_MATCHROOM_URL + matchid + '/ws')
    ws.onmessage = (event) => readMessage(event.data)
    setWebSocket(ws)
    return () => {
      // Disconnect
      ws?.close()
    }
  }, [])

  function readMessage(message: any) {
    console.log("FROM WEBSOCKET")
    console.log(message)
    JSON.parse(message)
  }

  return (
    <ChessBoard />
  )
}

export default App
