import { createContext, useContext, useEffect, useRef, useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import { parseGameStateFromFEN, PieceVariant, PieceColour } from './Chess.tsx'
import {
  BrowserRouter as Router,
  Routes,
  Route,
  useParams,
  useNavigate,
} from "react-router-dom";
import { AlignJustify, ChevronFirst, ChevronLast, ChevronLeft, ChevronRight, CornerUpLeft, Equal, Flag, Microscope } from 'lucide-react'

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
    <p className="read-the-docs">
      Click on the Vite and React logos to learn more
    </p>
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

const GameContext = createContext(null)

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
  const [playerColour, setPlayerColour] = useState(PieceColour.Spectator)


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
      var parsedMsg = JSON.parse(msg)[0]
      console.log(parsedMsg)
      if (parsedMsg.hasOwnProperty("playerCode")) {
        var playerCode = parsedMsg["playerCode"]
        if (playerCode == 0) {
          setPlayerColour(PieceColour.White)
        } else if (playerCode == 1) {
          setPlayerColour(PieceColour.Black)
        }
      } else if (parsedMsg.hasOwnProperty("newFEN")) {
        console.log("Setting")
        setGameState(parseGameStateFromFEN(parsedMsg["newFEN"]))
        if (parsedMsg["lastMove"][0] != parsedMsg["lastMove"][1]) {
          setLastMove(parsedMsg["lastMove"])
        }
        setGameOverStatus(parsedMsg["gameOverStatus"])
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

  function setToBare() {
    setSelectedPiece(null)
    setMoves([])
    setCaptures([])
    setPromotionNextMove(false)
    setPromotionActive(false)
    setPromotionSquare(0)
  }

  async function clickHandler(event: React.MouseEvent) {
    console.log(`${event.clientX}, ${event.clientY}`)
    // Calculate board position
    if (rect === null) {
      throw new Error("Bounding rect for board is not defined")
    }
    console.log(rect)
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
    console.log(`Click Action: ${position}`)
    console.log(`Click Action: ${gameState.board[position][0]}`)
    var clickAction = ClickAction.clear
    if (promotionActive && [0, 8, 16, 24].includes(Math.abs(position - promotionSquare))) {
      clickAction = ClickAction.choosePromotion
    } else if ([...moves, ...captures].includes(position)) {
      clickAction = ClickAction.makeMove
    }else if (gameState.board[position][0] == playerColour && position != selectedPiece) {
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
        wsPostMove(position, selectedPiece, promotionString)

        // Clear cache, clear moves
        setToBare()
        break;

      case ClickAction.showMoves:
        // Fetch moves
        var data = await fetchPossibleMoves(position)
        
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

  function GameInfo() {

    function PlayerInfo() {
      return (
        <div className='playerInfo'>
          <div className='playerPingStatus'>O</div>
          <div className='playerName'>Player</div>
        </div>
      )
    }

    function MoveHistoryControls() {
      return (
        <div className='moveHistoryControlsContainer'>
          <div className='moveHistoryControlsButton'>
            <Microscope size={12}/>
          </ div>  
          <div className='moveHistoryControlsButton'>
            <ChevronFirst size={12}/>
          </ div>
          <div className='moveHistoryControlsButton'>
            <ChevronLeft size={12}/>
          </ div>          
          <div className='moveHistoryControlsButton'>
            <ChevronRight size={12}/>
          </ div>          
          <div className='moveHistoryControlsButton'>
            <ChevronLast size={12}/>
          </ div>          
          <div className='moveHistoryControlsButton'>
            <AlignJustify size={12}/>
          </ div>
        </div>
      )
    }

    function GameControls() {
      return (
        <div className='gameControlsContainer'>
          <div className='spacer' />
          <div className='gameControlsButton'>
            <CornerUpLeft size={12} color='#000000'/>
          </div>
          <div className='gameControlsButton'>
            <Equal size={12} color='#000000'/>
          </div>
          <div className='gameControlsButton'>
            <Flag size={12} color='#000000'/>
          </div>
          <div className='spacer' />
        </div>
      )
    }

    interface moveHistoryRowData {
      rowNumber: string,
      leftMove: string,
      leftFEN: string,
      rightMove: string | null,
      rightFEN: string | null
    }

    var moveMap = new Map([
      ["e4", "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1"],
      ["d5", "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 2"],
      ["nf3", "rnbqkbnr/ppp1pppp/8/3p4/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2"],
      ["bg4", "rn1qkbnr/ppp1pppp/8/3p4/4P1b1/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3"],
      ["d4", "rn1qkbnr/ppp1pppp/8/3p4/3PP1b1/5N2/PPP2PPP/RNBQKB1R b KQkq - 0 3"],
    ])

    function Moves({ moveMap } : { moveMap: Map<string, string> }) {
      const [orderedMap, setOrderedMap] = useState([...moveMap.entries()])
      const [selectedMove, setSelectedMove] = useState(orderedMap.length)

      var tableData = []

      for (var i = 0; i < orderedMap.length; i+=2) {
        var rowNumber = `${Math.floor(i / 2) + 1}.`
        var leftMove = orderedMap[i][0]
        var leftFEN = orderedMap[i][1]
        var rightMove, rightFEN: string | null

        if ((i+1) < orderedMap.length) {
          rightMove = orderedMap[i+1][0]
          rightFEN = orderedMap[i+1][1]
        } else {
          rightMove = null
          rightFEN = null
        }

        var rowData: moveHistoryRowData = {
          rowNumber,
          leftMove,
          leftFEN,
          rightMove,
          rightFEN,
        }

        tableData.push(rowData)
      }


      return (
        <div className='movesContainer'>
          <table>
            <tbody>
              {tableData.map((rowData) => {
                return (
                  <tr className='movesRow'>
                    <td>{rowData.rowNumber}</td>
                    <td onClick={() => setGameState(parseGameStateFromFEN(rowData.leftFEN))}>{rowData.leftMove}</td>
                    {rowData.rightMove ? <td onClick={() => setGameState(parseGameStateFromFEN(rowData.rightFEN as string))}>{rowData.rightMove}</td> : <></>}
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )

    }


    return (
      <div className='gameInfo'>
        <PlayerInfo />
        <MoveHistoryControls />
        <Moves moveMap={moveMap}/>
        <GameControls />
        <PlayerInfo />
      </div>
    )
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
    <div className='chessMatch'>
      <GameInfo />
      <div className='chessboard' onClick={clickHandler} ref={boardRef}>
        {LastMoveComponent}
        {PiecesComponent}
        {MovesComponent}
        {CapturesComponent}
        <PromotionComponent promotionSquare={promotionSquare}/>
        <GameOverComponent />
      </div>
    </div>
  )
}

export default App
