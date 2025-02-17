import { Microscope, ChevronFirst, ChevronLeft, ChevronRight, ChevronLast, AlignJustify, CornerUpLeft, Flag, Handshake, LoaderCircle } from "lucide-react"
import { useState, useEffect, useRef, useContext, createContext, ReactNode } from "react"
import React from 'react';
import { useParams, useNavigate, useLocation } from "react-router-dom"
import { PieceColour, PieceVariant, parseGameStateFromFEN } from "./ChessLogic"

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

export function MatchRoom() {
  const { matchid } = useParams()
  const location = useLocation();
  const { timeFormatInMilliseconds } = location.state || {};
  const parsedTimeFormatInMilliseconds = parseInt(timeFormatInMilliseconds)

  return (
    <GameWrapper matchID={matchid as string} timeFormatInMilliseconds={parsedTimeFormatInMilliseconds}>
      <div className='chessMatch'>
        <GameInfoTile />
        <ChessBoard />
      </div>
    </GameWrapper>
  )
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

  enum ClickAction {
    clear,
    showMoves,
    makeMove,
    choosePromotion,
  }

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
    setToBare()
  }, [game.matchData.activeMove])

  function wsPostMove(position: number, piece: number, promotion: string) {
    game?.webSocket?.send(JSON.stringify({
      "messageType": "postMove",
      "body": {
        "piece": piece,
        "move": position,
        "promotionString": promotion,
      }
    }))
  }

  async function fetchPossibleMoves(position: number) {
    if (!game) {
      return
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

    const boardXPosition = Math.floor((event.clientX - rect.left) / (rect.width / 8))
    const boardYPosition = Math.floor((event.clientY - rect.top) / (rect.height / 8))
    let position = boardYPosition * 8 + boardXPosition

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
      setToBare()
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
      setToBare()
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
  
  const PiecesComponent = game?.matchData.activeState.board.map((square, idx) => {
    const [colour, variant] = square
    if (colour === null || variant === null) {
      return <React.Fragment key={idx} />
    }

    const row = Math.floor(idx / 8)
    const col = idx % 8
    return (
      <div key={idx} className={`${colourToString.get(colour)}-${variantToString.get(variant)}`} style={{ transform: `translate(${col * 50}px, ${row * 50}px)` }} />
    )
  })

  const MovesComponent = (moves).map((move, idx) => {
    const row = Math.floor(move / 8)
    const col = move % 8
    return (
      <div key={idx} className='potential-move' style={{ transform: `translate(${col * 50}px, ${row * 50}px)` }} />
    )
  })

  const CapturesComponent = captures.map((move, idx) => {
    const row = Math.floor(move / 8)
    const col = move % 8
    return (
      <div key={idx} className='potential-capture' style={{ transform: `translate(${col * 50}px, ${row * 50}px)` }} />
    )
  })

  const LastMoveComponent = game?.matchData.activeState.lastMove.map((move, idx) => {
    if (game.matchData.activeMove == 0) {
      return <React.Fragment key={idx} />
    }
    const row = Math.floor(move / 8)
    const col = move % 8
    return (
      <div key={idx} className='last-move' style={{ transform: `translate(${col * 50}px, ${row * 50}px)` }} />
    )
  })

  const PromotionComponent = ({ promotionSquare }: { promotionSquare: number }) => {
    if (!promotionActive) {
      return <></>
    }

    const col = promotionSquare % 8

    let promotionColour = colourToString.get(PieceColour.Black)
    let promotionDirection = -1
    let verticalOffset = 350

    if (promotionSquare <= 7) {
      promotionColour = colourToString.get(PieceColour.White)
      promotionDirection = 1
      verticalOffset = 0
    }

    return (
      <>
        <div className={`${promotionColour}-queen promotion`} style={{ transform: `translate(${col * 50}px, ${verticalOffset}px)` }} />
        <div className={`${promotionColour}-knight promotion`} style={{ transform: `translate(${col * 50}px, ${verticalOffset + promotionDirection * 50}px)` }} />
        <div className={`${promotionColour}-rook promotion`} style={{ transform: `translate(${col * 50}px, ${verticalOffset + promotionDirection * 50 * 2}px)` }} />
        <div className={`${promotionColour}-bishop promotion`} style={{ transform: `translate(${col * 50}px, ${verticalOffset + promotionDirection * 50 * 3}px)` }} />
      </>
    )
  }

  const GameOverComponent = () => {
    if (game?.matchData.gameOverStatus == 0) {
      return <></>
    }

    const gameOverStatusCodes = ["Ongoing", "Stalemate", "Checkmate", "Threefold Repetition", "Insufficient Material", "White Flagged", "Black Flagged"]
    const gameOverText = gameOverStatusCodes[game?.matchData.gameOverStatus || 0]

    return <div style={{ transform: `translate(${0}px, ${180}px)`, color: "black" }}>{gameOverText}</div>
  }

  return (
    <div className='chessboard' onClick={clickHandler} ref={boardRef}>
      {LastMoveComponent}
      {PiecesComponent}
      {MovesComponent}
      {CapturesComponent}
      <PromotionComponent promotionSquare={promotionSquare} />
      <GameOverComponent />
    </div>
  )
}

interface boardInfo {
  board: [PieceColour | null, PieceVariant | null][],
  lastMove: [number, number],
  FEN: string,
  whitePlayerTimeRemainingMilliseconds: number
  blackPlayerTimeRemainingMilliseconds: number
}

interface boardHistory {
  FEN: string,
  lastMove: [number, number]
  algebraicNotation: string
  whitePlayerTimeRemainingMilliseconds: number
  blackPlayerTimeRemainingMilliseconds: number
}

interface matchData {
  activeState: boardInfo,
  stateHistory: boardHistory[],
  activeColour: PieceColour,
  activeMove: number,
  gameOverStatus: number,
}

interface gameContext {
  matchData: matchData,
  setMatchData: React.Dispatch<React.SetStateAction<matchData>>,
  webSocket: WebSocket | null,
  playerColour: PieceColour,
  isWhiteConnected: boolean,
  isBlackConnected: boolean,
}

const GameContext = createContext<gameContext | null>(null)

interface MatchStateHistory {
  FEN: string
  lastMove: [number, number]
  algebraicNotation: string
  whitePlayerTimeRemainingMilliseconds: number
  blackPlayerTimeRemainingMilliseconds: number
}

interface OnConnectMessage {
  matchStateHistory: MatchStateHistory[]
  gameOverStatus: number
  threefoldRepetition: boolean
  whitePlayerConnected: boolean
  blackPlayerConnected: boolean
}

interface OnMoveMessage {
  matchStateHistory: MatchStateHistory[]
  gameOverStatus: number
  threefoldRepetition: boolean
}

interface ConnectionStatusMessage {
  playerColour: string
  isConnected: boolean
}

interface PlayerCodeMessage {
  playerCode: number
}

interface ChessWebSocketMessage {
  messageType: string
  body: OnConnectMessage | OnMoveMessage | ConnectionStatusMessage | PlayerCodeMessage
}

function GameWrapper({ children, matchID, timeFormatInMilliseconds }: { children: ReactNode, matchID: string, timeFormatInMilliseconds: number }) {
  const [matchData, setMatchData] = useState<matchData>(
    {
      activeState: {
        board: parseGameStateFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")["board"],
        lastMove: [0, 0],
        FEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
        whitePlayerTimeRemainingMilliseconds: timeFormatInMilliseconds,
        blackPlayerTimeRemainingMilliseconds: timeFormatInMilliseconds,
      },
      stateHistory: [{
        FEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
        lastMove: [0, 0],
        algebraicNotation: "",
        whitePlayerTimeRemainingMilliseconds: timeFormatInMilliseconds,
        blackPlayerTimeRemainingMilliseconds: timeFormatInMilliseconds,
      }],
      activeColour: PieceColour.White,
      activeMove: 0,
      gameOverStatus: 0,
    })
  const [webSocket, setWebSocket] = useState<WebSocket | null>(null)
  const [playerColour, setPlayerColour] = useState(PieceColour.Spectator)
  const [isWhiteConnected, setIsWhiteConnected] = useState(false)
  const [isBlackConnected, setIsBlackConnected] = useState(false)

  useEffect(() => {
    // Connect to websocket for matchroom
    console.log("Connecting to ws")
    const ws = new WebSocket(import.meta.env.VITE_API_MATCHROOM_URL + matchID + '/ws')
    setWebSocket(ws)
    return () => {
      // Disconnect
      ws?.close()
    }
  }, [])

  useEffect(() => {
    console.log("GAMEWRAPPER:")
    console.log(matchData)
  }, [matchData])

  if (webSocket) {
    webSocket.onmessage = (event) => readMessage(event.data)
    webSocket.onerror = (event) => console.error(event)
    webSocket.onclose = () => console.log("Websocket closed")
  }

  function sendPlayerCodeHandler(body: PlayerCodeMessage) {
    if (body["playerCode"] == 0) {
      setPlayerColour(PieceColour.White)
    } else if (body["playerCode"] == 1) {
      setPlayerColour(PieceColour.Black)
    }
  }

  function onConnectHandler(body: OnConnectMessage) {
    onMoveHandler(body as OnMoveMessage)
    setIsWhiteConnected(body["whitePlayerConnected"])
    setIsBlackConnected(body["blackPlayerConnected"])
  }

  function connectionStatusHandler(body: ConnectionStatusMessage) {
    if (body["playerColour"] == "white") {
      setIsWhiteConnected(body["isConnected"])
    } else if (body["playerColour"] == "black") {
      setIsBlackConnected(body["isConnected"])
    }

  }

  function onMoveHandler(body: OnMoveMessage) {
    const newHistory = body["matchStateHistory"]
    if (newHistory.length == 0) {
      console.error("New history has length 0")
      return
    }
    const latestHistoryEntry = newHistory.at(-1) as MatchStateHistory
    const latestFEN = latestHistoryEntry["FEN"]
    const activeColour = parseGameStateFromFEN(latestFEN)["activeColour"]
    const gameOverStatus = body["gameOverStatus"]

    let activeState = {
      ...matchData.activeState,
      whitePlayerTimeRemainingMilliseconds:  latestHistoryEntry["whitePlayerTimeRemainingMilliseconds"],
      blackPlayerTimeRemainingMilliseconds:  latestHistoryEntry["blackPlayerTimeRemainingMilliseconds"],
    }
    let activeMove = matchData.activeMove

    if (matchData.activeState.FEN == matchData.stateHistory.at(-1)?.FEN) {
      activeState = {
        ...activeState,
        board: parseGameStateFromFEN(latestFEN).board,
        lastMove: latestHistoryEntry["lastMove"],
        FEN: latestFEN,
      }
      activeMove = newHistory.length - 1
    }

    const newMatchData: matchData = {
      activeState: activeState,
      stateHistory: newHistory,
      activeColour: activeColour,
      gameOverStatus: gameOverStatus,
      activeMove: activeMove,
    }

    setMatchData(newMatchData)
  }

  function readMessage(message: unknown) {
    console.log("FROM WEBSOCKET")
    console.log(message)

    if (typeof message != "string") {
      return
    }

    for (const msg of message.split("\n")) {
      const parsedMsg: ChessWebSocketMessage = JSON.parse(msg)

      const messageType = parsedMsg["messageType"]

      switch (messageType) {
      case "sendPlayerCode":
        sendPlayerCodeHandler(parsedMsg["body"] as PlayerCodeMessage)
        break;
      case "onConnect":
        onConnectHandler(parsedMsg["body"] as OnConnectMessage)
        break;
      case "connectionStatus":
        connectionStatusHandler(parsedMsg["body"] as ConnectionStatusMessage)
        break;
      case "onMove":
        onMoveHandler(parsedMsg["body"] as OnMoveMessage)
        break;
      default:
        console.error("Could not understand message from websocket")
        console.log(message)
      }
    }
  }

  return (
    <GameContext.Provider value={{ matchData, setMatchData, webSocket, playerColour, isWhiteConnected, isBlackConnected }}>
      {children}
    </GameContext.Provider>
  )
}

function GameInfoTile() {

  const game = useContext(GameContext)
  if (!game) {
    throw new Error('GameInfoTile must be used within a GameContext Provider');
  }

  function PlayerInfo({ connected }: { connected: boolean }) {
    let classname = "playerPingStatus"
    if (connected) {
      classname += " connected"
    }
    return (
      <div className='playerInfo'>
        <div className={classname}></div>
        <div className='playerName'>Player</div>
      </div>
    )
  }

  function isBlackClockPaused() {
    if (!game) {
      return true
    }
    return game.matchData.activeColour != PieceColour.Black || game.matchData.stateHistory.length <= 2
  }

  function isWhiteClockPaused() {
    if (!game) {
      return true
    }
    return game.matchData.activeColour != PieceColour.White || game.matchData.stateHistory.length <= 2
  }

  return (
    <div>
      <CountdownTimer className="playerTimeBlack" paused={isBlackClockPaused()} countdownTimerMilliseconds={game.matchData.activeState.blackPlayerTimeRemainingMilliseconds}/>
      <div className='gameInfo'>
        <PlayerInfo connected={game.isWhiteConnected}/>
        <MoveHistoryControls />
        <Moves />
        <GameControls />
        <PlayerInfo connected={game.isBlackConnected}/>
      </div>
      <CountdownTimer className="playerTimeWhite" paused={isWhiteClockPaused()} countdownTimerMilliseconds={game.matchData.activeState.whitePlayerTimeRemainingMilliseconds}/>
    </div>
  )
}

function GameControls() {

  return (
    <div className='gameControlsContainer'>
      <div className='spacer' />
      <div className='gameControlsButton'>
        <CornerUpLeft size={12} color='#000000' />
      </div>
      <div className='gameControlsButton'>
        <Handshake size={12} color='#000000' />
      </div>
      <div className='gameControlsButton'>
        <Flag size={12} color='#000000' />
      </div>
      <div className='spacer' />
    </div>
  )
}

function Moves() {
  const game = useContext(GameContext)
  if (!game) {
    throw new Error('Move History must be used within a GameContext Provider');
  }

  const boardHistory = game.matchData.stateHistory
  const tableRef = useRef<HTMLDivElement | null>(null)
  const activeRef = useRef<HTMLTableCellElement | null>(null)


  useEffect(() => {
    if (activeRef.current) {
      activeRef.current.scrollIntoView({ behavior: "auto", block: "nearest" })
    }
  }, [game.matchData.activeMove])

  const tableData: boardHistory[][] = []
  for (let i = 1; i < boardHistory.length; i += 2) {
    const rowData: boardHistory[] = []
    rowData.push(boardHistory[i])
    if (i + 1 < boardHistory.length) {
      rowData.push(boardHistory[i + 1])
    }
    tableData.push(rowData)
  }


  return (
    <div className='movesContainer' ref={tableRef}>
      <table>
        <tbody>
          {tableData.map((data, idx) => {
            return (
              <tr key={idx} className='movesRow'>
                <td>{Math.floor(idx) + 1}</td>
                <td
                  onClick={() => updateActiveState(idx * 2 + 1, game)}
                  className={game.matchData.activeMove == idx * 2 + 1 ? "highlight" : ""}
                  ref={(game.matchData.activeMove == idx * 2 + 1) || (game.matchData.activeMove == 0 && idx == 1) ? activeRef : null}
                >
                  {data[0]["algebraicNotation"]}
                </td>
                {
                  data.length > 1 ?
                    <td
                      onClick={() => updateActiveState(idx * 2 + 2, game)}
                      className={game.matchData.activeMove == idx * 2 + 2 ? "highlight" : ""}
                      ref={game.matchData.activeMove == idx * 2 + 2 ? activeRef : null}
                    >
                      {data[1]["algebraicNotation"]}
                    </td>
                    :
                    <></>
                }
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )

}

function MoveHistoryControls() {

  const game = useContext(GameContext)

  if (!game) {
    return
  }

  let latestMoveButtonClassName = "moveHistoryControlsButton"
  if (game.matchData.activeMove != game.matchData.stateHistory.length - 1) {
    latestMoveButtonClassName += " newMoveNotification"
  }

  return (
    <div className='moveHistoryControlsContainer'>
      <div className='moveHistoryControlsButton'>
        <Microscope size={12} />
      </ div>
      <div onClick={() => updateActiveState(0, game)} className='moveHistoryControlsButton'>
        <ChevronFirst size={12} />
      </ div>
      <div onClick={() => updateActiveState(game?.matchData.activeMove - 1, game)} className='moveHistoryControlsButton'>
        <ChevronLeft size={12} />
      </ div>
      <div className='moveHistoryControlsButton'>
        <ChevronRight onClick={() => updateActiveState(game?.matchData.activeMove + 1, game)} size={12} />
      </ div>
      <div onClick={() => updateActiveState(game?.matchData.stateHistory.length - 1, game)} className={latestMoveButtonClassName}>
        <ChevronLast size={12} />
      </ div>
      <div className='moveHistoryControlsButton'>
        <AlignJustify size={12} />
      </ div>
    </div>
  )
}

function updateActiveState(stateHistoryIndex: number, game: gameContext) {
  console.log("updateActiveState Called")
  console.log(stateHistoryIndex)
  if (!game) {
    throw new Error("updateActiveState must be called within a GameContext")
  }
  if (stateHistoryIndex == game.matchData.activeMove) {
    return
  }
  if (stateHistoryIndex < 0 || game.matchData.stateHistory.length - 1 < stateHistoryIndex) {
    return
  }
  const activeMoveNumber = stateHistoryIndex
  const matchData = {
    ...game.matchData
  }
  matchData.activeMove = activeMoveNumber
  matchData.activeState = {
    board: parseGameStateFromFEN(matchData.stateHistory[activeMoveNumber]["FEN"])["board"],
    lastMove: matchData.stateHistory[activeMoveNumber]["lastMove"],
    FEN: matchData.stateHistory[activeMoveNumber]["FEN"],
    whitePlayerTimeRemainingMilliseconds: matchData.activeState.whitePlayerTimeRemainingMilliseconds,
    blackPlayerTimeRemainingMilliseconds: matchData.activeState.blackPlayerTimeRemainingMilliseconds,
  }
  game.setMatchData(matchData)
}

function CountdownTimer({ countdownTimerMilliseconds, paused, className } : { countdownTimerMilliseconds: number, paused: boolean, className: string }) {
  // On mount, record the time using Date.now(), use a prop to get the count
  // Create a state to hold the remaining time
  // Create a function in a use effect on mount that sets the remaining time by taking
  // the elapsed time from the initial count
  //
  const [remainingTime, setRemainingTime] = useState(countdownTimerMilliseconds)

  // May need to be called when countdownTimerMilliseconds changes
  useEffect(() => {
    const start = Date.now()
    const updateTimer = () => {
      const delta = Date.now() - start
      setRemainingTime(countdownTimerMilliseconds - delta)
    }

    if (paused) {
      return
    }

    const intervalID = setInterval(updateTimer, 1000)

    return () => {
      clearInterval(intervalID)      
    }
  }, [paused])

  return (
    <div className={className}>
      {formatDuration(remainingTime)}
    </div>
  )
}

function formatDuration(durationInMilliseconds: number): string {
  if (durationInMilliseconds <= 0) {
    return "00:00:00.0"
  }
  const time = new Date(durationInMilliseconds);
  // const hours = time.getUTCHours();
  const minutes = time.getUTCMinutes();
  const seconds = time.getUTCSeconds();
  const milliseconds = time.getUTCMilliseconds();

  let result = String(minutes).padStart(2, "0") + ":" + String(seconds).padStart(2, "0")
  if (durationInMilliseconds < 10_000) {
    result += "." + milliseconds.toPrecision(1)[0]
  }
  return result;
}

interface QueueObject {
  timeFormatInMilliseconds: number,
  incrementInMilliseconds: number,
}

const queueObjectsMap = new Map<string, QueueObject>()

function addQueueObject(timeFormatInMinutes: number, incrementInSeconds: number) {
  queueObjectsMap.set(`${timeFormatInMinutes} + ${incrementInSeconds}`, {
    timeFormatInMilliseconds: timeFormatInMinutes * 60 * 1000,
    incrementInMilliseconds: incrementInSeconds * 1000,
  })
}

addQueueObject(1, 0)
addQueueObject(2, 1)
addQueueObject(3, 0)

addQueueObject(3, 2)
addQueueObject(5, 0)
addQueueObject(5, 3)

addQueueObject(10, 0)
addQueueObject(10, 5)
addQueueObject(15, 10)

export function QueueTiles() {
  const [waiting, setWaiting] = useState(false)
  const [inQueue, setInQueue] = useState(false)
  const [queueName, setQueueName] = useState("")
  const queueNameRef = useRef(queueName)
  const [eventSource, setEventSource] = useState<EventSource | null>(null)
  const navigate = useNavigate()

  useEffect(() => {
    queueNameRef.current = queueName
  }, [queueName])

  useEffect(() => {
    const leaveOnUnmount = async () => {
      try {
        await tryLeaveQueue(queueNameRef.current)
      } catch (e) {
        console.error(e)
      }
    }
    return () => {
      leaveOnUnmount()
    }
  }, [])

  enum ClickAction {
    leaveQueue,
    joinQueue,
    changeQueue,
  }

  async function tryJoinQueue(queueName: string) {
    const queueObject = queueObjectsMap.get(queueName)
    if (queueObject === undefined) {
      throw new Error("Queue object not found")
    }

    const response = await fetch(import.meta.env.VITE_API_JOIN_QUEUE_URL, {
      signal: AbortSignal.timeout(5000),
      method: "POST",
      credentials: 'include',
      body: JSON.stringify({
        "timeFormatInMilliseconds": queueObject.timeFormatInMilliseconds,
        "incrementInMilliseconds": queueObject.incrementInMilliseconds,
        "action": "join",
      })
    })

    if (!response.ok) {
      throw new Error(response.statusText)
    }

    // Joined, start listening for events
    const eventSource = new EventSource(import.meta.env.VITE_API_MATCH_LISTEN_URL, {
      withCredentials: true,
    })
    eventSource.onmessage = (event) => {
      console.log(`message: ${event.data}`)
      const splitData = event.data.split(",")
      const matchRoom = splitData[0]
      const timeFormatInMilliseconds = splitData[1]
      const incrementInMilliseconds = splitData[2]
      const state = {
        timeFormatInMilliseconds,
        incrementInMilliseconds,
      }
      navigate("matchroom/" + matchRoom, { state })
    }
    setEventSource(eventSource)
  }

  async function tryLeaveQueue(queueName: string) {
    const queueObject = queueObjectsMap.get(queueName)
    if (queueObject === undefined) {
      throw new Error("Queue object not found")
    }
    const response = await fetch(import.meta.env.VITE_API_JOIN_QUEUE_URL, {
      method: "POST",
      credentials: 'include',
      body: JSON.stringify({
        "timeFormatInMilliseconds": queueObject.timeFormatInMilliseconds,
        "incrementInMilliseconds": queueObject.incrementInMilliseconds,
        "action": "leave",
      })
    })

    if (!response.ok) {
      throw new Error(response.statusText)
    }

    // Left
    eventSource?.close()
  }

  async function toggleQueue(newQueueName: string) {
    if (waiting) {
      return
    }

    setWaiting(true)
    let clickAction
    if (!inQueue) {
      clickAction = ClickAction.joinQueue
    } else if (queueName == newQueueName) {
      clickAction = ClickAction.leaveQueue
    } else {
      clickAction = ClickAction.changeQueue
    }

    try {
      switch(clickAction) {
      case ClickAction.leaveQueue:
        await tryLeaveQueue(queueName)
        setInQueue(false)
        setQueueName("")
        break

      case ClickAction.changeQueue:
        await tryLeaveQueue(queueName)
        await tryJoinQueue(newQueueName)
        setQueueName(newQueueName)
        break
      
      case ClickAction.joinQueue:
        await tryJoinQueue(newQueueName)
        setInQueue(true)
        setQueueName(newQueueName)
      }
    } catch (e) {
      console.error(e)
    } finally {
      setWaiting(false)
    }

  }

  function QueueButton({ nameOfQueue, queueType }: { nameOfQueue: string, queueType: string }) {
    const loading = nameOfQueue == queueName
    return (
      <>
        {loading ?
          <button onClick={() => toggleQueue(nameOfQueue)}><LoaderCircle className="loaderSpin"/></button>
          :
          <button onClick={() => toggleQueue(nameOfQueue)}><span>{nameOfQueue}<br />{queueType}</span></button>
        }
      </>
    )
  }

  return (
    <div className="queueTilesContainer">
      <QueueButton nameOfQueue="1 + 0" queueType="Bullet"/>
      <QueueButton nameOfQueue="2 + 1" queueType="Bullet"/>
      <QueueButton nameOfQueue="3 + 0" queueType="Blitz"/>

      <QueueButton nameOfQueue="3 + 2" queueType="Blitz"/>
      <QueueButton nameOfQueue="5 + 0" queueType="Blitz"/>
      <QueueButton nameOfQueue="5 + 3" queueType="Blitz"/>

      <QueueButton nameOfQueue="10 + 0" queueType="Rapid"/>
      <QueueButton nameOfQueue="10 + 5" queueType="Rapid"/>
      <QueueButton nameOfQueue="15 + 10" queueType="Rapid"/>
    </div>
  )
}