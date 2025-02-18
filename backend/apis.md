# Websocket to clients

    ## On connect
    {  
        MessageType: "onConnect",
        body: {
            MatchStateHistory: [
                {
                    FEN: string,
                    lastMove: [int, int],
                    algebraicNotation: string,
                    whitePlayerTimeRemainingMilliseconds: int,
                    blackPlayerTimeRemainingMilliseconds: int,
                },
            ],
            gameOverStatusCode: int,
            threefoldRepetition: bool,

            whitePlayerConnected: bool,
            blackPlayerConnected: bool,
        }
    }

    ## On Move
    {
        MessageType: "onMove",
        body: {
            MatchStateHistory: [
                {
                    FEN: string,
                    lastMove: [int, int],
                    algebraicNotation: string,
                    whitePlayerTimeRemainingMilliseconds: int,
                    blackPlayerTimeRemainingMilliseconds: int,
                },
            ],
            gameOverStatusCode: int,
            threefoldRepetition: bool,
        }
    }

    ## On Player Connect/Disconnect
    {
        MessageType: "connectionMessage",
        body: {
            playerColour: string,
            isConnected: bool,
        }
    }

    ## Opponent event
    {
        MessageType: "opponentEvent",
        body: {
            sender: string,
            eventType: string, (takeback, draw, resign, extra time, abort, rematch) accepts
        }
    }

    ## On User Message
    {
        MessageType: "userMessage"
        body: {
            sender: string,
            messageContent: string,
        }
    }

# Client to websocket

    ## Post Move
    {
        MessageType: "postMove"
        body: {
            piece: int,
            move: int,
            promotionString: string,
        }
    }

    ## Player event
    {
        messageType: "playerEvent"
        body: {
            eventType: string, (takeback, draw, resign, extra time, abort, rematch),
        }
    }

    ## User Message
    {
        messageType: "userMessage"
        body: {
            messageContent: string,
        }
    }