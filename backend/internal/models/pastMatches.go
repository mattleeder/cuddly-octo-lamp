package models

import (
	"database/sql"
)

// @TODO: DOES SENDING THINGS AS sql.NullType GIVE AWAY THAT IT IS SQL DATABASE?

type PastMatch struct {
	MatchID                  int64          `json:"matchID"`
	WhitePlayerID            sql.NullString `json:"whitePlayerID"`
	BlackPlayerID            sql.NullString `json:"blackPlayerID"`
	LastMovePiece            sql.NullInt64  `json:"lastMovePiece"`
	LastMoveMove             sql.NullInt64  `json:"lastMoveMove"`
	FinalFEN                 string         `json:"currentFEN"`
	TimeFormatInMilliseconds int64          `json:"timeFormatInMilliseconds"`
	IncrementInMilliseconds  int64          `json:"incrementInMilliseconds"`
	GameHistoryJSONString    []byte         `json:"gameHistoryJSONstring"` // []MatchStateHistory{}
	Result                   int64          `json:"result"`
	ResultReason             int64          `json:"resultReason"`
	WhitePlayerElo           float64        `json:"whitePlayerElo"`
	BlackPlayerElo           float64        `json:"blackPlayerElo"`
	WhitePlayerEloGain       float64        `json:"whitePlayerEloGain"`
	BlackPlayerEloGain       float64        `json:"blackPlayerEloGain"`
	AverageElo               float64        `json:"averageElo"`
	MatchStartTime           int64          `json:"matchStartTime"`
	MatchEndTime             int64          `json:"matchEndTime"`
}

type PastMatchSummary struct {
	MatchID                  int64          `json:"matchID"`
	WhitePlayerUsername      sql.NullString `json:"whitePlayerUsername"`
	BlackPlayerUsername      sql.NullString `json:"blackPlayerUsername"`
	LastMovePiece            sql.NullInt64  `json:"lastMovePiece"`
	LastMoveMove             sql.NullInt64  `json:"lastMoveMove"`
	FinalFEN                 string         `json:"finalFEN"`
	TimeFormatInMilliseconds int64          `json:"timeFormatInMilliseconds"`
	IncrementInMilliseconds  int64          `json:"incrementInMilliseconds"`
	Result                   int64          `json:"result"`
	ResultReason             int64          `json:"resultReason"`
	WhitePlayerElo           float64        `json:"whitePlayerElo"`
	BlackPlayerElo           float64        `json:"blackPlayerElo"`
	WhitePlayerEloGain       float64        `json:"whitePlayerEloGain"`
	BlackPlayerEloGain       float64        `json:"blackPlayerEloGain"`
	AverageElo               float64        `json:"averageElo"`
	MatchStartTime           int64          `json:"matchStartTime"`
	MatchEndTime             int64          `json:"matchEndTime"`
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
	// Left join for anonymous players
	sqlStmt := `
	SELECT m.match_id,
	       white_player.username,
		   black_player.username,
		   m.last_move_piece,
		   m.last_move_move,
		   m.final_fen,
		   m.time_format_in_milliseconds,
		   m.increment_in_milliseconds,
		   m.result,
		   m.result_reason,
		   m.white_player_elo,
		   m.black_player_elo,
		   m.white_player_elo_gain,
		   m.black_player_elo_gain,
		   m.average_elo,
		   m.match_start_time,
		   m.match_end_time
	  FROM past_matches as m
	  LEFT JOIN users as white_player
	    ON m.white_player_id = white_player.player_id
	  LEFT JOIN users as black_player
	    on m.black_player_id = black_player.player_id
	 WHERE m.time_format_in_milliseconds > ?
	   AND m.time_format_in_milliseconds <= ?
	`

	var output []PastMatchSummary
	var matchID int64
	var whitePlayerUsername sql.NullString
	var blackPlayerUsername sql.NullString
	var lastMovePiece sql.NullInt64
	var lastMoveMove sql.NullInt64
	var finalFEN string
	var timeFormatMilliseconds int64
	var incrementMilliseconds int64
	var result int64
	var resultReason int64
	var whitePlayerElo float64
	var blackPlayerElo float64
	var whitePlayerEloGain float64
	var blackPlayerEloGain float64
	var averageElo float64
	var matchStartTime int64
	var matchEndTime int64

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
			&result,
			&resultReason,
			&whitePlayerElo,
			&blackPlayerElo,
			&whitePlayerEloGain,
			&blackPlayerEloGain,
			&averageElo,
			&matchStartTime,
			&matchEndTime,
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
			Result:                   result,
			ResultReason:             resultReason,
			WhitePlayerElo:           whitePlayerElo,
			BlackPlayerElo:           blackPlayerElo,
			WhitePlayerEloGain:       whitePlayerEloGain,
			BlackPlayerEloGain:       blackPlayerEloGain,
			AverageElo:               averageElo,
			MatchStartTime:           matchStartTime,
			MatchEndTime:             matchEndTime,
		})
	}

	return output, nil
}
