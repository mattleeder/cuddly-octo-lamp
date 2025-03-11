import React, { useEffect, useState } from 'react';

function fetchUserPing() {
  return 10
}

function fetchServerLatency() {
  return 50
}


function LatencyBars({ userPing, serverLatency }: { userPing: number, serverLatency: number }) {
  const pingThresholds = [50, 250, 500]
  let pingLevel = 0

  for (null; pingLevel < pingThresholds.length; pingLevel++) {
    if (userPing + serverLatency <= pingThresholds[pingLevel]) {
      break
    }
  }

  switch (pingLevel) {
  case 0:
    return (
      <div>0</div>
    )

  case 1:
    return (
      <div>1</div>
    )

  case 2:
    return (
      <div>2</div>
    )

  case 3:
    return (
      <div>3</div>
    )
  }
}

export function LatencyDisplay() {
  const [userPing, setUserPing] = useState<number | null>(null)
  const [serverLatency, setServerLatency] = useState<number | null>(null)

  useEffect(() => {
    setUserPing(fetchUserPing())
    setServerLatency(fetchServerLatency())
  }, [])

  return (
    <div>
      <div>
        <div>
          Ping: {userPing == null ? "?" : userPing}
        </div>
        <div>
          Server: {serverLatency == null ? "?" : serverLatency}
        </div>
      </div>

      <div>
        <LatencyBars userPing={userPing == null ? 0 : userPing} serverLatency={serverLatency == null ? 0 : serverLatency}/>
      </div>
    </div>
  )
}