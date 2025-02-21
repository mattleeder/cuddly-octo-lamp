import React, { useContext, useRef, useEffect, useState } from "react";
import { CornerUpLeft, Handshake, Flag, Microscope, ChevronFirst, ChevronLeft, ChevronRight, ChevronLast, AlignJustify } from "lucide-react";
import { PieceColour, parseGameStateFromFEN } from "./ChessLogic";
import { GameContext, OpponentEventType, gameContext, boardHistory } from "./GameContext";


function isClockPaused(game: gameContext, colour: PieceColour) {
  if (!game) {
    return true
  }
  return game.matchData.gameOverStatus != 0 || game.matchData.activeColour != colour|| game.matchData.stateHistory.length <= 2
}

function isBlackClockPaused(game: gameContext) {
  return isClockPaused(game, PieceColour.Black)
}
  
function isWhiteClockPaused(game: gameContext) {
  return isClockPaused(game, PieceColour.White)
}
  
function sendDrawEvent(websocket: WebSocket | null) {
  if (!websocket) {
    console.error("Websocket is null")
    return
  }
  
  websocket.send(JSON.stringify({
    "messageType": "playerEvent",
    "body": {
      "eventType": OpponentEventType.Draw,
    }
  }))
}
  
function sendResignEvent(websocket: WebSocket | null) {
  if (!websocket) {
    console.error("Websocket is null")
    return
  }
  
  websocket.send(JSON.stringify({
    "messageType": "playerEvent",
    "body": {
      "eventType": OpponentEventType.Resign,
    }
  }))
}

function acceptEvent(game: gameContext, eventType: OpponentEventType) {
  if (game.webSocket == null) {
    console.error("Websocket is null")
    return
  }
  
  game.webSocket.send(JSON.stringify({
    "messageType": "playerEvent",
    "body": {
      "eventType": eventType,
    }
  }))

  game.setOpponentEventType(OpponentEventType.None)
  
}
  
function declineEvent(game: gameContext, isThreefold = false) {
  if (game.webSocket == null) {
    console.error("Websocket is null")
    return
  }

  if (isThreefold) {
    game.setThreefoldRepetition(false)
    return
  }
  
  game.webSocket.send(JSON.stringify({
    "messageType": "playerEvent",
    "body": {
      "eventType": OpponentEventType.Decline,
    }
  }))

  game.setOpponentEventType(OpponentEventType.None)
  
    
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

function GameControls() {
  const game = useContext(GameContext)
  if (!game) {
    throw new Error("GameControls must be used within a GameContext")
  }
  
  return (
    <div className='gameControlsContainer'>
      <div className='spacer' />
      <div className='gameControlsButton'>
        <CornerUpLeft size={12} color='#000000' />
      </div>
      <div className='gameControlsButton'>
        <Handshake onClick={() => sendDrawEvent(game.webSocket)} size={12} color='#000000' />
      </div>
      <div className='gameControlsButton'>
        <Flag onClick={() => sendResignEvent(game.webSocket)} size={12} color='#000000' />
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

  
function EventTypeDialog() {
  const game = useContext(GameContext)
  if (!game) {
    throw new Error("EventTypeDialog must be used within a gameContext")
  }

  console.log(`Threefold Repetition? ${game.threefoldRepetition}`)

  if (game.threefoldRepetition) {
    return (
      <div className="eventTypeDialog">
        <span>Threefold Repetition</span>
        <div>
          <button onClick={() => acceptEvent(game, OpponentEventType.ThreefoldRepetition)}>Accept</button>
          <button onClick={() => declineEvent(game, true)}>Decline</button>
        </div>
      </div>
    )
  }
  
  if (game.opponentEventType == OpponentEventType.None) {
    return <></>
  }
  
  return (
    <div className="eventTypeDialog">
      <span>Event</span>
      <div>
        <button onClick={() => acceptEvent(game, game.opponentEventType)}>Accept</button>
        <button onClick={() => declineEvent(game)}>Decline</button>
      </div>
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

export function GameInfoTile() {

  const game = useContext(GameContext)
  if (!game) {
    throw new Error('GameInfoTile must be used within a GameContext Provider');
  }
  
  return (
    <div>
      <CountdownTimer className="playerTimeBlack" paused={isBlackClockPaused(game)} countdownTimerMilliseconds={game.matchData.activeState.blackPlayerTimeRemainingMilliseconds}/>
      <div className='gameInfo'>
        <EventTypeDialog />
        <PlayerInfo connected={game.isWhiteConnected}/>
        <MoveHistoryControls />
        <Moves />
        <GameControls />
        <PlayerInfo connected={game.isBlackConnected}/>
      </div>
      <CountdownTimer className="playerTimeWhite" paused={isWhiteClockPaused(game)} countdownTimerMilliseconds={game.matchData.activeState.whitePlayerTimeRemainingMilliseconds}/>
    </div>
  )
}