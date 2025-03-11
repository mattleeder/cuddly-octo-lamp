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

interface RatingsObject {
  bullet: number
  blitz: number
  rapid: number
  classical: number
}

interface PlayerInfoTileData {
  pingStatus: boolean
  playerID: number
  displayName: string
  ratings: RatingsObject
  numberOfGames: number
  joinDate: string
}

export interface PlayerInfoTileContextInterface {
  spawnPlayerInfoTile: (arg0: number, arg1: React.MouseEvent<HTMLElement, MouseEvent>) => void
  lightFusePlayerInfoTile: (arg0: number, arg1: React.MouseEvent<HTMLElement, MouseEvent>) => void
}

export const PlayerInfoTileContext = createContext<PlayerInfoTileContextInterface>({
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  spawnPlayerInfoTile: (_arg0, _arg1) => {return},
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  lightFusePlayerInfoTile: (_arg0, _arg1) => {return},
})

let playerTileInfoIndex = 0
const playerTileInfoArray = [
  {
    pingStatus: true,
    playerID: 0,
    displayName: "userOne",
    ratings: {bullet: 1500, blitz: 1500, rapid: 1500, classical: 1500},
    numberOfGames: 1500,
    joinDate: "20-01-2020",
  },
  {
    pingStatus: true,
    playerID: 1,
    displayName: "userTwo",
    ratings: {bullet: 1400, blitz: 1400, rapid: 1400, classical: 1400},
    numberOfGames: 1400,
    joinDate: "21-01-2020",
  },
  {
    pingStatus: true,
    playerID: 2,
    displayName: "userThree",
    ratings: {bullet: 1300, blitz: 1300, rapid: 1300, classical: 1300},
    numberOfGames: 1300,
    joinDate: "22-01-2020",
  },
]

function fetchPlayerTileData(playerID: number): PlayerInfoTileData {
  console.log(`Fetching tile data: ${playerID}`)
  playerTileInfoIndex = (playerTileInfoIndex + 1) % playerTileInfoArray.length
  return playerTileInfoArray[playerTileInfoIndex]
}

function makePlayerInfoTile(
  setActive: React.Dispatch<React.SetStateAction<boolean>>,
  setPlayerID: React.Dispatch<React.SetStateAction<number | null>>,
  setPlayerData: React.Dispatch<React.SetStateAction<PlayerInfoTileData | null>>,
  setPosition: React.Dispatch<React.SetStateAction<PlayerInfoTilePosition>>,
  playerID: number,
  playerData: PlayerInfoTileData | null,
  position: PlayerInfoTilePosition,
) {
  setActive(false)
  setPlayerID(playerID)
  // if (playerData === null) {
  //   playerData = await fetchPlayerTileData(playerID)
  // }
  // setPlayerData(playerData)
  setPosition(position)
  setActive(true)
  console.log(`Set active, newPlayerID: ${playerID}, newPosition.x: ${position.x}, newPosition.y: ${position.y}`)
}

function queuePlayerInfoTile(
  queuedPlayerID: React.RefObject<number | null>,
  queuedPlayerData: React.RefObject<PlayerInfoTileData | null>,
  queuedPosition: React.RefObject<PlayerInfoTilePosition>,
  playerID: number,
  position: PlayerInfoTilePosition,
) {
  // const playerData = await fetchPlayerTileData(playerID)
  queuedPlayerID.current = playerID
  // queuedPlayerData.current = playerData
  queuedPosition.current = position
}

function spawnHandler(
  makePlayerInfoTile: (playerID: number, playerData: PlayerInfoTileData | null, position: PlayerInfoTilePosition) => void,
  queuePlayerInfoTile: (playerID: number, position: PlayerInfoTilePosition) => void,
  setFuseActive: React.Dispatch<React.SetStateAction<boolean>>,
  active: boolean,
  activePlayerID: number | null,
  activePosition: PlayerInfoTilePosition,
  newPlayerID: number,
  newPosition: PlayerInfoTilePosition,
) {
  console.log("In spawn handler")
  console.log(`Active: ${active}`)
  // If none active, make new immediately
  if (active == false) {
    console.log("Making")
    queuePlayerInfoTile(newPlayerID, newPosition)
    setFuseActive(true)
    return
  }

  // If active check if playerID && position are same, if so no change
  if (newPlayerID == activePlayerID && newPosition.x == activePosition.x && newPosition.y == activePosition.y) {
    console.log("Clear Fuse")
    setFuseActive(false)
    return
  }

  // If different, set queued playerID, queued Position and fuse
  console.log("Queuing")
  queuePlayerInfoTile(newPlayerID, newPosition)
  setFuseActive(true)
}

export function spawnPlayerInfoTile(
  spawnHandler: (active: boolean, activePlayerID: number | null, activePosition: PlayerInfoTilePosition, playerID: number, position: PlayerInfoTilePosition) => void,
  active: boolean,
  activePlayerID: number | null,
  activePosition: PlayerInfoTilePosition,
  playerID: number,
  position: PlayerInfoTilePosition
) {
  spawnHandler(active, activePlayerID, activePosition, playerID, position)
}

function fuseHandler(
  setFuseActive: React.Dispatch<React.SetStateAction<boolean>>,
  queuedPlayerID: React.RefObject<number | null>,
  queuedPlayerData: React.RefObject<PlayerInfoTileData | null>,
  active: boolean,
  activePlayerID: number | null,
  activePosition: PlayerInfoTilePosition,
  newPlayerID: number,
  newPosition: PlayerInfoTilePosition,
) {
  // If none active do nothing
  if (!active) {
    return
  }

  console.log(`activePlayerID: ${activePlayerID}, activePosition.x: ${activePosition.x}, activePosition.y:${activePosition.y}`)

  // If active check if playerID && position are same, if so light fuse
  if (newPlayerID == activePlayerID && newPosition.x == activePosition.x && newPosition.y == activePosition.y) {
    console.log("Light fuse")
    setFuseActive(true)
    return
  }

  // Else clear queue
  console.log("Clear queue")
  queuedPlayerID.current = null
  queuedPlayerData.current = null
}

export function lightFusePlayerInfoTile(
  fuseHandler: (playerID: number, position: PlayerInfoTilePosition) => void,
  playerID: number,
  position: PlayerInfoTilePosition
) {
  fuseHandler(playerID, position)
}

function updatePlayerData(playerID: number, setPlayerData: React.Dispatch<React.SetStateAction<PlayerInfoTileData | null>>) {
  // const playerData = await fetchPlayerTileData(playerID)
  const playerData = fetchPlayerTileData(playerID)
  setPlayerData(playerData)
}

function clearPlayerData(setPlayerID: React.Dispatch<React.SetStateAction<number | null>>, setPlayerData: React.Dispatch<React.SetStateAction<PlayerInfoTileData | null>>) {
  // setPlayerID(null)
  setPlayerData(null)
}

function destroyTile(
  makePlayerInfoTile: (playerID: number, playerData: PlayerInfoTileData | null, position: PlayerInfoTilePosition) => void,
  setActive: React.Dispatch<React.SetStateAction<boolean>>,
  setFuseActive: React.Dispatch<React.SetStateAction<boolean>>,
  queuedPlayerID: React.RefObject<number | null>,
  queuedPosition: React.RefObject<PlayerInfoTilePosition>,
  queuedPlayerData: React.RefObject<PlayerInfoTileData | null>,
) {
  console.log("Destroying")
  setActive(false)
  setFuseActive(false)
  if (queuedPlayerID.current != null) {
    makePlayerInfoTile(queuedPlayerID.current, null, queuedPosition.current)

    queuedPlayerID.current = null
    queuedPosition.current = {x: 0, y: 0}
    queuedPlayerData.current = null
  }
}

function getPositionFromMouseEvent(event: React.MouseEvent<HTMLElement, MouseEvent>): PlayerInfoTilePosition {
  const element = event.target
  const rect = element.getBoundingClientRect()

  return {
    x: rect.left,
    y: rect.top + rect.height,
  }
}

export function PlayerInfoTile({ children }: { children: React.ReactNode }) {
  const [active, setActive] = useState(false)
  const [fuseActive, setFuseActive] = useState(false)
  const [playerID, setPlayerID] = useState<number | null>(null)
  const [playerData, setPlayerData] = useState<PlayerInfoTileData | null>(null)
  const [position, setPosition] = useState<PlayerInfoTilePosition>({x: 0, y: 0})

  const queuedPlayerID = useRef<number | null>(null)
  const queuedPlayerData = useRef<PlayerInfoTileData | null>(null)
  const queuedPosition = useRef<PlayerInfoTilePosition>({x: 0, y: 0})


  const FUSE_TIMER_MS = 500

  const makePlayerInfoClosure = (newPlayerID: number, newPlayerData: PlayerInfoTileData | null, newPosition: PlayerInfoTilePosition) => {
    makePlayerInfoTile(setActive, setPlayerID, setPlayerData, setPosition, newPlayerID, newPlayerData, newPosition)
  }

  const queuePlayerInfoClosure = (newPlayerID: number, newPosition: PlayerInfoTilePosition) => {
    queuePlayerInfoTile(queuedPlayerID, queuedPlayerData, queuedPosition, newPlayerID, newPosition)
  }

  const spawnHandlerClosure = (newPlayerID: number, newPosition: PlayerInfoTilePosition) => {
    spawnHandler(makePlayerInfoClosure, queuePlayerInfoClosure, setFuseActive, active, playerID, position, newPlayerID, newPosition)
  }

  const fuseHandlerClosure = (newPlayerID: number, newPosition: PlayerInfoTilePosition) => {
    fuseHandler(setFuseActive, queuedPlayerID, queuedPlayerData, active, playerID, position, newPlayerID, newPosition)
  }

  const spawnPlayerInfoTile = (newPlayerID: number, event: React.MouseEvent) => {
    const newPosition = getPositionFromMouseEvent(event)
    console.log(`Spawn info tile, active: ${active}, newPlayerID: ${newPlayerID}, position.x:${newPosition.x}, position.y:${newPosition.y}`)
    spawnHandlerClosure(newPlayerID, newPosition)
  }

  const lightFusePlayerInfoTile = (newPlayerID: number, event: React.MouseEvent) => {
    const newPosition = getPositionFromMouseEvent(event)
    console.log(`Light tile fuse, active: ${active}, newPlayerID: ${newPlayerID}, position:${newPosition.x}, position.y:${newPosition.y}`)
    fuseHandlerClosure(newPlayerID, newPosition)
  }

  const destroyTileClosure = () => {
    destroyTile(makePlayerInfoClosure, setActive, setFuseActive, queuedPlayerID, queuedPosition, queuedPlayerData)
  }

  const [tileContext, setTileContext] = useState<PlayerInfoTileContextInterface>({
    spawnPlayerInfoTile,
    lightFusePlayerInfoTile,
  })

  useEffect(() => {
    console.log(`useEffect [playerID]: ${playerID}`)
    if (playerID != null) {
      updatePlayerData(playerID, setPlayerData)
    } else {
      clearPlayerData(setPlayerID, setPlayerData)
    }
    return () => {
      clearPlayerData(setPlayerID, setPlayerData)
    }
  }, [playerID])

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
  }, [active, playerID, fuseActive, playerData, position, queuedPlayerID, queuedPlayerData, queuedPosition])

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
            {`Player Name: ${playerData?.displayName}`}
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
              {"100 Games"}
            </div>
            <div style={{float:"right"}}>
              {"Joined This Long Ago"}
            </div>
          </div>
        </div>

        :

        <></>
      }
    </PlayerInfoTileContext.Provider>
  )
}