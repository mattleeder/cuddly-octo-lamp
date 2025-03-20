import { LoaderCircle } from "lucide-react";
import React, { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

interface matchData {
  matchID: number
  whitePlayerUsername: string
  blackPlayerUsername: string
  lastMovePiece: number | null
  lastMoveMove: number | null
  finalFEN: string
  timeFormatInMilliseconds: number
  incrementInMilliseconds: number
  whitePlayerPoints: number
  blackPlayerPoints: number
  averageElo: number
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
      const responseData = await response.json()
      return responseData
    }

  } catch (e) {
    console.error(e)
  }
}

function MatchTile({ matchData }: { matchData: matchData }) {
    return (
        <div>
          {matchData.matchID}
        </div>
    )
}

export function WatchPage() {
    const [fetchingMatches, setFetchingMatches] = useState(true)
    const [searchParams, _setSearchParams] = useSearchParams()
    const [matchList, setMatchList] = useState([])

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
          {matchList.map((matchData) => {
            return (
              <MatchTile matchData={matchData} />
            )
          })}
        </div>
    )
}