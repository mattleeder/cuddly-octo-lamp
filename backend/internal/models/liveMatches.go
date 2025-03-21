package models

import (
	"burrchess/internal/chess"
	"database/sql"
	"errors"
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
	GameHistoryJSONString                []byte        `json:"gameHistoryJSONstring"` // []MatchStateHistory{}
	UnixMsTimeOfLastMove                 int64         `json:"unixTimeOfLastMove"`
	AverageElo                           float64       `json:"averageElo"`
	WhitePlayerELo                       float64       `json:"whitePlayerElo"`
	BlackPlayerElo                       float64       `json:"blackPlayerElo"`
	MatchStartTime                       int64         `json:"matchStartTime"`
}

type LiveMatchModel struct {
	DB *sql.DB
}

func (m *LiveMatchModel) InsertNew(playerOneID int64, playerTwoID int64, playerOneIsWhite bool, timeFormatInMilliseconds int64, incrementInMilliseconds int64, gameHistory []byte, averageElo float64, whitePlayerElo float64, blackPlayerElo float64) (int64, error) {
	defer m.LogAll()
	app.infoLog.Printf("Inserting new match")
	var result sql.Result
	var err error

	// // Get ID for match
	// sqlStmt := `
	// select coalesce(max(match_id), 0) from past_matches
	// `
	// row := m.DB.QueryRow(sqlStmt)
	// err = row.Scan(&matchID)
	// if err != nil {
	// 	app.errorLog.Printf("Error getting new matchID %s", err.Error())
	// 	return 0, err
	// }

	// matchID += 1

	// app.infoLog.Printf("Inserting new match with: matchID: %v, timeFormat: %v, increment: %v\n", matchID, timeFormatInMilliseconds, incrementInMilliseconds)

	app.infoLog.Printf("Inserting new match with: timeFormat: %v, increment: %v\n", timeFormatInMilliseconds, incrementInMilliseconds)

	sqlStmt := `
	INSERT INTO live_matches (
	    white_player_id,
		black_player_id,
		time_format_in_milliseconds,
		increment_in_milliseconds,
		white_player_time_remaining_in_milliseconds,
		black_player_time_remaining_in_milliseconds,
		game_history_json_string,
		unix_ms_time_of_last_move,
		average_elo,
		white_player_elo,
		black_player_elo,
		match_start_time
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	// Set white and black remaining time equal to the time format
	for {
		if playerOneIsWhite {
			result, err = m.DB.Exec(sqlStmt, playerOneID, playerTwoID, timeFormatInMilliseconds, incrementInMilliseconds, timeFormatInMilliseconds, timeFormatInMilliseconds, gameHistory, time.Time.UnixMilli(time.Now()), averageElo, whitePlayerElo, blackPlayerElo, time.Time.Unix(time.Now()))
		} else {
			result, err = m.DB.Exec(sqlStmt, playerTwoID, playerOneID, timeFormatInMilliseconds, incrementInMilliseconds, timeFormatInMilliseconds, timeFormatInMilliseconds, gameHistory, time.Time.UnixMilli(time.Now()), averageElo, whitePlayerElo, blackPlayerElo, time.Time.Unix(time.Now()))
		}

		if err != nil && err.Error() == "database is locked (5) (SQLITE_BUSY)" {
			app.errorLog.Printf("%v, sleeping for 50ms\n", err.Error())
			time.Sleep(50 * time.Millisecond)
			continue
		} else if err != nil {
			app.errorLog.Println(err)
			return 0, err
		}

		break
	}

	insertID, err := result.LastInsertId()
	if err != nil {
		app.errorLog.Printf("Unsuccesfully inserted new match with err: %s", err.Error())
	} else {
		app.infoLog.Printf("Succesfully inserted new match with id: %v", insertID)
	}

	return result.LastInsertId()
}

func (m *LiveMatchModel) EnQueueReturnInsertNew(playerOneID int64, playerTwoID int64, playerOneIsWhite bool, timeFormatInMilliseconds int64, incrementInMilliseconds int64, gameHistory []byte, averageElo float64, whitePlayerElo float64, blackPlayerElo float64) (int64, error) {
	result, err := DBTaskQueue.EnQueueReturn(func() (any, error) {
		return m.InsertNew(playerOneID, playerTwoID, playerOneIsWhite, timeFormatInMilliseconds, incrementInMilliseconds, gameHistory, averageElo, whitePlayerElo, blackPlayerElo)
	})
	if err != nil {
		return 0, err
	}
	coercedResult, ok := result.(int64)
	if !ok {
		app.errorLog.Println("coercedResult is not int64")
		return 0, errors.New("coercedResult is not int64")
	}
	return coercedResult, nil
}

func (m *LiveMatchModel) EnQueueInsertNew(playerOneID int64, playerTwoID int64, playerOneIsWhite bool, timeFormatInMilliseconds int64, incrementInMilliseconds int64, gameHistory []byte, averageElo float64, whitePlayerElo float64, blackPlayerElo float64) {
	DBTaskQueue.EnQueue(func() (any, error) {
		return m.InsertNew(playerOneID, playerTwoID, playerOneIsWhite, timeFormatInMilliseconds, incrementInMilliseconds, gameHistory, averageElo, whitePlayerElo, blackPlayerElo)
	})
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
	var gameHistoryJSONString []byte
	var unixMsTimeOfLastMove int64
	var averageElo float64
	var whitePlayerElo float64
	var blackPlayerElo float64
	var matchStartTime int64

	err := row.Scan(&_matchID, &whitePlayerID, &blackPlayerID, &lastMovePiece, &lastMoveMove, &currentFEN, &timeFormatInMilliseconds, &incrementInMilliseconds, &whitePlayerTimeRemainingMilliseconds, &blackPlayerTimeRemainingMilliseconds, &gameHistoryJSONString, &unixMsTimeOfLastMove, &averageElo)
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
		GameHistoryJSONString:                gameHistoryJSONString,
		UnixMsTimeOfLastMove:                 unixMsTimeOfLastMove,
		AverageElo:                           averageElo,
		WhitePlayerELo:                       whitePlayerElo,
		BlackPlayerElo:                       blackPlayerElo,
		MatchStartTime:                       matchStartTime,
	}

	app.infoLog.Printf("%+v", match)

	return match, nil
}

func (m *LiveMatchModel) EnQueueReturnGetFromMatchID(matchID int64) (*LiveMatch, error) {
	result, err := DBTaskQueue.EnQueueReturn(func() (any, error) {
		return m.GetFromMatchID(matchID)
	})
	if err != nil {
		return nil, err
	}
	matchState, ok := result.(*LiveMatch)
	if !ok {
		app.errorLog.Println("matchState is not *models.LiveMatch")
		return nil, errors.New("matchState is not *models.LiveMatch")
	}
	return matchState, nil
}

func (m *LiveMatchModel) LogAll() {
	app.infoLog.Println("Live Matches:")

	rows, err := m.DB.Query("select * from live_matches;")
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	defer rows.Close()
	app.rowsLog.Println(rows.Columns())
	for rows.Next() {
		app.rowsLog.Printf("%v\n", rows)
	}
	err = rows.Err()
	if err != nil {
		app.errorLog.Println(err)
	}
}

func (m *LiveMatchModel) EnQueueLogAll() {
	DBTaskQueue.EnQueueErrorOnlyTask(func() error {
		m.LogAll()
		return nil
	})
}

func (m *LiveMatchModel) UpdateLiveMatch(matchID int64, newFEN string, lastMovePiece int, lastMoveMove int, whitePlayerTimeRemainingMilliseconds int64, blackPlayerTimeRemainingMilliseconds int64, matchStateHistoryJSONstr []byte, timeOfLastMove time.Time) error {
	defer m.LogAll()
	sqlStmt := `
	UPDATE live_matches
	   SET last_move_piece = ?, 
	       last_move_move = ?, 
		   current_fen = ?, 
		   white_player_time_remaining_in_milliseconds = ?, 
		   black_player_time_remaining_in_milliseconds = ?, 
		   game_history_json_string = ?, 
		   unix_ms_time_of_last_move = ?
	 WHERE match_id = ?
	`
	_, err := m.DB.Exec(sqlStmt, lastMovePiece, lastMoveMove, newFEN, whitePlayerTimeRemainingMilliseconds, blackPlayerTimeRemainingMilliseconds, matchStateHistoryJSONstr, time.Time.UnixMilli(timeOfLastMove), matchID)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	return nil
}

func (m *LiveMatchModel) EnQueueReturnUpdateLiveMatch(matchID int64, newFEN string, lastMovePiece int, lastMoveMove int, whitePlayerTimeRemainingMilliseconds int64, blackPlayerTimeRemainingMilliseconds int64, matchStateHistoryJSONstr []byte, timeOfLastMove time.Time) error {
	err := DBTaskQueue.EnQueueReturnErrorOnlyTask(func() error {
		return m.UpdateLiveMatch(matchID, newFEN, lastMovePiece, lastMoveMove, whitePlayerTimeRemainingMilliseconds, blackPlayerTimeRemainingMilliseconds, matchStateHistoryJSONstr, timeOfLastMove)
	})
	return err
}

func (m *LiveMatchModel) EnQueueUpdateLiveMatch(matchID int64, newFEN string, lastMovePiece int, lastMoveMove int, whitePlayerTimeRemainingMilliseconds int64, blackPlayerTimeRemainingMilliseconds int64, matchStateHistoryJSONstr []byte, timeOfLastMove time.Time) {
	DBTaskQueue.EnQueueErrorOnlyTask(func() error {
		return m.UpdateLiveMatch(matchID, newFEN, lastMovePiece, lastMoveMove, whitePlayerTimeRemainingMilliseconds, blackPlayerTimeRemainingMilliseconds, matchStateHistoryJSONstr, timeOfLastMove)
	})
}

func (m *LiveMatchModel) MoveMatchToPastMatches(matchID int64, result int, resultReason chess.GameOverStatusCode, whitePlayerEloGain float64, blackPlayerEloGain float64) error {
	// outcome int
	// draw      = 0
	// whiteWins = 1
	// blackWins = 2
	app.infoLog.Printf("Moving %v to past matches", matchID)

	defer m.LogAll()

	stepOne := `
		-- Step 1: Insert row into past_matches table
	INSERT INTO past_matches (
	    match_id,
		white_player_id,
		black_player_id,
		last_move_piece,
		last_move_move,
		final_fen,
		time_format_in_milliseconds,
		increment_in_milliseconds,
		game_history_json_string,
		result,
		result_reason,
		white_player_elo,
		black_player_elo,
		white_player_elo_gain,
        black_player_elo_gain,
		average_elo,
		match_start_time,
		match_end_time
		)

	SELECT match_id,
           white_player_id,
           black_player_id,
           last_move_piece,
           last_move_move,
           current_fen as final_fen,
           time_format_in_milliseconds,
           increment_in_milliseconds,
           game_history_json_string,
		   ?,
		   ?,
		   white_player_elo,
		   black_player_elo,
		   ?,
		   ?,
		   average_elo,
		   match_start_time,
		   ?
	  FROM live_matches
	 WHERE match_id = ?;`

	stepTwo := `
	-- Step 2: Delete row from live_matches table
	DELETE FROM live_matches
	 WHERE match_id = ?;
	`

	var stmtOne, stmtTwo *sql.Stmt
	// var resultOne, resultTwo sql.Result

	tx, err := m.DB.Begin()
	if err != nil {
		app.errorLog.Printf("Error starting transaction: %v\n", err)
		return err
	}

	stmtOne, err = tx.Prepare(stepOne)
	if err != nil {
		app.errorLog.Printf("Error preparing first statement: %v\n", err)
		return err
	}
	defer stmtOne.Close()

	stmtTwo, err = tx.Prepare(stepTwo)
	if err != nil {
		app.errorLog.Printf("Error preparing second statement: %v\n", err)
		return err
	}
	defer stmtTwo.Close()

	_, err = stmtOne.Exec(result, resultReason, whitePlayerEloGain, blackPlayerEloGain, matchID)
	if err != nil {
		app.errorLog.Printf("Error executing first statement: %v\n", err)
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			app.errorLog.Printf("insert past_matches: unable to rollback: %v", rollbackErr)
		}
		return err
	}

	_, err = stmtTwo.Exec(matchID)
	if err != nil {
		app.errorLog.Printf("Error executing second statement: %v\n", err)
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			app.errorLog.Printf("delete live_matches: unable to rollback: %v", rollbackErr)
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		app.errorLog.Printf("Error commiting transaction: %v\n", err)
		return err
	}

	return nil
}

func (m *LiveMatchModel) EnQueueReturnMoveMatchToPastMatches(matchID int64, result int, resultReason chess.GameOverStatusCode, whitePlayerEloGain float64, blackPlayerEloGain float64) error {
	err := DBTaskQueue.EnQueueReturnErrorOnlyTask(func() error {
		return m.MoveMatchToPastMatches(matchID, result, resultReason, whitePlayerEloGain, blackPlayerEloGain)
	})
	return err
}

func (m *LiveMatchModel) EnQueueMoveMatchToPastMatches(matchID int64, result int, resultReason chess.GameOverStatusCode, whitePlayerEloGain float64, blackPlayerEloGain float64) {
	DBTaskQueue.EnQueueErrorOnlyTask(func() error {
		return m.MoveMatchToPastMatches(matchID, result, resultReason, whitePlayerEloGain, blackPlayerEloGain)
	})
}

func (m *LiveMatchModel) GetHighestEloMatch() (matchID int64, err error) {

	var matchIDorNull sql.NullInt64

	sqlStmt := `
	SELECT match_id
	  FROM live_matches
	 ORDER by average_elo DESC
	 LIMIT 1
	`

	row := m.DB.QueryRow(sqlStmt)
	err = row.Scan(&matchIDorNull)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			app.errorLog.Println("No matches currently being played")
		} else {
			app.errorLog.Printf("Error getting matchID: %s", err.Error())
		}
		return 0, err
	}

	if !matchIDorNull.Valid {
		app.errorLog.Println("MatchID is null")
		return 0, errors.New("MatchID is null")
	}

	matchID = matchIDorNull.Int64

	return matchID, nil
}

func (m *LiveMatchModel) IsPlayerInMatch(playerID int64) (bool, error) {
	sqlStmt := `
	SELECT match_id
	  FROM live_matches
	  WHERE white_player_id = ?
	     OR black_player_id = ?
	 LIMIT 1
	`

	var matchIDorNull sql.NullInt64

	row := m.DB.QueryRow(sqlStmt, playerID, playerID)
	err := row.Scan(&matchIDorNull)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, nil
		} else {
			app.errorLog.Printf("Error getting matchID: %s", err.Error())
		}
	}

	app.infoLog.Printf("Player: %v in match %v\n", playerID, matchIDorNull)

	return true, err
}
