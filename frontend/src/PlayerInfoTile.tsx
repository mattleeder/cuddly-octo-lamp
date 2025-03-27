import { Flame, Rabbit, TrainFront, Turtle } from 'lucide-react';
import React, { createContext, useEffect, useRef, useState } from 'react';

// Display Ping Status
// Player Name
// Ratings
// Number of games player
// Join date

interface PlayerInfoTilePosition {
  x: number,
  y: number,
}

export interface RatingsObject {
  bullet: number
  blitz: number
  rapid: number
  classical: number
}

export interface PlayerInfoTileData {
  playerID: number
  username: string
  pingStatus: boolean
  joinDate: number
  lastSeen: number
  ratings: RatingsObject
  numberOfGames: number
}

export interface PlayerInfoTileContextInterface {
  spawnPlayerInfoTile: (arg0: string, arg1: React.MouseEvent<HTMLElement, MouseEvent>) => void
  lightFusePlayerInfoTile: (arg0: string, arg1: React.MouseEvent<HTMLElement, MouseEvent>) => void
}

export const PlayerInfoTileContext = createContext<PlayerInfoTileContextInterface>({
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  spawnPlayerInfoTile: (_arg0, _arg1) => {return},
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  lightFusePlayerInfoTile: (_arg0, _arg1) => {return},
})

// let playerTileInfoIndex = 0
// const playerTileInfoArray = [
//   {
//     pingStatus: true,
//     username: 0,
//     displayName: "userOne",
//     ratings: {bullet: 1500, blitz: 1500, rapid: 1500, classical: 1500},
//     numberOfGames: 1500,
//     joinDate: "20-01-2020",
//   },
//   {
//     pingStatus: true,
//     username: 1,
//     displayName: "userTwo",
//     ratings: {bullet: 1400, blitz: 1400, rapid: 1400, classical: 1400},
//     numberOfGames: 1400,
//     joinDate: "21-01-2020",
//   },
//   {
//     pingStatus: true,
//     username: 2,
//     displayName: "userThree",
//     ratings: {bullet: 1300, blitz: 1300, rapid: 1300, classical: 1300},
//     numberOfGames: 1300,
//     joinDate: "22-01-2020",
//   },
// ]

// function fetchPlayerTileData(username: number): PlayerInfoTileData {
//   console.log(`Fetching tile data: ${username}`)
//   playerTileInfoIndex = (playerTileInfoIndex + 1) % playerTileInfoArray.length
//   return playerTileInfoArray[playerTileInfoIndex]
// }

async function fetchPlayerTileData(username: string) {
  console.log(`Search string: ${username}`)
  const url = import.meta.env.VITE_API_GET_TILE_INFO_URL + `?search=${username}`

  try {
    const response = await fetch(url, {
      method: "GET",
    })

    if (response.ok) {
      const responseData = await response.json()
      console.log(responseData)
      return responseData
    }

  } catch (e) {
    console.error(e)
  }
}

function makePlayerInfoTile(
  setActive: React.Dispatch<React.SetStateAction<boolean>>,
  setUsername: React.Dispatch<React.SetStateAction<string | null>>,
  setPlayerData: React.Dispatch<React.SetStateAction<PlayerInfoTileData | null>>,
  setPosition: React.Dispatch<React.SetStateAction<PlayerInfoTilePosition>>,
  username: string,
  playerData: PlayerInfoTileData | null,
  position: PlayerInfoTilePosition,
) {
  setActive(false)
  setUsername(username)
  // if (playerData === null) {
  //   playerData = await fetchPlayerTileData(username)
  // }
  // setPlayerData(playerData)
  setPosition(position)
  setActive(true)
  console.log(`Set active, newUsername: ${username}, newPosition.x: ${position.x}, newPosition.y: ${position.y}`)
}

function queuePlayerInfoTile(
  queuedUsername: React.RefObject<string | null>,
  queuedPlayerData: React.RefObject<PlayerInfoTileData | null>,
  queuedPosition: React.RefObject<PlayerInfoTilePosition>,
  username: string,
  position: PlayerInfoTilePosition,
) {
  // const playerData = await fetchPlayerTileData(username)
  queuedUsername.current = username
  // queuedPlayerData.current = playerData
  queuedPosition.current = position
}

function spawnHandler(
  makePlayerInfoTile: (username: string, playerData: PlayerInfoTileData | null, position: PlayerInfoTilePosition) => void,
  queuePlayerInfoTile: (username: string, position: PlayerInfoTilePosition) => void,
  setFuseActive: React.Dispatch<React.SetStateAction<boolean>>,
  active: boolean,
  activeUsername: string | null,
  activePosition: PlayerInfoTilePosition,
  newUsername: string,
  newPosition: PlayerInfoTilePosition,
) {
  console.log("In spawn handler")
  console.log(`Active: ${active}`)
  // If none active, make new immediately
  if (active == false) {
    console.log("Making")
    queuePlayerInfoTile(newUsername, newPosition)
    setFuseActive(true)
    return
  }

  // If active check if username && position are same, if so no change
  if (newUsername == activeUsername && newPosition.x == activePosition.x && newPosition.y == activePosition.y) {
    console.log("Clear Fuse")
    setFuseActive(false)
    return
  }

  // If different, set queued username, queued Position and fuse
  console.log("Queuing")
  queuePlayerInfoTile(newUsername, newPosition)
  setFuseActive(true)
}

export function spawnPlayerInfoTile(
  spawnHandler: (active: boolean, activeUsername: string | null, activePosition: PlayerInfoTilePosition, username: string, position: PlayerInfoTilePosition) => void,
  active: boolean,
  activeUsername: string | null,
  activePosition: PlayerInfoTilePosition,
  username: string,
  position: PlayerInfoTilePosition
) {
  spawnHandler(active, activeUsername, activePosition, username, position)
}

function fuseHandler(
  setFuseActive: React.Dispatch<React.SetStateAction<boolean>>,
  queuedUsername: React.RefObject<string | null>,
  queuedPlayerData: React.RefObject<PlayerInfoTileData | null>,
  active: boolean,
  activeUsername: string | null,
  activePosition: PlayerInfoTilePosition,
  newUsername: string,
  newPosition: PlayerInfoTilePosition,
) {
  // If none active do nothing
  if (!active) {
    setFuseActive(false)
    return
  }

  console.log(`activeUsername: ${activeUsername}, activePosition.x: ${activePosition.x}, activePosition.y:${activePosition.y}`)

  // If active check if username && position are same, if so light fuse
  if (newUsername == activeUsername && newPosition.x == activePosition.x && newPosition.y == activePosition.y) {
    console.log("Light fuse")
    setFuseActive(true)
    return
  }

  // Else clear queue
  console.log("Clear queue")
  setFuseActive(false)
  queuedUsername.current = null
  queuedPlayerData.current = null
}

export function lightFusePlayerInfoTile(
  fuseHandler: (username: string, position: PlayerInfoTilePosition) => void,
  username: string,
  position: PlayerInfoTilePosition
) {
  fuseHandler(username, position)
}

async function updatePlayerData(username: string, setPlayerData: React.Dispatch<React.SetStateAction<PlayerInfoTileData | null>>) {
  const playerData = await fetchPlayerTileData(username)
  // const playerData = fetchPlayerTileData(username)
  setPlayerData(playerData)
}

function clearPlayerData(setUsername: React.Dispatch<React.SetStateAction<string | null>>, setPlayerData: React.Dispatch<React.SetStateAction<PlayerInfoTileData | null>>) {
  // setUsername(null)
  setPlayerData(null)
}

function destroyTile(
  makePlayerInfoTile: (username: string, playerData: PlayerInfoTileData | null, position: PlayerInfoTilePosition) => void,
  setActive: React.Dispatch<React.SetStateAction<boolean>>,
  setFuseActive: React.Dispatch<React.SetStateAction<boolean>>,
  queuedUsername: React.RefObject<string | null>,
  queuedPosition: React.RefObject<PlayerInfoTilePosition>,
  queuedPlayerData: React.RefObject<PlayerInfoTileData | null>,
) {
  console.log("Destroying")
  setActive(false)
  setFuseActive(false)
  if (queuedUsername.current != null) {
    makePlayerInfoTile(queuedUsername.current, null, queuedPosition.current)

    queuedUsername.current = null
    queuedPosition.current = {x: 0, y: 0}
    queuedPlayerData.current = null
  }
}

function getPositionFromMouseEvent(event: React.MouseEvent<Element, MouseEvent>): PlayerInfoTilePosition {
  const element = event.target as HTMLElement
  const rect = element.getBoundingClientRect()

  return {
    x: rect.left,
    y: rect.top + rect.height,
  }
}

export function formatTimePassed(millisecondsSince: number) {
  const intervals: [string, number][] = [
    [" year",   31_536_000_000],
    [" month",  2_592_000_000],     // 1 Month = 30 Days
    [" week",   604_800_000],
    [" day",    86_400_000],
    [" hour",   3_600_000],
    [" minute", 60_000],
    [" second", 1000],
  ]

  let output = ""

  for (const [text, milliseconds] of intervals) {
    if (millisecondsSince >= milliseconds) {
      const multiplier = Math.floor(millisecondsSince / milliseconds)
      output += multiplier
      output += text
      if (multiplier) {
        output += "s"
      }
      return output
    }
  }

  return "less than a second"
}

export function PlayerInfoTile({ children }: { children: React.ReactNode }) {
  const [active, setActive] = useState(false)
  const [fuseActive, setFuseActive] = useState(false)
  const [username, setUsername] = useState<string | null>(null)
  const [playerData, setPlayerData] = useState<PlayerInfoTileData | null>(null)
  const [position, setPosition] = useState<PlayerInfoTilePosition>({x: 0, y: 0})

  const queuedUsername = useRef<string | null>(null)
  const queuedPlayerData = useRef<PlayerInfoTileData | null>(null)
  const queuedPosition = useRef<PlayerInfoTilePosition>({x: 0, y: 0})


  const FUSE_TIMER_MS = 500

  const makePlayerInfoClosure = (newUsername: string, newPlayerData: PlayerInfoTileData | null, newPosition: PlayerInfoTilePosition) => {
    makePlayerInfoTile(setActive, setUsername, setPlayerData, setPosition, newUsername, newPlayerData, newPosition)
  }

  const queuePlayerInfoClosure = (newUsername: string, newPosition: PlayerInfoTilePosition) => {
    queuePlayerInfoTile(queuedUsername, queuedPlayerData, queuedPosition, newUsername, newPosition)
  }

  const spawnHandlerClosure = (newUsername: string, newPosition: PlayerInfoTilePosition) => {
    spawnHandler(makePlayerInfoClosure, queuePlayerInfoClosure, setFuseActive, active, username, position, newUsername, newPosition)
  }

  const fuseHandlerClosure = (newUsername: string, newPosition: PlayerInfoTilePosition) => {
    fuseHandler(setFuseActive, queuedUsername, queuedPlayerData, active, username, position, newUsername, newPosition)
  }

  const spawnPlayerInfoTile = (newUsername: string, event: React.MouseEvent) => {
    const newPosition = getPositionFromMouseEvent(event)
    console.log(`Spawn info tile, active: ${active}, newUsername: ${newUsername}, position.x:${newPosition.x}, position.y:${newPosition.y}`)
    spawnHandlerClosure(newUsername, newPosition)
  }

  const lightFusePlayerInfoTile = (newUsername: string, event: React.MouseEvent) => {
    const newPosition = getPositionFromMouseEvent(event)
    console.log(`Light tile fuse, active: ${active}, newUsername: ${newUsername}, position:${newPosition.x}, position.y:${newPosition.y}`)
    fuseHandlerClosure(newUsername, newPosition)
  }

  const destroyTileClosure = () => {
    destroyTile(makePlayerInfoClosure, setActive, setFuseActive, queuedUsername, queuedPosition, queuedPlayerData)
  }

  const [tileContext, setTileContext] = useState<PlayerInfoTileContextInterface>({
    spawnPlayerInfoTile,
    lightFusePlayerInfoTile,
  })

  useEffect(() => {
    console.log(`useEffect [username]: ${username}`)
    if (username != null) {
      updatePlayerData(username, setPlayerData)
    } else {
      clearPlayerData(setUsername, setPlayerData)
    }
    return () => {
      clearPlayerData(setUsername, setPlayerData)
    }
  }, [username])

  useEffect(() => {
    let timeout = null
    if (fuseActive) {
      timeout = setTimeout(() => {
        destroyTileClosure()
      }, FUSE_TIMER_MS)
    }
    return () => {
      if (timeout != null) {
        clearTimeout(timeout)
      }
    }
  }, [fuseActive])

  useEffect(() => {
    console.log(`useEffect: ${active}`)
  }, [active])

  useEffect(() => {
    console.log("Updating context")
    setTileContext({
      spawnPlayerInfoTile: spawnPlayerInfoTile,
      lightFusePlayerInfoTile: lightFusePlayerInfoTile,
    })
  }, [active, username, fuseActive, playerData, position, queuedUsername, queuedPlayerData, queuedPosition])

  let timeSince: number | null = null
  if (playerData) {
    const joinDate = playerData.joinDate * 1000
    const now = Date.now()
    timeSince = now - joinDate
  }


  return (
    <PlayerInfoTileContext.Provider value={tileContext}>
      {children}
      {active ?
      
        <div 
          className="playerInfoTile" 
          style={{transform: `translate(${position.x}px, ${position.y}px)`}}
          onMouseEnter={() => setFuseActive(false)}
          onMouseLeave={() => setFuseActive(true)}
        >
          <div className="Name&Ping">
            {`Player Name: ${playerData?.username}`}
          </div>

          <div className="playerInfoTileRatings">
            <div>
              <TrainFront />
              {playerData?.ratings.bullet}
            </div>

            <div>
              <Flame />
              {playerData?.ratings.blitz}
            </div>

            <div>
              <Rabbit />
              {playerData?.ratings.rapid}
            </div>

            <div>
              <Turtle />
              {playerData?.ratings.classical}
            </div>
          </div>

          <div className="Games&JoinDate">
            <div style={{float:"left"}}>
              {`${playerData?.numberOfGames} games`}
            </div>
            <div style={{float:"right"}}>
              {timeSince != null ? `Joined ${formatTimePassed(timeSince)} ago` : "Unknown"}
            </div>
          </div>
        </div>

        :

        <></>
      }
    </PlayerInfoTileContext.Provider>
  )
}