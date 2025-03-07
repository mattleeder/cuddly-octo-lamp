import React, { useEffect, useRef, useState } from "react"
import { LoaderCircle } from "lucide-react"
import { NavigateFunction, useNavigate } from "react-router-dom"

interface QueueObject {
  timeFormatInMilliseconds: number,
  incrementInMilliseconds: number,
}

interface QueueState {
    waiting: boolean, 
    setWaiting: React.Dispatch<React.SetStateAction<boolean>>,
    inQueue: boolean,
    setInQueue: React.Dispatch<React.SetStateAction<boolean>>,
    queueName: string, 
    setQueueName:  React.Dispatch<React.SetStateAction<string>>, 
    eventSource: EventSource | null, 
    setEventSource:  React.Dispatch<React.SetStateAction<EventSource | null>>, 
    navigate: NavigateFunction
}

enum ClickAction {
  leaveQueue,
  joinQueue,
  changeQueue,
}
  
const queueObjectsMap = new Map<string, QueueObject>()
  
function addQueueObject(timeFormatInMinutes: number, incrementInSeconds: number) {
  queueObjectsMap.set(`${timeFormatInMinutes} + ${incrementInSeconds}`, {
    timeFormatInMilliseconds: timeFormatInMinutes * 60 * 1000,
    incrementInMilliseconds: incrementInSeconds * 1000,
  })
}
  
addQueueObject(1, 0)
addQueueObject(2, 1)
addQueueObject(3, 0)
  
addQueueObject(3, 2)
addQueueObject(5, 0)
addQueueObject(5, 3)
  
addQueueObject(10, 0)
addQueueObject(10, 5)
addQueueObject(15, 10)

async function tryJoinQueue(queueName: string, navigate: NavigateFunction, setEventSource: React.Dispatch<React.SetStateAction<EventSource | null>>) {
  const queueObject = queueObjectsMap.get(queueName)
  if (queueObject === undefined) {
    throw new Error("Queue object not found")
  }

  const response = await fetch(import.meta.env.VITE_API_JOIN_QUEUE_URL, {
    signal: AbortSignal.timeout(5000),
    method: "POST",
    credentials: 'include',
    body: JSON.stringify({
      "timeFormatInMilliseconds": queueObject.timeFormatInMilliseconds,
      "incrementInMilliseconds": queueObject.incrementInMilliseconds,
      "action": "join",
    })
  })

  if (!response.ok) {
    throw new Error(response.statusText)
  }

  // Joined, start listening for events
  const eventSource = new EventSource(import.meta.env.VITE_API_MATCH_LISTEN_URL, {
    withCredentials: true,
  })
  eventSource.onmessage = (event) => {
    console.log(`message: ${event.data}`)
    const splitData = event.data.split(",")
    const matchRoom = splitData[0]
    const timeFormatInMilliseconds = splitData[1]
    const incrementInMilliseconds = splitData[2]
    const state = {
      timeFormatInMilliseconds,
      incrementInMilliseconds,
    }
    navigate("matchroom/" + matchRoom, { state })
  }
  setEventSource(eventSource)
}

async function tryLeaveQueue(queueName: string, eventSource: EventSource | null) {
  const queueObject = queueObjectsMap.get(queueName)
  if (queueObject === undefined) {
    throw new Error("Queue object not found")
  }
  const response = await fetch(import.meta.env.VITE_API_JOIN_QUEUE_URL, {
    method: "POST",
    credentials: 'include',
    body: JSON.stringify({
      "timeFormatInMilliseconds": queueObject.timeFormatInMilliseconds,
      "incrementInMilliseconds": queueObject.incrementInMilliseconds,
      "action": "leave",
    })
  })

  if (!response.ok) {
    throw new Error(response.statusText)
  }

  // Left
  eventSource?.close()
}

async function toggleQueue ({
  waiting, 
  setWaiting,
  inQueue,
  setInQueue,
  queueName, 
  setQueueName, 
  eventSource, 
  setEventSource, 
  navigate
}: QueueState, newQueueName: string) {
  if (waiting) {
    return
  }
  
  setWaiting(true)
  let clickAction
  if (!inQueue) {
    clickAction = ClickAction.joinQueue
  } else if (queueName == newQueueName) {
    clickAction = ClickAction.leaveQueue
  } else {
    clickAction = ClickAction.changeQueue
  }
  
  try {
    switch(clickAction) {
    case ClickAction.leaveQueue:
      await tryLeaveQueue(queueName, eventSource)
      setInQueue(false)
      setQueueName("")
      break
  
    case ClickAction.changeQueue:
      await tryLeaveQueue(queueName, eventSource)
      await tryJoinQueue(newQueueName, navigate, setEventSource)
      setQueueName(newQueueName)
      break
        
    case ClickAction.joinQueue:
      await tryJoinQueue(newQueueName, navigate, setEventSource)
      setInQueue(true)
      setQueueName(newQueueName)
    }
  } catch (e) {
    console.error(e)
  } finally {
    setWaiting(false)
  }
  
}

function QueueButton({ queueState, nameOfQueue, queueType }: { queueState: QueueState, nameOfQueue: string, queueType: string }) {
  const loading = nameOfQueue == queueState.queueName
  return (
    <>
      {loading ?
        <button onClick={() => toggleQueue(queueState, nameOfQueue)} className="queueButton"><LoaderCircle className="loaderSpin"/></button>
        :
        <button onClick={() => toggleQueue(queueState, nameOfQueue)} className="queueButton"><span>{nameOfQueue}<br />{queueType}</span></button>
      }
    </>
  )
}

  
export function QueueTiles() {
  const [waiting, setWaiting] = useState(false)
  const [inQueue, setInQueue] = useState(false)
  const [queueName, setQueueName] = useState("")
  const queueNameRef = useRef(queueName)
  const [eventSource, setEventSource] = useState<EventSource | null>(null)
  const eventSourceRef = useRef(eventSource)
  const navigate = useNavigate()

  const queueState: QueueState = {
    waiting,
    setWaiting,
    inQueue,
    setInQueue,
    queueName,
    setQueueName,
    eventSource,
    setEventSource,
    navigate,
  }
  
  useEffect(() => {
    queueNameRef.current = queueName
  }, [queueName])

  useEffect(() => {
    eventSourceRef.current = eventSource
  }, [eventSource])
  
  useEffect(() => {
    const leaveOnUnmount = async () => {
      try {
        await tryLeaveQueue(queueNameRef.current, eventSourceRef.current)
      } catch (e) {
        console.error(e)
      }
    }
    return () => {
      leaveOnUnmount()
    }
  }, [])
  
  return (
    <div className="queueTilesContainer">
      <QueueButton queueState={queueState} nameOfQueue="1 + 0" queueType="Bullet"/>
      <QueueButton queueState={queueState} nameOfQueue="2 + 1" queueType="Bullet"/>
      <QueueButton queueState={queueState} nameOfQueue="3 + 0" queueType="Blitz"/>

      <QueueButton queueState={queueState} nameOfQueue="3 + 2" queueType="Blitz"/>
      <QueueButton queueState={queueState} nameOfQueue="5 + 0" queueType="Blitz"/>
      <QueueButton queueState={queueState} nameOfQueue="5 + 3" queueType="Blitz"/>

      <QueueButton queueState={queueState} nameOfQueue="10 + 0" queueType="Rapid"/>
      <QueueButton queueState={queueState} nameOfQueue="10 + 5" queueType="Rapid"/>
      <QueueButton queueState={queueState} nameOfQueue="15 + 10" queueType="Rapid"/>
    </div>
  )
}