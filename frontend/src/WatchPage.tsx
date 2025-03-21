import React, { useEffect, useState } from "react";
import { LoaderCircle, Swords, Flame, Rabbit, TrainFront, Turtle } from "lucide-react";
import { Link, useSearchParams } from "react-router-dom";
import { FrozenChessBoard } from "./chess/ChessBoard";
import { parseGameStateFromFEN } from "./chess/ChessLogic";

interface SQLNullString {
  String: string
  Valid: boolean
}

interface SQLNullInt64 {
  Int64: number
  Valid: boolean
}

interface matchData {
  matchID: number
  whitePlayerUsername: SQLNullString
  blackPlayerUsername: SQLNullString
  lastMovePiece: SQLNullInt64
  lastMoveMove: SQLNullInt64
  finalFEN: string
  timeFormatInMilliseconds: number
  incrementInMilliseconds: number
  whitePlayerPoints: number
  blackPlayerPoints: number
  averageElo: number
}

function FormatIcon({ timeFormatInMilliseconds, style }: { timeFormatInMilliseconds: number, style?: React.CSSProperties }) {
  const minute = 60_000
  if (timeFormatInMilliseconds < 2 * minute) {
    return <TrainFront style={style}/>
  } else if (timeFormatInMilliseconds < 5 * minute) {
    return <Flame style={style}/>
  } else if (timeFormatInMilliseconds < 20 * minute) {
    return <Rabbit style={style}/>
  } else {
    return <Turtle style={style}/>
  }
}

async function fetchMatches(searchParams: URLSearchParams, signal: AbortSignal) {
  const timeFormat = searchParams.get("timeFormat") || ""
  console.log(`Time format: ${timeFormat}`)
  const url = import.meta.env.VITE_API_GET_PAST_MATCHES_URL + `?timeFormat=${timeFormat}`

  try {
    const response = await fetch(url, {
      signal: signal,
      method: "GET",
    })

    if (response.ok) {
      console.log(response)
      const responseData: matchData[] = await response.json()
      console.log(responseData)
      return responseData
    }

  } catch (e) {
    console.error(e)
  }
}

function getTimeFormatName(timeFormatInMilliseconds: number) {
  const minute = 60_000
  if (timeFormatInMilliseconds < 2 * minute) {
    return "Bullet"
  } else if (timeFormatInMilliseconds < 5 * minute) {
    return "Blitz"
  } else if (timeFormatInMilliseconds < 20 * minute) {
    return "Rapid"
  } else {
    return "Classical"
  }
}

function MatchTile({ matchData, idx }: { matchData: matchData, idx: number }) {
  let outcome = ""
  if (matchData.whitePlayerPoints > matchData.blackPlayerPoints) {
    outcome = "White wins"
  } else if (matchData.whitePlayerPoints < matchData.blackPlayerPoints) {
    outcome = "Black wins"
  } else {
    outcome = "Draw"
  }

  const gameState = parseGameStateFromFEN(matchData.finalFEN)
  let liClassname = ""
  if (idx % 2 == 0) {
    liClassname += "even"
  } else {
    liClassname += "odd"
  }
  liClassname += " matchRow"

    return (
        <li className={liClassname}>
          <Link to={`/matchroom/${matchData.matchID}`} style={{display: "block", position: "absolute", width: "100%", height: "100%", boxSizing: "content-box", zIndex: "3"}}/>
          <div style={{marginTop: "auto", marginBottom: "auto", boxShadow: "2px 2px 2px #000000", height: "inherit", width: "35vh"}}>
            {/* Chessboard, display final position */}
            <FrozenChessBoard board={gameState.board} lastMove={[matchData.lastMovePiece.Int64, matchData.lastMoveMove.Int64]} showLastMove={matchData.lastMovePiece.Valid && matchData.lastMoveMove.Valid}/>
          </div>
          <div style={{display: "grid", gridTemplateRows: "1fr 1fr", marginLeft: "1em", height: "inherit"}}>
            {/* Info, grid 2 rows, top row is format info and date, 2nd row is player Info and victory */}
            <div style={{display: "grid", gridTemplateRows: "1fr 1fr"}}>
              {/* Grid 2 columns, first column is icon for rating, 2nd is info */}
              <div>
                {/* Icon for rating */}
                <FormatIcon timeFormatInMilliseconds={matchData.timeFormatInMilliseconds} style={{float: "left"}}/>
                <span style={{float: "left"}}>{Math.floor(matchData.timeFormatInMilliseconds / 60_000)}+{Math.floor(matchData.incrementInMilliseconds / 1000)} • {getTimeFormatName(matchData.timeFormatInMilliseconds)}</span>
              </div>
              <div style={{display: "grid", gridTemplateRows: "1fr 1fr"}}>
                {/* Info, grid 2 rows, top is rating info bottom is date */}
                <div>
                  <span style={{float: "left"}}>Date</span>
                </div>
              </div>
            </div>
            <div style={{display: "grid", gridTemplateRows: "1fr 1fr"}}>
            {/* Player Info, grid 2 rows, top is player info, bottom is game outcome */}
            <div style={{display: "grid", gridTemplateColumns: "1fr 1fr 1fr"}}>
              {/* Player info, grid 3 columns, 1st is white info, 2nd is vs icon, 3rd is black info */}
              <div>
                {/* White info*/}
                {matchData.whitePlayerUsername.Valid ? matchData.whitePlayerUsername.String : "Anon" }
              </div>
              <div>
                {/* VS icon */}
                <Swords />
              </div>
              <div>
                {/* Black info */}
                {matchData.blackPlayerUsername.Valid ? matchData.blackPlayerUsername.String : "Anon" }
              </div>
            </div>
            <div>
              {outcome}
            </div>
            </div>
          </div>
        </li>
    )
}

export function WatchPage() {
    const [fetchingMatches, setFetchingMatches] = useState(true)
    const [searchParams, _setSearchParams] = useSearchParams()
    const [matchList, setMatchList] = useState<matchData[]>([])

    useEffect(() => {
      const controller = new AbortController()
      const signal = controller.signal
      const getMatchList = async () => {
        const matchList = await fetchMatches(searchParams, signal)
        console.log(matchList)
        setMatchList(matchList || [])
        setFetchingMatches(false)
      }
      getMatchList()
      return () => {
        controller.abort("WatchPage searchParam changed")
      }
    }, [searchParams])

    if (fetchingMatches) {
        return (
            <div>
                <LoaderCircle className="loaderSpin"/>
            </div>
        )
    }

    return (
        <div>
          <ul>
          {matchList.map((matchData, idx) => {
            return (
              <MatchTile key={`match_${matchData.matchID}`} matchData={matchData} idx={idx}/>
            )
          })}
          </ul>
        </div>
    )
}