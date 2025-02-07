package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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
	create table live_matches (id integer not null primary key, current_fen text default "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1");
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
		log.Fatal(err)
	}

	defer rows.Close()
	for rows.Next() {
		var playerid int
		err = rows.Scan(&playerid)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(playerid)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

}

func addMatchToDatabase(playerOneID int64, playerTwoID int64) error {
	sqlStmt := `
	insert or ignore into live_matches (white_player_id, black_player_id) VALUES(?, ?);
	`
	_, err := db.Exec(sqlStmt, playerOneID, playerTwoID)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
