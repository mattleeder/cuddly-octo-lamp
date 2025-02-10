package main

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

var adb *sql.DB

// func initDatabase() *sql.DB {
// 	os.Remove("./chess_site.db")
// 	var err error

// 	db, err = sql.Open("sqlite", "./chess_site.db")
// 	if err != nil {
// 		defer db.Close()
// 		app.errorLog.Fatal(err)
// 	}

// 	sqlStmt := `
// 	create table live_matches (match_id integer not null primary key, white_player_id integer not null, black_player_id integer not null, last_move_piece integer, last_move_move integer, current_fen text default "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1");
// 	delete from live_matches;

// 	create table matchmaking_queue (playerid integer not null primary key);
// 	delete from matchmaking_queue;
// 	`
// 	_, err = db.Exec(sqlStmt)
// 	if err != nil {
// 		defer db.Close()
// 		app.errorLog.Fatalf("%q: %s\n", err, sqlStmt)
// 	}

// 	return db
// }
