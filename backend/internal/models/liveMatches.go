package models

import (
	"database/sql"
	"time"
)

type LiveMatch struct {
	MatchID                              int64         `json:"matchID"`
	WhitePlayerID                        int64         `json:"whitePlayerID"`
	BlackPlayerID                        int64         `json:"blackPlayerID"`
	LastMovePiece                        sql.NullInt64 `json:"lastMovePiece"`
	LastMoveMove                         sql.NullInt64 `json:"lastMoveMove"`
	CurrentFEN                           string        `json:"currentFEN"`
	TimeFormatInMilliseconds             int64         `json:"timeFormatInMilliseconds"`
	IncrementInMilliseconds              int64         `json:"incrementInMilliseconds"`
	WhitePlayerTimeRemainingMilliseconds int64         `json:"whitePlayerTimeRemainingMilliseconds"`
	BlackPlayerTimeRemainingMilliseconds int64         `json:"blackPlayerTimeRemainingMilliseconds"`
}

type LiveMatchModel struct {
	DB *sql.DB
}

func (m *LiveMatchModel) InsertNew(playerOneID int64, playerTwoID int64, playerOneIsWhite bool, timeFormatInMilliseconds int64, incrementInMilliseconds int64) (int64, error) {
	app.infoLog.Printf("Inserting new match with: %v, %v\n", timeFormatInMilliseconds, incrementInMilliseconds)
	sqlStmt := `
	insert or ignore into live_matches (white_player_id, black_player_id, time_format_in_milliseconds, increment_in_milliseconds, white_player_time_remaining_in_milliseconds, black_player_time_remaining_in_milliseconds) VALUES(?, ?, ?, ?, ?, ?);
	`
	var result sql.Result
	var err error
	// Set white and black remaining time equal to the time format
	for {
		if playerOneIsWhite {
			result, err = m.DB.Exec(sqlStmt, playerOneID, playerTwoID, timeFormatInMilliseconds, incrementInMilliseconds, timeFormatInMilliseconds, timeFormatInMilliseconds)
		} else {
			result, err = m.DB.Exec(sqlStmt, playerTwoID, playerOneID, timeFormatInMilliseconds, incrementInMilliseconds, timeFormatInMilliseconds, timeFormatInMilliseconds)
		}

		if err != nil && err.Error() == "database is locked (5) (SQLITE_BUSY)" {
			app.errorLog.Printf("%v, sleeping for 50ms\n", err.Error())
			time.Sleep(50 * time.Millisecond)
			continue
		}

		break
	}

	if err != nil {
		app.errorLog.Println(err)
		return 0, err
	}

	return result.LastInsertId()
}

func (m *LiveMatchModel) GetFromMatchID(matchID int64) (*LiveMatch, error) {
	row := m.DB.QueryRow("select * from live_matches where match_id=?;", matchID)

	var _matchID int64
	var whitePlayerID int64
	var blackPlayerID int64
	var lastMovePiece sql.NullInt64
	var lastMoveMove sql.NullInt64
	var currentFEN string
	var timeFormatInMilliseconds int64
	var incrementInMilliseconds int64
	var whitePlayerTimeRemainingMilliseconds int64
	var blackPlayerTimeRemainingMilliseconds int64

	err := row.Scan(&_matchID, &whitePlayerID, &blackPlayerID, &lastMovePiece, &lastMoveMove, &currentFEN, &timeFormatInMilliseconds, &incrementInMilliseconds, &whitePlayerTimeRemainingMilliseconds, &blackPlayerTimeRemainingMilliseconds)
	if err != nil {
		return nil, err
	}

	match := &LiveMatch{
		MatchID:                              matchID,
		WhitePlayerID:                        whitePlayerID,
		BlackPlayerID:                        blackPlayerID,
		LastMovePiece:                        lastMovePiece,
		LastMoveMove:                         lastMoveMove,
		CurrentFEN:                           currentFEN,
		TimeFormatInMilliseconds:             timeFormatInMilliseconds,
		IncrementInMilliseconds:              incrementInMilliseconds,
		WhitePlayerTimeRemainingMilliseconds: whitePlayerTimeRemainingMilliseconds,
		BlackPlayerTimeRemainingMilliseconds: blackPlayerTimeRemainingMilliseconds,
	}

	app.infoLog.Printf("%+v", match)

	return match, nil
}

func (m *LiveMatchModel) LogAll() {
	app.infoLog.Println("Live Matches:")

	rows, err := m.DB.Query("select * from live_matches;")
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		app.rowsLog.Println(rows)
	}
	err = rows.Err()
	if err != nil {
		app.errorLog.Println(err)
	}
}

func (m *LiveMatchModel) UpdateLiveMatch(matchID int64, newFEN string, lastMovePiece int, lastMoveMove int, whitePlayerTimeRemainingMilliseconds int64, blackPlayerTimeRemainingMilliseconds int64) error {
	defer m.LogAll()
	sqlStmt := `
	UPDATE live_matches
	SET last_move_piece = ?, last_move_move = ?, current_fen = ?, white_player_time_remaining_in_milliseconds = ?, black_player_time_remaining_in_milliseconds = ?
	WHERE match_id = ?
	`
	_, err := m.DB.Exec(sqlStmt, lastMovePiece, lastMoveMove, newFEN, whitePlayerTimeRemainingMilliseconds, blackPlayerTimeRemainingMilliseconds, matchID)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	return nil
}
