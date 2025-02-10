package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDatabase() *sql.DB {
	os.Remove("./chess_site.db")
	var err error

	db, err = sql.Open("sqlite", "./chess_site.db")
	if err != nil {
		defer db.Close()
		log.Fatal(err)
	}

	sqlStmt := `
	create table live_matches (match_id integer not null primary key, white_player_id integer not null, black_player_id integer not null, last_move_piece integer, last_move_move integer, current_fen text default "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1");
	delete from live_matches;

	create table matchmaking_queue (playerid integer not null primary key);
	delete from matchmaking_queue;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		defer db.Close()
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}

	return db
}

// func insertNewLiveMatch() (int, error) {
// 	// Inserts a new live match into db and returns the match_id
// 	var sqlStmt string
// 	var prevMaxId int
// 	var res sql.Result

// 	// Get max id
// 	sqlStmt = "select max(id) as id from live_matches"
// 	row := db.QueryRow(sqlStmt)
// 	err := row.Scan(&prevMaxId)
// 	if err != nil {
// 		log.Println(err)
// 		return 0, err
// 	}

// 	newId := prevMaxId + 1
// 	sqlStmt = `
// 	insert into live_matches (id) VALUES(?)
// 	`
// 	res, err = db.Exec(sqlStmt, newId)
// 	// Res will give last insert id

// 	if err != nil {
// 		log.Println(err)
// 		return 0, err
// 	}

// 	return newId, nil
// }

func insertNewLiveMatch() (int64, error) {
	// Inserts a new live match into db and returns the match_id
	sqlStmt := `
	insert into live_matches DEFAULT VALUES;
	`
	// Res will give last insert id

	res, err := db.Exec(sqlStmt)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return res.LastInsertId()
}

func addPlayerToQueue(playerid int64, time int, increment int) error {
	sqlStmt := `
	insert or ignore into matchmaking_queue (playerid) VALUES(?);
	`

	_, err := db.Exec(sqlStmt, playerid)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func removePlayerFromQueue(playerid int64, time int, increment int) error {
	sqlStmt := `
	delete from matchmaking_queue where playerid = ?;
	`

	_, err := db.Exec(sqlStmt, playerid)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func checkQueue() {
	fmt.Println("Queue state")

	rows, err := db.Query("select playerid from matchmaking_queue;")
	if err != nil {
		log.Println(err)
	}

	defer rows.Close()
	for rows.Next() {
		var playerid int
		err = rows.Scan(&playerid)
		if err != nil {
			log.Println(err)
		}
		fmt.Println(playerid)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}

}

func addMatchToDatabase(playerOneID int64, playerTwoID int64, playerOneIsWhite bool) (int64, error) {
	sqlStmt := `
	insert or ignore into live_matches (white_player_id, black_player_id) VALUES(?, ?);
	`
	var result sql.Result
	var err error
	for {
		if playerOneIsWhite {
			result, err = db.Exec(sqlStmt, playerOneID, playerTwoID)
		} else {
			result, err = db.Exec(sqlStmt, playerTwoID, playerOneID)
		}

		if err != nil && err.Error() == "database is locked (5) (SQLITE_BUSY)" {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		break
	}

	if err != nil {
		log.Println(err)
		return 0, err
	}

	return result.LastInsertId()
}

type MatchStateData struct {
	MatchID         int64         `json:"matchID"`
	White_player_id int64         `json:"whitePlayerID"`
	Black_player_id int64         `json:"blackPlayerID"`
	Last_move_piece sql.NullInt64 `json:"lastMovePiece"`
	Last_move_move  sql.NullInt64 `json:"lastMoveMove"`
	Current_fen     string        `json:"currentFEN"`
}

func getLiveMatchStateFromInt64(matchID int64) (*MatchStateData, error) {
	row := db.QueryRow("select * from live_matches where match_id=?;", matchID)

	var _matchID int64
	var white_player_id int64
	var black_player_id int64
	var last_move_piece sql.NullInt64
	var last_move_move sql.NullInt64
	var current_fen string

	err := row.Scan(&_matchID, &white_player_id, &black_player_id, &last_move_piece, &last_move_move, &current_fen)
	if err != nil {
		return nil, err
	}

	return &MatchStateData{
		MatchID:         matchID,
		White_player_id: white_player_id,
		Black_player_id: black_player_id,
		Last_move_piece: last_move_piece,
		Last_move_move:  last_move_move,
		Current_fen:     current_fen,
	}, nil
}

func checkLiveMatches() {
	fmt.Println("Queue state")

	rows, err := db.Query("select * from live_matches;")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		fmt.Println(rows)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}

}

func updateFENForLiveMatch(matchID int64, newFEN string, lastMovePiece int, lastMoveMove int) error {
	defer checkLiveMatches()
	sqlStmt := `
	UPDATE live_matches
	SET last_move_piece = ?, last_move_move = ?, current_fen = ?
	WHERE match_id = ?
	`
	_, err := db.Exec(sqlStmt, lastMovePiece, lastMoveMove, newFEN, matchID)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
