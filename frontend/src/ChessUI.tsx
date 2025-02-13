import { Microscope, ChevronFirst, ChevronLeft, ChevronRight, ChevronLast, AlignJustify, CornerUpLeft, Flag, Handshake } from "lucide-react"
import { useState, useEffect, useRef, useContext, createContext, ReactNode } from "react"
import React from 'react';
import { useParams, useNavigate } from "react-router-dom"
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

  return (
    <GameWrapper matchID={matchid as string}>
      <div className='chessMatch'>
        <GameInfoTile />
        <ChessBoard />
      </div>
    </GameWrapper>
  )
}

export function JoinQueue() {
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
      const response = await fetch(import.meta.env.VITE_API_JOIN_QUEUE_URL, {
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
      const eventSource = new EventSource(import.meta.env.VITE_API_MATCH_LISTEN_URL, {
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
      const response = await fetch(import.meta.env.VITE_API_JOIN_QUEUE_URL, {
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
    const clickAction = inQueue ? ClickAction.leaveQueue : ClickAction.joinQueue

    if (clickAction == ClickAction.leaveQueue) {
      const result = await leaveQueue()
      setInQueue(!result)
    } else {
      const result = await joinQueue()
      setInQueue(result)
    }

    setWaiting(false)
  }

  const buttonText = inQueue ? "In Queue" : "Join Queue"
  return (
    <button onClick={toggleQueue}>{buttonText}</button>
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

  useEffect(() => {
    console.log('ChessBoard Component mounted');
    return () => {
      console.log('ChessBoard Component unmounted');
    };
  }, []);

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
      "piece": piece,
      "move": position,
      "promotionString": promotion,
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
        console.log(error)
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
    console.log(rect)
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
    console.log(`Position : ${position}`)
    console.log(`Clicked On: ${game?.matchData.activeState.board[position][0]}`)
    let clickAction = ClickAction.clear
    if (game?.matchData.activeMove != (game as gameContext)?.matchData.stateHistory.length - 1) {
      console.log("Clicking disable in past moves")
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
      console.log("Post move")
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

    const gameOverStatusCodes = ["Ongoing", "Stalemate", "Checkmate", "Threefold Repetition", "Insufficient Material"]
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
}

interface boardHistory {
  FEN: string,
  lastMove: [number, number]
  algebraicNotation: string
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
}

const GameContext = createContext<gameContext | null>(null)

function GameWrapper({ children, matchID }: { children: ReactNode, matchID: string }) {
  const [matchData, setMatchData] = useState<matchData>(
    {
      activeState: {
        board: parseGameStateFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")["board"],
        lastMove: [0, 0],
        FEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
      },
      stateHistory: [{
        FEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
        lastMove: [0, 0],
        algebraicNotation: "",
      }],
      activeColour: PieceColour.White,
      activeMove: 0,
      gameOverStatus: 0,
    })
  const [webSocket, setWebSocket] = useState<WebSocket | null>(null)
  const [playerColour, setPlayerColour] = useState(PieceColour.Spectator)

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
  }

  function readMessage(message: unknown) {
    console.log("FROM WEBSOCKET")
    console.log(message)

    if (typeof message != "string") {
      return
    }

    for (const msg of message.split("\n")) {
      const parsedMsg = JSON.parse(msg)[0]

      if (Object.prototype.hasOwnProperty.call(parsedMsg, "playerCode")) {
        const playerCode = parsedMsg["playerCode"]

        if (playerCode == 0) {
          setPlayerColour(PieceColour.White)
        } else if (playerCode == 1) {
          setPlayerColour(PieceColour.Black)
        }

      } else if (Object.prototype.hasOwnProperty.call(parsedMsg, "pastMoves")) {

        console.log("Setting")
        const newHistory = parsedMsg["pastMoves"]
        const activeColour = parseGameStateFromFEN(parsedMsg["pastMoves"].at(-1)["FEN"])["activeColour"]
        const gameOverStatus = parsedMsg["gameOverStatus"]
        let activeState = {
          ...matchData.activeState
        }
        let activeMove = matchData.activeMove

        console.log(matchData)
        if (matchData.activeState.FEN == matchData.stateHistory.at(-1)?.FEN) {
          activeState = {
            board: parseGameStateFromFEN(newHistory.at(-1)["FEN"]).board,
            lastMove: newHistory.at(-1)["lastMove"],
            FEN: parsedMsg["pastMoves"].at(-1)["FEN"],
          }
          console.log("Increment activeMove")
          activeMove = newHistory.length - 1
        }

        const newMatchData: matchData = {
          activeState: activeState,
          stateHistory: newHistory,
          activeColour: activeColour,
          gameOverStatus: gameOverStatus,
          activeMove: activeMove,
        }

        console.log("New Match Data: ", newMatchData)

        setMatchData(newMatchData)

      }
    }
  }

  return (
    <GameContext.Provider value={{ matchData, setMatchData, webSocket, playerColour }}>
      {children}
    </GameContext.Provider>
  )
}

function GameInfoTile() {

  useEffect(() => {
    console.log('GameInfoTile Component mounted');
    return () => {
      console.log('GameInfoTile Component unmounted');
    };
  }, []);

  const game = useContext(GameContext)
  if (!game) {
    throw new Error('GameInfoTile must be used within a GameContext Provider');
  }

  function updateActiveState(stateHistoryIndex: number) {
    console.log("updateActiveState Called")
    console.log(stateHistoryIndex)
    if (!game) {
      return
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
    }
    console.log("MOVEHISTORY")
    console.log(matchData)
    game.setMatchData(matchData)
  }

  function PlayerInfo() {
    return (
      <div className='playerInfo'>
        <div className='playerPingStatus'>O</div>
        <div className='playerName'>Player</div>
      </div>
    )
  }

  function MoveHistoryControls() {
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
        <div onClick={() => updateActiveState(0)} className='moveHistoryControlsButton'>
          <ChevronFirst size={12} />
        </ div>
        <div onClick={() => updateActiveState(game?.matchData.activeMove - 1)} className='moveHistoryControlsButton'>
          <ChevronLeft size={12} />
        </ div>
        <div className='moveHistoryControlsButton'>
          <ChevronRight onClick={() => updateActiveState(game?.matchData.activeMove + 1)} size={12} />
        </ div>
        <div onClick={() => updateActiveState(game?.matchData.stateHistory.length - 1)} className={latestMoveButtonClassName}>
          <ChevronLast size={12} />
        </ div>
        <div className='moveHistoryControlsButton'>
          <AlignJustify size={12} />
        </ div>
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

    
    const gameCtx = useContext(GameContext)
    if (!gameCtx) {
      throw new Error('ChessBoard must be used within a GameContext Provider');
    }
    const boardHistory = gameCtx.matchData.stateHistory
    const scrollPosRef = useRef(0)
    const tableRef = useRef<HTMLDivElement | null>(null)

    useEffect(() => {
      console.log('Moves Component mounted');
      return () => {
        console.log('Moves Component unmounted');
      };
    }, []);
    


    function handleScroll() {
      console.log("Handling scroll")
      if (tableRef.current) {
        console.log(`Setting scrollPos to ${tableRef.current.scrollTop}`)
        if (scrollPosRef.current != null) {
          scrollPosRef.current = tableRef.current.scrollTop
        }
      }
    }

    useEffect(() => {
      console.log(`scrollPosRef: ${scrollPosRef}`)
      console.log(scrollPosRef)
      if (scrollPosRef.current != null) {
        console.log(`scrollPosRef: ${scrollPosRef.current}`)
      }
      if (tableRef.current) {
        console.log("Scrolling")
      }
    }, [boardHistory])

    const tableData: boardHistory[][] = []
    for (let i = 1; i < boardHistory.length; i += 2) {
      const rowData: boardHistory[] = []
      rowData.push(boardHistory[i])
      if (i + 1 < boardHistory.length) {
        rowData.push(boardHistory[i + 1])
      }
      tableData.push(rowData)
    }

    //@TODO
    // Make some setBoardState instead for move history to use


    return (
      <div className='movesContainer' ref={tableRef} onScroll={handleScroll}>
        <table>
          <tbody>
            {tableData.map((data, idx) => {
              return (
                <tr key={idx} className='movesRow'>
                  <td>{Math.floor(idx) + 1}</td>
                  <td
                    onClick={() => updateActiveState(idx * 2 + 1)}
                    className={gameCtx.matchData.activeMove == idx * 2 + 1 ? "highlight" : ""}
                  >
                    {data[0]["algebraicNotation"]}
                  </td>
                  {
                    data.length > 1 ?
                      <td
                        onClick={() => updateActiveState(idx * 2 + 2)}
                        className={gameCtx.matchData.activeMove == idx * 2 + 2 ? "highlight" : ""}
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

  return (
    <div className='gameInfo'>
      <PlayerInfo />
      <MoveHistoryControls />
      <Moves />
      <GameControls />
      <PlayerInfo />
    </div>
  )
}