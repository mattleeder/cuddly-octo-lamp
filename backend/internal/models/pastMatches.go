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
