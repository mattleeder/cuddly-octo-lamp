import React, { useEffect, useState } from "react"
import { PlayerInfoTileData } from "../PlayerInfoTile"
import { LoaderCircle } from "lucide-react"
import { useParams } from "react-router-dom"
import { matchData, MatchTile } from "../WatchPage"

enum Page {
  All = "All",
  Bullet = "Bullet",
  Blitz = "Blitz",
  Rapid = "Rapid",
  Classical = "Classical",
}

interface PageData {
  matchList: matchData[]
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
      height: "20vh",
      backgroundColor: "#ababaa",
      borderRadius: "4px",
      
    }}>
      {accountInfo.username}
    </div>
  )
}


function AccountContent({ pageData }: { pageData: PageData | undefined }) {

  if (pageData === undefined) {
    return (
      <div>
        No data.
      </div>
    )
  }
  return (
    <div>
      <ul>
        {pageData.matchList.map((matchData, idx) => {
          return (
            <MatchTile key={`match_${matchData.matchID}`} matchData={matchData} idx={idx}/>
          )
        })}
      </ul>
    </div>
  )
}

async function fetchPageData(username: string, activePage: Page, setPageCache: React.Dispatch<React.SetStateAction<Map<Page, PageData>>>, signal: AbortSignal) {
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
    url += `${searchTerm}=${value}`
  }

  try {
    const response = await fetch(url, {
      signal: signal,
      method: "GET",
    })

    if (response.ok) {
      console.log(response)
      const responseData: PageData = await response.json()
      console.log(responseData)
      setPageCache((currentCache) => {
        return {
          ...currentCache,
          activePage: responseData
        }
      })
      return responseData
    }

  } catch (e) {
    console.error(e)
  }

  return undefined
}

const exampleAccountInfo: PlayerInfoTileData = {
  playerID: 123,
  username: "Test",
  pingStatus: true,
  joinDate: 1,
  lastSeen: 2,
  ratings: {
    bullet: 1500,
    blitz: 1500,
    rapid: 1500,
    classical: 1500,
  },
  numberOfGames: 2,
}

async function fetchUserData(username: string, signal: AbortSignal) {
  const url = import.meta.env.VITE_API_GET_TILE_INFO_URL + `?search=${username}`
  return exampleAccountInfo

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
  const [pageData, setPageData] = useState<PageData | undefined>(undefined)
  const [playerData, setPlayerData] = useState<PlayerInfoTileData | undefined>(undefined)
  const [pageCache, setPageCache] = useState(new Map<Page, PageData>()) // Should be a ref instead
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
      const pageData = pageCache.get(activePage) || fetchPageData(username || "", activePage, setPageCache, signal)
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
    <div>
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
            All
          </li>
          <li className={`userTimeFormat${activePage == Page.Bullet ? " active" : ""}`} onClick={() => setActivePage(Page.Bullet)}>
            Bullet
          </li>
          <li className={`userTimeFormat${activePage == Page.Blitz ? " active" : ""}`} onClick={() => setActivePage(Page.Blitz)}>
            Blitz
          </li>
          <li className={`userTimeFormat${activePage == Page.Rapid ? " active" : ""}`} onClick={() => setActivePage(Page.Rapid)}>
            Rapid
          </li>
          <li className={`userTimeFormat${activePage == Page.Classical ? " active" : ""}`} onClick={() => setActivePage(Page.Classical)}>
            Classical
          </li>
        </ul>
      </div>
      {loadingPlayerData ? <LoaderCircle className="loaderSpin"/> : <AccountInfoDisplay accountInfo={playerData}/>}
      {loadingContent ? <LoaderCircle className="loaderSpin"/> : <AccountContent pageData={pageData}/>}
    </div>
    )
}