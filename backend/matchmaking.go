package main

import (
	"fmt"
	"log"
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

// Lock openPol?
var openPool = OpenPool{openPool: 0} // The pool to add waiting players too, 0 for A, 1 for B
var waitingToJoinPoolA []playerMatchmakingData
var waitingToJoinPoolB []playerMatchmakingData
var matchmakingPool []playerMatchmakingData
var awaitingRemoval = RemovalMap{awaitingRemoval: make(map[int64]bool)}
var pendingRemovalRequests []int // Array of playerIDs

const defaultMatchmakingThreshold = 400

func addPlayerToWaitingPool(playerID int64) {
	awaitingRemoval.mu.Lock()
	_, ok := awaitingRemoval.awaitingRemoval[playerID]

	if ok {
		// Player already in queue, remove leave request if it exists
		awaitingRemoval.awaitingRemoval[playerID] = false
		awaitingRemoval.mu.Unlock()
		return
	}
	awaitingRemoval.mu.Unlock()

	var pools = []*[]playerMatchmakingData{&waitingToJoinPoolA, &waitingToJoinPoolB}

	// Should openPool be locked?
	openPool.mu.Lock()
	*pools[openPool.openPool] = append(*pools[openPool.openPool],
		playerMatchmakingData{
			playerID:             playerID,
			elo:                  1500,
			matchmakingThreshold: defaultMatchmakingThreshold,
			isMatched:            false,
		})

	openPool.mu.Unlock()
	// Add playerID to awaitingRemoval map for easier check later
	awaitingRemoval.mu.Lock()
	awaitingRemoval.awaitingRemoval[playerID] = false
	awaitingRemoval.mu.Unlock()
}

func removePlayerFromWaitingPool(playerID int64) {
	// If value does not exist, then player is not in queue
	awaitingRemoval.mu.Lock()
	_, ok := awaitingRemoval.awaitingRemoval[playerID]
	if ok {
		awaitingRemoval.awaitingRemoval[playerID] = true
	}
	awaitingRemoval.mu.Unlock()
	return
}

func calculateMatchingScore(playerOne playerMatchmakingData, playerOneIdx int, playerTwo playerMatchmakingData, playerTwoIdx int) matchingScore {
	return matchingScore{
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

func createMatch(playerOneID int64, playerTwoID int64) error {
	playerOneIsWhite := rand.Intn(2) == 1
	matchID, err := addMatchToDatabase(playerOneID, playerTwoID, playerOneIsWhite)
	if err != nil {
		return err
	}
	clients.mu.Lock()
	defer clients.mu.Unlock()
	var ok bool

	_, ok = clients.clients[playerOneID]
	if !ok {
		clients.clients[playerOneID] = &Client{id: playerOneID, channel: make(chan string)}
	}
	clients.clients[playerOneID].channel <- fmt.Sprintf("%v", matchID)

	_, ok = clients.clients[playerTwoID]
	if !ok {
		clients.clients[playerTwoID] = &Client{id: playerTwoID, channel: make(chan string)}
	}
	clients.clients[playerTwoID].channel <- fmt.Sprintf("%v", matchID)

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

	// Lock to change pool
	var poolToEmpty int
	openPool.mu.Lock()
	if openPool.openPool == 0 {
		openPool.openPool = 1
		poolToEmpty = 0
	} else {
		openPool.openPool = 0
		poolToEmpty = 1
	}
	openPool.mu.Unlock()

	// Empty the pool
	var pools = []*[]playerMatchmakingData{&waitingToJoinPoolA, &waitingToJoinPoolB}
	matchmakingPool = append(matchmakingPool, *pools[poolToEmpty]...)
	*pools[poolToEmpty] = []playerMatchmakingData{}

	var validMatches = []matchingScore{}

	// Should we sort validMatches first?

	// Score players against all other players
	for playerOneIdx, playerOne := range matchmakingPool {
		for playerTwoIdx, playerTwo := range matchmakingPool[playerOneIdx+1:] {
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
		playerOne := &matchmakingPool[score.playerOneIdx]
		playerTwo := &matchmakingPool[score.playerTwoIdx]

		awaitingRemoval.mu.Lock()
		if playerOne.isMatched || awaitingRemoval.awaitingRemoval[playerOne.playerID] {
			awaitingRemoval.mu.Unlock()
			continue
		}

		if playerTwo.isMatched || awaitingRemoval.awaitingRemoval[playerTwo.playerID] {
			awaitingRemoval.mu.Unlock()
			continue
		}

		awaitingRemoval.mu.Unlock()
		// Match players
		err := createMatch(playerOne.playerID, playerTwo.playerID)
		if err != nil {
			log.Println(err)
			continue
		}
		playerOne.isMatched = true
		playerTwo.isMatched = true

	}

	// Cleanup pool
	for i := len(matchmakingPool) - 1; i >= 0; i-- {
		player := matchmakingPool[i]
		// @TODO
		// Is it possible for an awaitingRemoval key to be deleted whilst the player is in a waiting pool?
		awaitingRemoval.mu.Lock()
		if player.isMatched || awaitingRemoval.awaitingRemoval[player.playerID] {
			matchmakingPool = swapRemove(matchmakingPool, i)
			delete(awaitingRemoval.awaitingRemoval, player.playerID)
		}
		awaitingRemoval.mu.Unlock()
	}

}

func matchmakingService() {
	for {
		fmt.Println("Matching")
		fmt.Println(waitingToJoinPoolA)
		fmt.Println(waitingToJoinPoolB)
		fmt.Println(awaitingRemoval.awaitingRemoval)
		fmt.Println(matchmakingPool)
		matchPlayers()
		time.Sleep(500 * time.Millisecond)
	}
}
