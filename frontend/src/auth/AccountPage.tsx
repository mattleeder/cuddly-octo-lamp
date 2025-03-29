import React, { useEffect, useRef, useState } from "react"
import { formatTimePassed, PlayerInfoTileData } from "../PlayerInfoTile"
import { Flame, LoaderCircle, Rabbit, TrainFront, Turtle } from "lucide-react"
import { useParams } from "react-router-dom"
import { matchData, MatchTile } from "../WatchPage"
import { PingStatus } from "../chess/GameInfoTile"

enum Page {
  All = "all",
  Bullet = "bullet",
  Blitz = "blitz",
  Rapid = "rapid",
  Classical = "classical",
}

function AccountInfoDisplay({ accountInfo }: { accountInfo: PlayerInfoTileData | undefined }) {

  if (accountInfo === undefined) {
    return (
      <div>No data.</div>
    )
  }

  return (
    <div style={{
      position: "relative",
      width: "70vw",
      backgroundColor: "#ababaa",
      borderRadius: "4px",
      padding: "1em",
      marginLeft: "auto",
      marginRight: "auto",
      
    }}>
      <div style={{display: "flex", flexDirection: "row"}}>
        <div style={{display: "flex", flexDirection: "row", alignItems: "center"}}>
          <PingStatus style={{width: "1em", paddingRight: "0.5em"}} connected={accountInfo.pingStatus}/>
          <span>{accountInfo.username}</span>
        </div>
        <div style={{
            display: "grid",
            gridTemplateColumns: "auto auto",
            columnGap: "1em",
            marginLeft: "auto",
            alignItems: "start",
            textAlign: "left",
          }}>
          <div>
            Joined:
          </div>
          <div>
            {formatTimePassed(Date.now() - accountInfo.joinDate * 1000)} ago
          </div>
          <div>
            Last Seen:
          </div>
          <div>
            {formatTimePassed(Date.now() - accountInfo.lastSeen * 1000)} ago  
          </div>
          <div>
            Number of Games:
          </div>
          <div>
            {accountInfo.numberOfGames}
          </div>
        </div>
      </div>
    </div>
  )
}


function AccountContent({ pageData }: { pageData: matchData[] | undefined }) {

  if (pageData === undefined || pageData === null) {
    return (
      <div>
        No data.
      </div>
    )
  }
  return (
    <div style={{
      marginLeft: "auto",
      marginRight: "auto",
    }}>
      <ul>
        {pageData.map((matchData, idx) => {
          return (
            <MatchTile key={`match_${matchData.matchID}`} matchData={matchData} idx={idx}/>
          )
        })}
      </ul>
    </div>
  )
}

async function fetchPageData(username: string, activePage: Page, pageCache: React.RefObject<Map<Page, matchData[]>>, signal: AbortSignal) {
  let url = import.meta.env.VITE_API_GET_PAST_MATCHES_URL
  let searchParams: [string, string][] = []

  if (username != "") {
    searchParams.push(["username", username])
  }

  if (activePage != Page.All) {
    searchParams.push(["timeFormat", activePage])
  }

  if (searchParams.length > 0) {
    url += "?"
  }
  
  for (let [searchTerm, value] of searchParams) {
    url += `${searchTerm}=${value}&`
  }

  try {
    const response = await fetch(url, {
      signal: signal,
      method: "GET",
    })

    if (response.ok) {
      console.log(response)
      const responseData: matchData[] = await response.json()
      console.log(responseData)
      pageCache.current?.set(activePage, responseData)
      return responseData
    }

  } catch (e) {
    console.error(e)
  }

  return undefined
}

async function fetchUserData(username: string, signal: AbortSignal) {
  const url = import.meta.env.VITE_API_GET_TILE_INFO_URL + `?search=${username}`

  try {
    const response = await fetch(url, {
      signal: signal,
      method: "GET",
    })

    if (response.ok) {
      console.log(response)
      const responseData: PlayerInfoTileData = await response.json()
      console.log(responseData)
      return responseData
    }

  } catch (e) {
    console.error(e)
  }

  return undefined
}

export function AccountPage() {
  const { username } = useParams()
  const [activePage, setActivePage] = useState(Page.All)
  const [pageData, setPageData] = useState<matchData[] | undefined>(undefined)
  const [playerData, setPlayerData] = useState<PlayerInfoTileData | undefined>(undefined)
  const pageCache = useRef(new Map<Page, matchData[]>()) // Should be a ref instead
  const [loadingPlayerData, setLoadingPlayerData] = useState(true)
  const [loadingContent, setLoadingContent] = useState(true)

  // Get user data
  useEffect(() => {
    let ignore = false
    setLoadingPlayerData(true)
    const controller = new AbortController()
    const signal = controller.signal;

    (async() => {
      const playerData = fetchUserData(username || "", signal)
      if (!ignore) {
        setPlayerData(await playerData)
        setLoadingPlayerData(false)
      }}
    )()

    return () => {
      ignore = true
      controller.abort("username changed")
    }
  }, [username])

  // Get Page Content
  useEffect(() => {
    let ignore = false
    setLoadingContent(true)
    const controller = new AbortController()
    const signal = controller.signal;

    (async() => {
      console.log(pageCache)
      const pageData = pageCache.current?.get(activePage) || fetchPageData(username || "", activePage, pageCache, signal)
      if (!ignore) {
        setPageData(await pageData)
        setLoadingContent(false)
      }}
    )()

    return () => {
      ignore = true
      controller.abort("activePage changed")
    }
  }, [username, activePage])

    return (
    <>
      <div style={{
        display: "block",
        position: "fixed",
        left: 0,
        top: "3em",
        textAlign: "left",
        padding: 0,
        margin: 0,
      }}>
        <ul style={{
          display: "block",
          listStyle: "none",
          padding: 0,
          margin: 0,
        }}>
          <li className={`userTimeFormat${activePage == Page.All ? " active" : ""}`} onClick={() => setActivePage(Page.All)}>
            <div>
              <span>All</span>
            </div>
          </li>
          <li className={`userTimeFormat${activePage == Page.Bullet ? " active" : ""}`} onClick={() => setActivePage(Page.Bullet)}>
            <div style={{display: "flex", flexDirection: "column", width: "100%"}}>
              <div style={{display: "flex", justifyItems: "center"}}>
                <span>Bullet</span>
                <TrainFront style={{marginLeft: "auto"}}/>
              </div>
              <div style={{textAlign: "center"}}>
                {playerData?.ratings.bullet || "-"}
              </div>
            </div>
          </li>
          <li className={`userTimeFormat${activePage == Page.Blitz ? " active" : ""}`} onClick={() => setActivePage(Page.Blitz)}>
            <div style={{display: "flex", flexDirection: "column", width: "100%"}}>
              <div style={{display: "flex", justifyItems: "center"}}>
                <span>Blitz</span>
                <Flame style={{marginLeft: "auto"}}/>
              </div>
              <div style={{textAlign: "center"}}>
                {playerData?.ratings.blitz || "-"}
              </div>
            </div>
          </li>
          <li className={`userTimeFormat${activePage == Page.Rapid ? " active" : ""}`} onClick={() => setActivePage(Page.Rapid)}>
            <div style={{display: "flex", flexDirection: "column", width: "100%"}}>
              <div style={{display: "flex", justifyItems: "center"}}>
                <span>Rapid</span>
                <Rabbit style={{marginLeft: "auto"}}/>
              </div>
              <div style={{textAlign: "center"}}>
                {playerData?.ratings.rapid || "-"}
              </div>
            </div>
          </li>
          <li className={`userTimeFormat${activePage == Page.Classical ? " active" : ""}`} onClick={() => setActivePage(Page.Classical)}>
            <div style={{display: "flex", flexDirection: "column", width: "100%"}}>
              <div style={{display: "flex", justifyItems: "center"}}>
                <span>Classical</span>
                <Turtle style={{marginLeft: "auto"}}/>
              </div>
              <div style={{textAlign: "center"}}>
                {playerData?.ratings.classical || "-"}
              </div>
            </div>
          </li>
        </ul>
      </div>
      <div style={{
        display:"flex",
        flexDirection: "column",
        height: "100vh",
        paddingTop: "3em",
        marginLeft: "auto",
        marginRight: "auto",
      }}>
        {loadingPlayerData ? <LoaderCircle className="loaderSpin"/> : <AccountInfoDisplay accountInfo={playerData}/>}
        {loadingContent ? <LoaderCircle className="loaderSpin"/> : <AccountContent pageData={pageData}/>}
      </div>
    </>
    )
}