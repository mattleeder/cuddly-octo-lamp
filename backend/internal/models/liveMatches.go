package models

import (
	"database/sql"
	"time"
)

type LiveMatch struct {
	MatchID       int64         `json:"matchID"`
	WhitePlayerID int64         `json:"whitePlayerID"`
	BlackPlayerID int64         `json:"blackPlayerID"`
	LastMovePiece sql.NullInt64 `json:"lastMovePiece"`
	LastMoveMove  sql.NullInt64 `json:"lastMoveMove"`
	CurrentFEN    string        `json:"currentFEN"`
}

type LiveMatchModel struct {
	DB *sql.DB
}

func (m *LiveMatchModel) InsertNew(playerOneID int64, playerTwoID int64, playerOneIsWhite bool) (int64, error) {
	sqlStmt := `
	insert or ignore into live_matches (white_player_id, black_player_id) VALUES(?, ?);
	`
	var result sql.Result
	var err error
	for {
		if playerOneIsWhite {
			result, err = m.DB.Exec(sqlStmt, playerOneID, playerTwoID)
		} else {
			result, err = m.DB.Exec(sqlStmt, playerTwoID, playerOneID)
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

	err := row.Scan(&_matchID, &whitePlayerID, &blackPlayerID, &lastMovePiece, &lastMoveMove, &currentFEN)
	if err != nil {
		return nil, err
	}

	return &LiveMatch{
		MatchID:       matchID,
		WhitePlayerID: whitePlayerID,
		BlackPlayerID: blackPlayerID,
		LastMovePiece: lastMovePiece,
		LastMoveMove:  lastMoveMove,
		CurrentFEN:    currentFEN,
	}, nil
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

func (m *LiveMatchModel) UpdateFENForLiveMatch(matchID int64, newFEN string, lastMovePiece int, lastMoveMove int) error {
	defer m.LogAll()
	sqlStmt := `
	UPDATE live_matches
	SET last_move_piece = ?, last_move_move = ?, current_fen = ?
	WHERE match_id = ?
	`
	_, err := m.DB.Exec(sqlStmt, lastMovePiece, lastMoveMove, newFEN, matchID)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	return nil
}
