import React, { useContext, useEffect, useState } from "react";
import { GameContext, GameWrapper } from "./GameContext";
import { ChessBoard } from "./ChessBoard";


function sleep(delayMs: number){
  return new Promise((resolve) => setTimeout(resolve, delayMs));
}

async function setHighestEloMatchID(setMatchID: (matchID: string) => void) {
  let initialRetryDelayMs = 1000

  while (true) {

    try {
      const response = await fetch(import.meta.env.VITE_API_FETCH_HIGHEST_ELO_MATCH_URL, {
        "method": "GET",
        "mode": "cors",
      })

      console.log(response)
      
      if (!response.ok) {
        throw new Error(`Response status: ${response.status}`)
      }
      
      const data = await response.json()
      setMatchID(data["matchID"])
      return
    }
      
    catch (error: unknown) {
      if (error instanceof Error) {
        console.error(error.message)
      } else {
        console.error(error)
      }
    }

    await sleep(initialRetryDelayMs)
    initialRetryDelayMs *= 2
  }
}

function GameOverListener({ callbackFunction }: { callbackFunction: () => void }) {
  const gameContext = useContext(GameContext)

  if (gameContext?.matchData.gameOverStatus !== undefined && gameContext?.matchData.gameOverStatus != 0) {
    callbackFunction()
  }

  return (
    <></>
  )

}

export function HomepageMatch() {
  const [matchID, setMatchID] = useState<undefined | string>(undefined)
  const [loading, setLoading] = useState(true)
  const parsedTimeFormatInMilliseconds = 0

  const onMatchEnd = () => {
    setLoading(true)
  }

  useEffect(() => {
    if (matchID !== undefined) {
      setLoading(false)
    }
  }, [matchID])

  if (loading) {
    setHighestEloMatchID(setMatchID)
    return (
      <div className='chessboard' />
    )
  }

  return (
    <GameWrapper matchID={matchID as string} timeFormatInMilliseconds={parsedTimeFormatInMilliseconds}>
      <div className='chessMatch'>
        <GameOverListener callbackFunction={onMatchEnd} />
        <ChessBoard resizeable={false} defaultWidth={300} chessboardContainerStyles={{transform: "translate(-800px, 200px)"}} enableClicking={false}/>
      </div>
    </GameWrapper>
  )
}