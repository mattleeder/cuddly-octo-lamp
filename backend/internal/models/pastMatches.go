package models

import (
	"database/sql"
)

type PastMatch struct {
	MatchID                  int64         `json:"matchID"`
	WhitePlayerID            int64         `json:"whitePlayerID"`
	BlackPlayerID            int64         `json:"blackPlayerID"`
	LastMovePiece            sql.NullInt64 `json:"lastMovePiece"`
	LastMoveMove             sql.NullInt64 `json:"lastMoveMove"`
	FinalFEN                 string        `json:"currentFEN"`
	TimeFormatInMilliseconds int64         `json:"timeFormatInMilliseconds"`
	IncrementInMilliseconds  int64         `json:"incrementInMilliseconds"`
	WhitePlayerPoints        float64       `json:"whitePlayerTimeRemainingMilliseconds"`
	BlackPlayerPoints        float64       `json:"blackPlayerTimeRemainingMilliseconds"`
	GameHistoryJSONString    []byte        `json:"gameHistoryJSONstring"` // []MatchStateHistory{}
}

type PastMatchSummary struct {
	MatchID                  int64         `json:"matchID"`
	WhitePlayerUsername      string        `json:"whitePlayerUsername"`
	BlackPlayerUsername      string        `json:"blackPlayerUsername"`
	LastMovePiece            sql.NullInt64 `json:"lastMovePiece"`
	LastMoveMove             sql.NullInt64 `json:"lastMoveMove"`
	FinalFEN                 string        `json:"finalFEN"`
	TimeFormatInMilliseconds int64         `json:"timeFormatInMilliseconds"`
	IncrementInMilliseconds  int64         `json:"incrementInMilliseconds"`
	WhitePlayerPoints        float64       `json:"whitePlayerPoints"`
	BlackPlayerPoints        float64       `json:"blackPlayerPoints"`
	AverageElo               float64       `json:"averageElo"`
}

type PastMatchModel struct {
	DB *sql.DB
}

func (m *PastMatchModel) LogAll() {
	app.infoLog.Println("Past Matches:")

	rows, err := m.DB.Query("select * from past_matches;")
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

func (m *PastMatchModel) GetPastMatchesWithFormat(timeFormatLower int64, timeFormatUpper int64) ([]PastMatchSummary, error) {
	sqlStmt := `
	SELECT m.match_id,
	       white_player.username,
		   black_player.username,
		   m.last_move_piece,
		   m.last_move_move,
		   m.final_fen,
		   m.time_format_in_milliseconds,
		   m.increment_in_milliseconds,
		   m.white_player_points,
		   m.black_player_points,
		   m.average_elo
	  FROM past_matches as m
	 INNER JOIN users as white_player
	    ON m.white_player_id = white_player.player_id
	 INNER JOIN users as black_player
	    on m.black_player_id = black_player.player_id
	 WHERE m.time_format_in_milliseconds > ?
	   AND m.time_format_in_milliseconds <= ?
	`

	var output []PastMatchSummary
	var matchID int64
	var whitePlayerUsername string
	var blackPlayerUsername string
	var lastMovePiece sql.NullInt64
	var lastMoveMove sql.NullInt64
	var finalFEN string
	var timeFormatMilliseconds int64
	var incrementMilliseconds int64
	var whitePlayerPoints float64
	var blackPlayerPoints float64
	var averageElo float64

	rows, err := m.DB.Query(sqlStmt, timeFormatLower, timeFormatUpper)
	if err != nil {
		app.errorLog.Printf("Error getting past matches with format: %s\n", err.Error())
		return nil, err
	}

	for rows.Next() {
		err := rows.Scan(
			&matchID,
			&whitePlayerUsername,
			&blackPlayerUsername,
			&lastMovePiece,
			&lastMoveMove,
			&finalFEN,
			&timeFormatMilliseconds,
			&incrementMilliseconds,
			&whitePlayerPoints,
			&blackPlayerPoints,
			&averageElo,
		)

		if err != nil {
			app.errorLog.Printf("Error in GetPastMatchesWithFormat: %s\n", err.Error())
			return nil, err
		}

		output = append(output, PastMatchSummary{
			MatchID:                  matchID,
			WhitePlayerUsername:      whitePlayerUsername,
			BlackPlayerUsername:      blackPlayerUsername,
			LastMovePiece:            lastMovePiece,
			LastMoveMove:             lastMoveMove,
			FinalFEN:                 finalFEN,
			TimeFormatInMilliseconds: timeFormatMilliseconds,
			IncrementInMilliseconds:  incrementMilliseconds,
			WhitePlayerPoints:        whitePlayerPoints,
			BlackPlayerPoints:        blackPlayerPoints,
			AverageElo:               averageElo,
		})
	}

	return output, nil
}
