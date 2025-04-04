import React, { useEffect } from "react";
import { useParams, useLocation } from "react-router-dom"
import { GameWrapper } from "./GameContext";
import { ChessBoard } from "./ChessBoard";
import { GameInfoTile } from "./GameInfoTile";

export function MatchRoom() {
  const { matchid } = useParams()
  const location = useLocation();
  const { timeFormatInMilliseconds } = location.state || {};
  const parsedTimeFormatInMilliseconds = parseInt(timeFormatInMilliseconds)

  useEffect(() => {
    console.log("MatchRoom mount")
    return () => {
      console.log("MatchRoom unmount")
    }
  }, [])

  return (
    <GameWrapper matchID={matchid as string} timeFormatInMilliseconds={parsedTimeFormatInMilliseconds}>
      <div className='chessMatch'>
        <ChessBoard resizeable={true} defaultWidth={800} enableClicking={true}/>
        <GameInfoTile />
      </div>
    </GameWrapper>
  )
}