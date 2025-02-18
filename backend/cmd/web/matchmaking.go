package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// Could make a map from playerid to some struct with a mutex
// When remove request comes in, check map and toggle a bool in the struct

// When a player joins the queue add them to the matchmaking pool
// Pause joining
// Find matches
// If no matches found for a player increase threshold
// Enable joining
// Sleep 500ms

// Matchmaking requirements
// 1. Time
// 2. Increment

// Matchmaking factors
// 1. Elo

type playerMatchmakingData struct {
	playerID             int64
	elo                  int
	matchmakingThreshold int
	isMatched            bool
}

type matchingScore struct {
	playerOneID  int64
	playerOneIdx int
	playerTwoID  int64
	playerTwoIdx int
	score        int
}

type OpenPool struct {
	mu       sync.Mutex
	openPool int
}

// Could we use a similar idea to the OpenPool struct
// Have a bool that indicates if the map is in use
// and if it is then put the requests in a queue instead
type RemovalMap struct {
	mu              sync.Mutex
	awaitingRemoval map[int64]bool
}

type QueueData struct {
	openPool                 *OpenPool
	waitingToJoinPoolA       *[]*playerMatchmakingData
	waitingToJoinPoolB       *[]*playerMatchmakingData
	matchmakingPool          *[]*playerMatchmakingData
	awaitingRemoval          *RemovalMap
	pendingRemovalRequests   *[]int
	timeFormatInMilliseconds int64
	incrementInMilliseconds  int64
}

var queueMap = make(map[string]*QueueData)

func addNewQueue(timeFormatInMilliseconds int64, incrementInMilliseconds int64) {
	app.infoLog.Printf("Creating new queue: %v %v\n", timeFormatInMilliseconds, incrementInMilliseconds)
	var key string = fmt.Sprintf("%v + %v", timeFormatInMilliseconds, incrementInMilliseconds)
	queueMap[key] = &QueueData{
		openPool:                 &OpenPool{openPool: 0},
		waitingToJoinPoolA:       &[]*playerMatchmakingData{},
		waitingToJoinPoolB:       &[]*playerMatchmakingData{},
		matchmakingPool:          &[]*playerMatchmakingData{},
		awaitingRemoval:          &RemovalMap{awaitingRemoval: make(map[int64]bool)},
		pendingRemovalRequests:   &[]int{},
		timeFormatInMilliseconds: timeFormatInMilliseconds,
		incrementInMilliseconds:  incrementInMilliseconds,
	}
}

// // Lock openPol?
// var openPool = OpenPool{openPool: 0} // The pool to add waiting players too, 0 for A, 1 for B
// var waitingToJoinPoolA []*playerMatchmakingData
// var waitingToJoinPoolB []*playerMatchmakingData
// var matchmakingPool []*playerMatchmakingData
// var awaitingRemoval = RemovalMap{awaitingRemoval: make(map[int64]bool)}
// var pendingRemovalRequests []int // Array of playerIDs

const defaultMatchmakingThreshold = 400

func addPlayerToWaitingPool(playerID int64, timeFormatInMilliseconds int64, incrementInMilliseconds int64) {
	var key string = fmt.Sprintf("%v + %v", timeFormatInMilliseconds, incrementInMilliseconds)
	queue, ok := queueMap[key]
	if !ok {
		addNewQueue(timeFormatInMilliseconds, incrementInMilliseconds)
		queue = queueMap[key]
	}

	queue.awaitingRemoval.mu.Lock()
	_, ok = queue.awaitingRemoval.awaitingRemoval[playerID]

	if ok {
		// Player already in queue, remove leave request if it exists
		queue.awaitingRemoval.awaitingRemoval[playerID] = false
		queue.awaitingRemoval.mu.Unlock()
		return
	}
	queue.awaitingRemoval.mu.Unlock()

	var pools = []*[]*playerMatchmakingData{queue.waitingToJoinPoolA, queue.waitingToJoinPoolB}

	// Should openPool be locked?
	queue.openPool.mu.Lock()
	*pools[queue.openPool.openPool] = append(*pools[queue.openPool.openPool],
		&playerMatchmakingData{
			playerID:             playerID,
			elo:                  1500,
			matchmakingThreshold: defaultMatchmakingThreshold,
			isMatched:            false,
		})

	queue.openPool.mu.Unlock()
	// Add playerID to awaitingRemoval map for easier check later
	queue.awaitingRemoval.mu.Lock()
	queue.awaitingRemoval.awaitingRemoval[playerID] = false
	queue.awaitingRemoval.mu.Unlock()
}

func removePlayerFromWaitingPool(playerID int64, timeFormatInMilliseconds int64, incrementInMilliseconds int64) {
	var key string = fmt.Sprintf("%v + %v", timeFormatInMilliseconds, incrementInMilliseconds)
	queue, ok := queueMap[key]
	if !ok {
		app.errorLog.Println("Queue not found")
		app.errorLog.Println(queue)
		return
	}
	// If value does not exist, then player is not in queue
	queue.awaitingRemoval.mu.Lock()
	_, ok = queue.awaitingRemoval.awaitingRemoval[playerID]
	if ok {
		queue.awaitingRemoval.awaitingRemoval[playerID] = true
	}
	queue.awaitingRemoval.mu.Unlock()
	return
}

func calculateMatchingScore(playerOne *playerMatchmakingData, playerOneIdx int, playerTwo *playerMatchmakingData, playerTwoIdx int) *matchingScore {
	return &matchingScore{
		playerOneID:  playerOne.playerID,
		playerOneIdx: playerOneIdx,
		playerTwoID:  playerTwo.playerID,
		playerTwoIdx: playerTwoIdx,
		score:        abs(playerOne.elo - playerTwo.elo),
	}
}

func swapRemove[T any](arr []T, idx int) []T {
	arr[idx] = arr[len(arr)-1]
	return arr[:len(arr)-1]
}

func startingMatchHistory(timeFormatInMilliseconds int64) ([]byte, error) {
	startingHistory := []MatchStateHistory{{
		FEN:                                  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		LastMove:                             [2]int{0, 0},
		AlgebraicNotation:                    "a8",
		WhitePlayerTimeRemainingMilliseconds: timeFormatInMilliseconds,
		BlackPlayerTimeRemainingMilliseconds: timeFormatInMilliseconds,
	}}

	app.infoLog.Printf("Starting history: %+v\n", startingHistory)

	jsonStr, err := json.Marshal(startingHistory)
	if err != nil {
		app.errorLog.Printf("Error marshalling JSON: %v\n", err)
		return nil, err
	}

	app.infoLog.Printf("Original data: %s\n", jsonStr)

	var decoded []MatchStateHistory

	err = json.Unmarshal(jsonStr, &decoded)
	if err != nil {
		app.errorLog.Printf("Error unmarshalling JSON: %v\n", err)
		return nil, err
	}

	app.infoLog.Printf("Decoded: %+v\n", decoded)

	return jsonStr, nil
}

func createMatch(playerOneID int64, playerTwoID int64, timeFormatInMilliseconds int64, incrementInMilliseconds int64) error {
	playerOneIsWhite := rand.Intn(2) == 1
	startingHistory, err := startingMatchHistory(timeFormatInMilliseconds)
	if err != nil {
		app.errorLog.Printf("Error creating starting history for new match: %v\n", err)
		return err
	}

	var matchID int64
	matchID, err = app.liveMatches.InsertNew(playerOneID, playerTwoID, playerOneIsWhite, timeFormatInMilliseconds, incrementInMilliseconds, startingHistory)
	if err != nil {
		app.errorLog.Printf("Error inserting new match: %v\n", err)
		return err
	}

	clients.mu.Lock()
	defer clients.mu.Unlock()
	var ok bool

	_, ok = clients.clients[playerOneID]
	if !ok {
		clients.clients[playerOneID] = &Client{id: playerOneID, channel: make(chan string, 1)}
	}

	clients.clients[playerOneID].channel <- fmt.Sprintf("%v,%v,%v", matchID, timeFormatInMilliseconds, incrementInMilliseconds)

	_, ok = clients.clients[playerTwoID]
	if !ok {
		clients.clients[playerTwoID] = &Client{id: playerTwoID, channel: make(chan string, 1)}
	}
	clients.clients[playerTwoID].channel <- fmt.Sprintf("%v,%v,%v", matchID, timeFormatInMilliseconds, incrementInMilliseconds)

	close(clients.clients[playerOneID].channel)
	close(clients.clients[playerTwoID].channel)
	return nil
}

func matchPlayers() {
	// Merge Waiting Pool with Real Pool
	// Go through all players
	// Score all players against all other players
	// If score below threshold for both players, add to list
	// Order valid scores
	// Go through all scores
	// If both players not matched, match them
	// Go through pool in reverse, if a player has been matched, swap remove them

	for _, queue := range queueMap {

		var timeFormatInMilliseconds = queue.timeFormatInMilliseconds
		var incrementInMilliseconds = queue.incrementInMilliseconds

		// Lock to change pool
		var poolToEmpty int
		queue.openPool.mu.Lock()
		if queue.openPool.openPool == 0 {
			queue.openPool.openPool = 1
			poolToEmpty = 0
		} else {
			queue.openPool.openPool = 0
			poolToEmpty = 1
		}
		queue.openPool.mu.Unlock()

		// Empty the pool
		var pools = []*[]*playerMatchmakingData{queue.waitingToJoinPoolA, queue.waitingToJoinPoolB}
		*queue.matchmakingPool = append(*queue.matchmakingPool, *pools[poolToEmpty]...)
		*pools[poolToEmpty] = []*playerMatchmakingData{}

		var validMatches = []*matchingScore{}

		// Should we sort validMatches first?

		// Score players against all other players
		for playerOneIdx, playerOne := range *queue.matchmakingPool {
			for playerTwoIdx, playerTwo := range (*queue.matchmakingPool)[playerOneIdx+1:] {
				// playerTwoIdx starts from 0
				playerTwoIdx += playerOneIdx + 1
				matchingScore := calculateMatchingScore(playerOne, playerOneIdx, playerTwo, playerTwoIdx)

				// Allows players with high threshold to find matches easier
				if matchingScore.score*2 <= playerOne.matchmakingThreshold+playerTwo.matchmakingThreshold {
					validMatches = append(validMatches, matchingScore)
				}
			}
		}

		// Sort Scores
		sort.Slice(validMatches, func(i, j int) bool {
			return validMatches[i].score < validMatches[j].score
		})

		// Go through all scores
		for _, score := range validMatches {
			playerOne := (*queue.matchmakingPool)[score.playerOneIdx]
			playerTwo := (*queue.matchmakingPool)[score.playerTwoIdx]

			queue.awaitingRemoval.mu.Lock()
			if playerOne.isMatched || queue.awaitingRemoval.awaitingRemoval[playerOne.playerID] {
				queue.awaitingRemoval.mu.Unlock()
				continue
			}

			if playerTwo.isMatched || queue.awaitingRemoval.awaitingRemoval[playerTwo.playerID] {
				queue.awaitingRemoval.mu.Unlock()
				continue
			}

			queue.awaitingRemoval.mu.Unlock()
			// Match players
			err := createMatch(playerOne.playerID, playerTwo.playerID, timeFormatInMilliseconds, incrementInMilliseconds)
			if err != nil {
				app.errorLog.Println(err)
				continue
			}
			playerOne.isMatched = true
			playerTwo.isMatched = true

		}

		// Cleanup pool
		for i := len(*queue.matchmakingPool) - 1; i >= 0; i-- {
			player := (*queue.matchmakingPool)[i]
			// @TODO
			// Is it possible for an awaitingRemoval key to be deleted whilst the player is in a waiting pool?
			queue.awaitingRemoval.mu.Lock()
			if player.isMatched || queue.awaitingRemoval.awaitingRemoval[player.playerID] {
				*queue.matchmakingPool = swapRemove(*queue.matchmakingPool, i)
				delete(queue.awaitingRemoval.awaitingRemoval, player.playerID)
			}
			queue.awaitingRemoval.mu.Unlock()
		}
	}
}

func matchmakingService() {
	iterations := 0
	app.infoLog.Printf("Starting matchmakingService")
	defer app.infoLog.Printf("Ending matchmakingService")
	logMatchmaking := false
	for {
		// Could use time.After to ensure service runs every 500ms
		if logMatchmaking && iterations%20 == 0 {
			for key, queue := range queueMap {
				app.infoLog.Println(key)
				app.infoLog.Printf("Matching iteration: %v\n", iterations)
				app.infoLog.Println(queue.waitingToJoinPoolA)
				app.infoLog.Println(queue.waitingToJoinPoolB)
				app.infoLog.Println(queue.awaitingRemoval.awaitingRemoval)
				app.infoLog.Println(queue.matchmakingPool)
			}
		}
		matchPlayers()
		time.Sleep(500 * time.Millisecond)
		iterations += 1
	}
}
