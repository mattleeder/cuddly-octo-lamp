import React, { useRef } from 'react';
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
  // webSocket: WebSocket | null,
  webSocket: React.RefObject<WebSocket | null>,
  playerColour: PieceColour,
  isWhiteConnected: boolean,
  isBlackConnected: boolean,
  opponentEventType: OpponentEventType,
  setOpponentEventType: React.Dispatch<React.SetStateAction<OpponentEventType>>,
  millisecondsUntilOpponentTimeout: number | null,
  threefoldRepetition: boolean,
  setThreefoldRepetition: React.Dispatch<React.SetStateAction<boolean>>,
  flip: boolean,
  setFlip: React.Dispatch<React.SetStateAction<boolean>>,
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
  ThreefoldRepetition = "threefoldRepetition"
}

function sendPlayerCodeHandler(
  body: PlayerCodeMessage, 
  setPlayerColour: React.Dispatch<React.SetStateAction<PieceColour>>,
) {
  if (body["playerCode"] == 0) {
    setPlayerColour(PieceColour.White)
  } else if (body["playerCode"] == 1) {
    setPlayerColour(PieceColour.Black)
  }
}

function onConnectHandler(
  body: OnConnectMessage, 
  setIsWhiteConnected: React.Dispatch<React.SetStateAction<boolean>>, 
  setIsBlackConnected: React.Dispatch<React.SetStateAction<boolean>>,
  setOpponentEventType: React.Dispatch<React.SetStateAction<OpponentEventType>>,
  matchData: matchData,
  setMatchData: React.Dispatch<React.SetStateAction<matchData>>,
  setThreefoldRepetition: React.Dispatch<React.SetStateAction<boolean>>,
) {
  onMoveHandler(body as OnMoveMessage, setOpponentEventType, matchData, setMatchData, setThreefoldRepetition)
  setIsWhiteConnected(body["whitePlayerConnected"])
  setIsBlackConnected(body["blackPlayerConnected"])
}

function connectionStatusHandler(
  body: ConnectionStatusMessage,
  setIsWhiteConnected: React.Dispatch<React.SetStateAction<boolean>>, 
  setIsBlackConnected: React.Dispatch<React.SetStateAction<boolean>>,
  setMillisecondsUntilOpponentTimeout: React.Dispatch<React.SetStateAction<number | null>>,
  setOpponentEventType: React.Dispatch<React.SetStateAction<OpponentEventType>>,
) {
  if (body["playerColour"] == "white") {
    setIsWhiteConnected(body["isConnected"])
  } else if (body["playerColour"] == "black") {
    setIsBlackConnected(body["isConnected"])
  }
    
  if (body["isConnected"]) {
    // Should always be opponent
    setMillisecondsUntilOpponentTimeout(null)
  } else {
    setMillisecondsUntilOpponentTimeout(body["millisecondsUntilTimeout"])
    // setOpponentEventType(OpponentEventType.Disconnect)
  }
}

function onMoveHandler(
  body: OnMoveMessage,
  setOpponentEventType: React.Dispatch<React.SetStateAction<OpponentEventType>>,
  matchData: matchData,
  setMatchData: React.Dispatch<React.SetStateAction<matchData>>,
  setThreefoldRepetition: React.Dispatch<React.SetStateAction<boolean>>
) {
  setThreefoldRepetition(body["threefoldRepetition"])
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

function opponentEventHandler(body: OpponentEventMessage, setOpponentEventType: React.Dispatch<React.SetStateAction<OpponentEventType>>,) {
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

function readMessage(
  message: unknown,
  setPlayerColour: React.Dispatch<React.SetStateAction<PieceColour>>,
  setIsWhiteConnected: React.Dispatch<React.SetStateAction<boolean>>,
  setIsBlackConnected: React.Dispatch<React.SetStateAction<boolean>>,
  setMillisecondsUntilOpponentTimeout: React.Dispatch<React.SetStateAction<number | null>>,
  setOpponentEventType: React.Dispatch<React.SetStateAction<OpponentEventType>>,
  matchData: matchData,
  setMatchData: React.Dispatch<React.SetStateAction<matchData>>,
  setThreefoldRepetition: React.Dispatch<React.SetStateAction<boolean>>,
) {
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
      sendPlayerCodeHandler(parsedMsg["body"] as PlayerCodeMessage, setPlayerColour)
      break;
    case "onConnect":
      onConnectHandler(parsedMsg["body"] as OnConnectMessage, setIsWhiteConnected, setIsBlackConnected, setOpponentEventType, matchData, setMatchData, setThreefoldRepetition)
      break;
    case "connectionStatus":
      connectionStatusHandler(parsedMsg["body"] as ConnectionStatusMessage, setIsWhiteConnected, setIsBlackConnected, setMillisecondsUntilOpponentTimeout, setOpponentEventType)
      break;
    case "onMove":
      onMoveHandler(parsedMsg["body"] as OnMoveMessage, setOpponentEventType, matchData, setMatchData, setThreefoldRepetition)
      break;
    case "opponentEvent":
      opponentEventHandler(parsedMsg["body"] as OpponentEventMessage, setOpponentEventType)
      break;
    default:
      console.error("Could not understand message from websocket")
      console.log(message)
    }
  }
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
  // const [webSocket, setWebSocket] = useState<WebSocket | null>(null)
  const webSocket = useRef<WebSocket | null>(null)
  const webSocketReconnectTimeout = useRef(1000)
  // const shouldReconnect = useRef(true)
  const [playerColour, setPlayerColour] = useState(PieceColour.Spectator)
  const [isWhiteConnected, setIsWhiteConnected] = useState(false)
  const [isBlackConnected, setIsBlackConnected] = useState(false)
  const [millisecondsUntilOpponentTimeout, setMillisecondsUntilOpponentTimeout] = useState<number | null>(null)
  const [opponentEventType, setOpponentEventType] = useState(OpponentEventType.None)
  const [threefoldRepetition, setThreefoldRepetition] = useState(false)
  const [flip, setFlip] = useState(false)

  useEffect(() => {
    setFlip(playerColour == PieceColour.Black)
  }, [playerColour])

  useEffect(() => {
    const webSocketConnect = () => {
      webSocket.current = new WebSocket(import.meta.env.VITE_API_MATCHROOM_URL + matchID + '/ws')
      webSocket.current.onopen = () => console.log("Websocket connected")
      webSocket.current.onmessage = (event) => readMessage(event.data, setPlayerColour, setIsWhiteConnected, setIsBlackConnected, setMillisecondsUntilOpponentTimeout, setOpponentEventType, matchData, setMatchData, setThreefoldRepetition)
      webSocket.current.onerror = (event) => console.error(event)
      webSocket.current.onclose = () => {
        console.log(`WebSocket closed, attempting reconnect in ${Math.floor(webSocketReconnectTimeout.current / 1000)}s`)
        webSocketConnect()
      }
    }
    webSocketConnect()
    return () => {
      if (webSocket.current == null) {
        return
      }
      webSocket.current.onclose = () => console.log("Closing WebSocket")
      webSocket.current.close()
    }
  }, [matchID])
  
  useEffect(() => {
    console.log("GAMEWRAPPER:")
    console.log(matchData)
  }, [matchData])
  
  return (
    <GameContext.Provider value={{ matchData, setMatchData, webSocket, playerColour, isWhiteConnected, isBlackConnected, opponentEventType, setOpponentEventType, millisecondsUntilOpponentTimeout, threefoldRepetition, setThreefoldRepetition, flip, setFlip }}>
      {children}
    </GameContext.Provider>
  )
}