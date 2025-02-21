import React from 'react';
import { ReactNode, useState, useEffect, createContext } from "react"
import { parseGameStateFromFEN, PieceColour, PieceVariant } from "./ChessLogic"

export interface boardInfo {
  board: [PieceColour | null, PieceVariant | null][],
  lastMove: [number, number],
  FEN: string,
  whitePlayerTimeRemainingMilliseconds: number
  blackPlayerTimeRemainingMilliseconds: number
}

export interface boardHistory {
  FEN: string,
  lastMove: [number, number]
  algebraicNotation: string
  whitePlayerTimeRemainingMilliseconds: number
  blackPlayerTimeRemainingMilliseconds: number
}

export interface matchData {
  activeState: boardInfo,
  stateHistory: boardHistory[],
  activeColour: PieceColour,
  activeMove: number,
  gameOverStatus: number,
}

export interface gameContext {
  matchData: matchData,
  setMatchData: React.Dispatch<React.SetStateAction<matchData>>,
  webSocket: WebSocket | null,
  playerColour: PieceColour,
  isWhiteConnected: boolean,
  isBlackConnected: boolean,
  opponentEventType: OpponentEventType,
  setOpponentEventType: React.Dispatch<React.SetStateAction<OpponentEventType>>,
  millisecondsUntilOpponentTimeout: number | null,
}

export const GameContext = createContext<gameContext | null>(null)

export interface MatchStateHistory {
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
  millisecondsUntilTimeout: number,
}

interface PlayerCodeMessage {
  playerCode: number
}

interface ChessWebSocketMessage {
  messageType: string
  body: OnConnectMessage | OnMoveMessage | ConnectionStatusMessage | PlayerCodeMessage | OpponentEventMessage
}

interface OpponentEventMessage {
  sender: string,
  eventType: string,
}

export enum OpponentEventType {
  None = "none",
  Takeback = "takeback",
  Draw = "draw",
  Rematch = "rematch",
  Disconnect = "disconnect",
  Decline = "decline",
  Resign = "resign",
}

export function GameWrapper({ children, matchID, timeFormatInMilliseconds }: { children: ReactNode, matchID: string, timeFormatInMilliseconds: number }) {
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
  const [millisecondsUntilOpponentTimeout, setMillisecondsUntilOpponentTimeout] = useState<number | null>(null)
  const [opponentEventType, setOpponentEventType] = useState(OpponentEventType.None)
  
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
      
    if (body["isConnected"]) {
      // Should always be opponent
      setMillisecondsUntilOpponentTimeout(null)
    } else {
      setOpponentEventType(OpponentEventType.Disconnect)
    }
  }
  
  function onMoveHandler(body: OnMoveMessage) {
    setOpponentEventType(OpponentEventType.None)
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
  
  function opponentEventHandler(body: OpponentEventMessage) {
    switch (body.eventType) {
  
    case "takeback":
      setOpponentEventType(OpponentEventType.Takeback)
      break
  
    case "draw":
      setOpponentEventType(OpponentEventType.Draw)
      break
  
    case "rematch":
      setOpponentEventType(OpponentEventType.Rematch)
      break
  
    }
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
      case "opponentEvent":
        opponentEventHandler(parsedMsg["body"] as OpponentEventMessage)
        break;
      default:
        console.error("Could not understand message from websocket")
        console.log(message)
      }
    }
  }
  
  return (
    <GameContext.Provider value={{ matchData, setMatchData, webSocket, playerColour, isWhiteConnected, isBlackConnected, opponentEventType, setOpponentEventType, millisecondsUntilOpponentTimeout }}>
      {children}
    </GameContext.Provider>
  )
}