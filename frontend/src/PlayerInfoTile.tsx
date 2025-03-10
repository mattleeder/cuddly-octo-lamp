import React, { createContext, useEffect, useState } from 'react';

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

interface PlayerInfoTileContextInterface {
  spawnPlayerInfoTile: (arg0: number, arg1: PlayerInfoTilePosition) => void
  lightFusePlayerInfoTile: (arg0: number, arg1: PlayerInfoTilePosition) => void
}

export const PlayerInfoTileContext = createContext<PlayerInfoTileContextInterface>({
  spawnPlayerInfoTile: (_arg0, _arg1) => {return},
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
  console.log("Set active")
}

function queuePlayerInfoTile(
  setQueuedPlayerID: React.Dispatch<React.SetStateAction<number | null>>,
  setQueuedPlayerData: React.Dispatch<React.SetStateAction<PlayerInfoTileData | null>>,
  setQueuedPosition: React.Dispatch<React.SetStateAction<PlayerInfoTilePosition>>,
  playerID: number,
  position: PlayerInfoTilePosition,
) {
  // const playerData = await fetchPlayerTileData(playerID)
  setQueuedPlayerID(playerID)
  // setQueuedPlayerData(playerData)
  setQueuedPosition(position)
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
    makePlayerInfoTile(newPlayerID, null, newPosition)
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
  setQueuedPlayerID: React.Dispatch<React.SetStateAction<number | null>>,
  setQueuedPlayerData: React.Dispatch<React.SetStateAction<PlayerInfoTileData | null>>,
  active: boolean,
  activePlayerID: number | null,
  activePosition: PlayerInfoTilePosition,
  newPlayerID: number,
  newPosition: PlayerInfoTilePosition,
) {
  console.log("In fuse handler")
  console.log(`Active: ${active}`)
  // If none active do nothing
  if (!active) {
    return
  }

  // If active check if playerID && position are same, if so light fuse
  if (newPlayerID == activePlayerID && newPosition.x == activePosition.x && newPosition.y == activePosition.y) {
    setFuseActive(true)
    return
  }

  // Else clear queue
  setQueuedPlayerID(null)
  setQueuedPlayerData(null)
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
  setPlayerID(null)
  setPlayerData(null)
}

function destroyTile(
  makePlayerInfoTile: (playerID: number, playerData: PlayerInfoTileData | null, position: PlayerInfoTilePosition) => void,
  setActive: React.Dispatch<React.SetStateAction<boolean>>,
  setQueuedPlayerID: React.Dispatch<React.SetStateAction<number | null>>,
  setQueuedPosition: React.Dispatch<React.SetStateAction<PlayerInfoTilePosition>>,
  setQueuedPlayerData: React.Dispatch<React.SetStateAction<PlayerInfoTileData | null>>,
  queuedPlayerID: number | null,
  queuedPosition: PlayerInfoTilePosition,
) {
  console.log("Destroying")
  setActive(false)
  if (queuedPlayerID != null) {
    makePlayerInfoTile(queuedPlayerID, null, queuedPosition)

    setQueuedPlayerID(null)
    setQueuedPosition({x: 0, y: 0})
    setQueuedPlayerData(null)
  }
}

export function PlayerInfoTile({ children }: { children: React.ReactNode }) {
  const [active, setActive] = useState(false)
  const [fuseActive, setFuseActive] = useState(false)
  const [playerID, setPlayerID] = useState<number | null>(null)
  const [playerData, setPlayerData] = useState<PlayerInfoTileData | null>(null)
  const [position, setPosition] = useState<PlayerInfoTilePosition>({x: 0, y: 0})

  const [queuedPlayerID, setQueuedPlayerID] = useState<number | null>(null)
  const [queuedPlayerData, setQueuedPlayerData] = useState<PlayerInfoTileData | null>(null)
  const [queuedPosition, setQueuedPosition] = useState<PlayerInfoTilePosition>({x: 0, y: 0})


  const FUSE_TIMER_MS = 500

  const makePlayerInfoClosure = (newPlayerID: number, newPlayerData: PlayerInfoTileData | null, newPosition: PlayerInfoTilePosition) => {
    makePlayerInfoTile(setActive, setPlayerID, setPlayerData, setPosition, newPlayerID, newPlayerData, newPosition)
  }

  const queuePlayerInfoClosure = (newPlayerID: number, newPosition: PlayerInfoTilePosition) => {
    queuePlayerInfoTile(setQueuedPlayerID, setQueuedPlayerData, setQueuedPosition, newPlayerID, newPosition)
  }

  const spawnHandlerClosure = (newPlayerID: number, newPosition: PlayerInfoTilePosition) => {
    console.log("In spawn handler closure")
    console.log(`Active: ${active}`)
    spawnHandler(makePlayerInfoClosure, queuePlayerInfoClosure, setFuseActive, active, playerID, position, newPlayerID, newPosition)
  }

  const fuseHandlerClosure = (newPlayerID: number, newPosition: PlayerInfoTilePosition) => {
    console.log("In fuse handler closure")
    fuseHandler(setFuseActive, setQueuedPlayerID, setQueuedPlayerData, active, playerID, position, newPlayerID, newPosition)
  }

  // const spawnPlayerInfoTile = (playerID: number, position: PlayerInfoTilePosition) => {
  //   console.log("Spawn info tile")
  //   console.log(`Active: ${active}`)
  //   spawnHandlerClosure(playerID, position)
  // }

  const spawnPlayerInfoTile = (playerID: number, position: PlayerInfoTilePosition) => {
    console.log("Spawn info tile")
    console.log(`Active: ${active}`)
    setActive(true)
  }

  // const lightFusePlayerInfoTile = (playerID: number, position: PlayerInfoTilePosition) => {
  //   fuseHandlerClosure(playerID, position)
  // }

  const lightFusePlayerInfoTile = (playerID: number, position: PlayerInfoTilePosition) => {
    console.log("Fuse info tile")
    console.log(`Active: ${active}`)
    setActive(false)
  }


  const [tileContext, _setTileContext] = useState<PlayerInfoTileContextInterface>({
    spawnPlayerInfoTile,
    lightFusePlayerInfoTile,
  })

  useEffect(() => {
    if (playerID) {
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
        destroyTile(makePlayerInfoClosure, setActive, setQueuedPlayerID, setQueuedPosition, setQueuedPlayerData, queuedPlayerID, queuedPosition)
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

  if (active == false) {
    return (
      <PlayerInfoTileContext.Provider value={tileContext}>
        {children}
      </PlayerInfoTileContext.Provider>
    )
  }

  return (
    <PlayerInfoTileContext.Provider value={tileContext}>
      {children}
      
      <div className="playerInfoTile" style={{transform: `translate(${position.x}px, ${position.y}px);`}}>
        <div className="Name&Ping">
          {`Player Name: ${playerData?.displayName}`}
        </div>

        <div className="Ratings">

        </div>

        <div className="Games&JoinDate">

        </div>
      </div>
    </PlayerInfoTileContext.Provider>
  )
}