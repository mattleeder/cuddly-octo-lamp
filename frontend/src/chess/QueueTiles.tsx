import React, { useEffect, useRef, useState } from "react"
import { LoaderCircle } from "lucide-react"
import { useNavigate } from "react-router-dom"

interface QueueObject {
    timeFormatInMilliseconds: number,
    incrementInMilliseconds: number,
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
  
export function QueueTiles() {
  const [waiting, setWaiting] = useState(false)
  const [inQueue, setInQueue] = useState(false)
  const [queueName, setQueueName] = useState("")
  const queueNameRef = useRef(queueName)
  const [eventSource, setEventSource] = useState<EventSource | null>(null)
  const navigate = useNavigate()
  
  useEffect(() => {
    queueNameRef.current = queueName
  }, [queueName])
  
  useEffect(() => {
    const leaveOnUnmount = async () => {
      try {
        await tryLeaveQueue(queueNameRef.current)
      } catch (e) {
        console.error(e)
      }
    }
    return () => {
      leaveOnUnmount()
    }
  }, [])
  
    enum ClickAction {
      leaveQueue,
      joinQueue,
      changeQueue,
    }
  
    async function tryJoinQueue(queueName: string) {
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
  
    async function tryLeaveQueue(queueName: string) {
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
  
    async function toggleQueue(newQueueName: string) {
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
          await tryLeaveQueue(queueName)
          setInQueue(false)
          setQueueName("")
          break
  
        case ClickAction.changeQueue:
          await tryLeaveQueue(queueName)
          await tryJoinQueue(newQueueName)
          setQueueName(newQueueName)
          break
        
        case ClickAction.joinQueue:
          await tryJoinQueue(newQueueName)
          setInQueue(true)
          setQueueName(newQueueName)
        }
      } catch (e) {
        console.error(e)
      } finally {
        setWaiting(false)
      }
  
    }
  
    function QueueButton({ nameOfQueue, queueType }: { nameOfQueue: string, queueType: string }) {
      const loading = nameOfQueue == queueName
      return (
        <>
          {loading ?
            <button onClick={() => toggleQueue(nameOfQueue)}><LoaderCircle className="loaderSpin"/></button>
            :
            <button onClick={() => toggleQueue(nameOfQueue)}><span>{nameOfQueue}<br />{queueType}</span></button>
          }
        </>
      )
    }
  
    return (
      <div className="queueTilesContainer">
        <QueueButton nameOfQueue="1 + 0" queueType="Bullet"/>
        <QueueButton nameOfQueue="2 + 1" queueType="Bullet"/>
        <QueueButton nameOfQueue="3 + 0" queueType="Blitz"/>
  
        <QueueButton nameOfQueue="3 + 2" queueType="Blitz"/>
        <QueueButton nameOfQueue="5 + 0" queueType="Blitz"/>
        <QueueButton nameOfQueue="5 + 3" queueType="Blitz"/>
  
        <QueueButton nameOfQueue="10 + 0" queueType="Rapid"/>
        <QueueButton nameOfQueue="10 + 5" queueType="Rapid"/>
        <QueueButton nameOfQueue="15 + 10" queueType="Rapid"/>
      </div>
    )
}