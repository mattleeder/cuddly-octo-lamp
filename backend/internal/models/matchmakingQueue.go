package models

import (
	"database/sql"
)

type MatchmakingQueue struct {
	Playerid int64 `json:"playerID"`
}

type MatchmakingQueueModel struct {
	DB *sql.DB
}

func (m *MatchmakingQueueModel) AddPlayerToQueue(playerID int64) error {
	sqlStmt := `
	insert or ignore into matchmaking_queue (playerid) VALUES(?);
	`

	_, err := m.DB.Exec(sqlStmt, playerID)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	return nil
}

func (m *MatchmakingQueueModel) RemovePlayerFromQueue(playerID int64) error {
	sqlStmt := `
	delete from matchmaking_queue where playerid = ?;
	`

	_, err := m.DB.Exec(sqlStmt, playerID)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	return nil
}

func (m *MatchmakingQueueModel) LogQueue() {
	app.infoLog.Println("Matchmaking Queue:")

	rows, err := m.DB.Query("select playerid from matchmaking_queue;")
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		var playerid int
		err = rows.Scan(&playerid)
		if err != nil {
			app.errorLog.Println(err)
		}
		app.rowsLog.Println(playerid)
	}
	err = rows.Err()
	if err != nil {
		app.errorLog.Println(err)
	}
}
