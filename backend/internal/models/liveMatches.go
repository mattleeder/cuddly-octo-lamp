package models

import (
	"database/sql"
	"errors"
	"fmt"
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
}

type LiveMatchModel struct {
	DB *sql.DB
}

func (m *LiveMatchModel) InsertNew(playerOneID int64, playerTwoID int64, playerOneIsWhite bool, timeFormatInMilliseconds int64, incrementInMilliseconds int64, gameHistory []byte) (int64, error) {
	defer m.LogAll()
	app.infoLog.Printf("Inserting new match with: %v, %v\n", timeFormatInMilliseconds, incrementInMilliseconds)
	sqlStmt := `
	insert or ignore into live_matches (white_player_id, black_player_id, time_format_in_milliseconds, increment_in_milliseconds, white_player_time_remaining_in_milliseconds, black_player_time_remaining_in_milliseconds, game_history_json_string, unix_ms_time_of_last_move) VALUES(?, ?, ?, ?, ?, ?, ?, ?);
	`
	var result sql.Result
	var err error
	// Set white and black remaining time equal to the time format
	for {
		if playerOneIsWhite {
			result, err = m.DB.Exec(sqlStmt, playerOneID, playerTwoID, timeFormatInMilliseconds, incrementInMilliseconds, timeFormatInMilliseconds, timeFormatInMilliseconds, gameHistory, time.Time.UnixMilli(time.Now()))
		} else {
			result, err = m.DB.Exec(sqlStmt, playerTwoID, playerOneID, timeFormatInMilliseconds, incrementInMilliseconds, timeFormatInMilliseconds, timeFormatInMilliseconds, gameHistory, time.Time.UnixMilli(time.Now()))
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

	app.infoLog.Println("Succesfully inserted new match")

	return result.LastInsertId()
}

func (m *LiveMatchModel) EnQueueReturnInsertNew(playerOneID int64, playerTwoID int64, playerOneIsWhite bool, timeFormatInMilliseconds int64, incrementInMilliseconds int64, gameHistory []byte) (int64, error) {
	result, err := DBTaskQueue.EnQueueReturn(func() (any, error) {
		return m.InsertNew(playerOneID, playerTwoID, playerOneIsWhite, timeFormatInMilliseconds, incrementInMilliseconds, gameHistory)
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

func (m *LiveMatchModel) EnQueueInsertNew(playerOneID int64, playerTwoID int64, playerOneIsWhite bool, timeFormatInMilliseconds int64, incrementInMilliseconds int64, gameHistory []byte) {
	DBTaskQueue.EnQueue(func() (any, error) {
		return m.InsertNew(playerOneID, playerTwoID, playerOneIsWhite, timeFormatInMilliseconds, incrementInMilliseconds, gameHistory)
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

	err := row.Scan(&_matchID, &whitePlayerID, &blackPlayerID, &lastMovePiece, &lastMoveMove, &currentFEN, &timeFormatInMilliseconds, &incrementInMilliseconds, &whitePlayerTimeRemainingMilliseconds, &blackPlayerTimeRemainingMilliseconds, &gameHistoryJSONString, &unixMsTimeOfLastMove)
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
	SET last_move_piece = ?, last_move_move = ?, current_fen = ?, white_player_time_remaining_in_milliseconds = ?, black_player_time_remaining_in_milliseconds = ?, game_history_json_string = ?, unix_ms_time_of_last_move = ?
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

func (m *LiveMatchModel) MoveMatchToPastMatches(matchID int64, outcome int) error {
	// outcome int
	// draw      = 0
	// whiteWins = 1
	// blackWins = 2
	app.infoLog.Printf("Moving %v to past matches", matchID)

	defer m.LogAll()
	var white_player_points, black_player_points float32

	if outcome == 0 {
		white_player_points = 0.5
		black_player_points = 0.5
	} else if outcome == 1 {
		white_player_points = 1
		black_player_points = 0
	} else if outcome == 2 {
		white_player_points = 0
		black_player_points = 1
	} else {
		app.errorLog.Printf("Could not understand outcome: %v\n", outcome)
		return errors.New(fmt.Sprintf("Could not understand outcome: %v\n", outcome))
	}

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
		white_player_points,
		black_player_points
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

	_, err = stmtOne.Exec(white_player_points, black_player_points, matchID)
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

func (m *LiveMatchModel) EnQueueReturnMoveMatchToPastMatches(matchID int64, outcome int) error {
	err := DBTaskQueue.EnQueueReturnErrorOnlyTask(func() error {
		return m.MoveMatchToPastMatches(matchID, outcome)
	})
	return err
}

func (m *LiveMatchModel) EnQueueMoveMatchToPastMatches(matchID int64, outcome int) {
	DBTaskQueue.EnQueueErrorOnlyTask(func() error {
		return m.MoveMatchToPastMatches(matchID, outcome)
	})
}
