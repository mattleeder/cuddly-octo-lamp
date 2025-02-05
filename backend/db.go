package main

import (
	"database/sql"
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
