package models

import (
	"burrchess/internal/chess"
	"database/sql"
	"errors"
)

type RatingType int

const (
	bullet = iota
	blitz
	rapid
	classical
)

type UserRatingsModel struct {
	DB *sql.DB
}

type UserRatings struct {
	PlayerID        int64  `json:"playerID"`
	Username        string `json:"username"`
	BulletRating    int64  `json:"bulletRating"`
	BlitzRating     int64  `json:"blitzRating"`
	RapidRating     int64  `json:"rapidRating"`
	ClassicalRating int64  `json:"classicalRating"`
}

func (u *UserRatings) GetRatingForTimeFormat(timeFormatInMilliseconds int64) int64 {
	if timeFormatInMilliseconds < chess.Bullet[1] {
		return u.BulletRating
	} else if timeFormatInMilliseconds < chess.Blitz[1] {
		return u.BlitzRating
	} else if timeFormatInMilliseconds < chess.Rapid[1] {
		return u.RapidRating
	} else {
		return u.ClassicalRating
	}
}

func GetRatingTypeFromTimeFormat(timeFormatInMilliseconds int64) RatingType {
	if timeFormatInMilliseconds < chess.Bullet[1] {
		return bullet
	} else if timeFormatInMilliseconds < chess.Blitz[1] {
		return blitz
	} else if timeFormatInMilliseconds < chess.Rapid[1] {
		return rapid
	} else {
		return classical
	}
}

func (m *UserRatingsModel) getRating(username string, playerID int64, queryMode QueryMode) (UserRatings, error) {
	sqlStmt := `
	SELECT player_id,
	       username,
		   bullet_rating,
		   blitz_rating,
		   rapid_rating,
		   classical_rating  
	  FROM user_ratings
	`
	app.infoLog.Printf("Getting rating for username: %s, or playerID: %v\n", username, playerID)

	var _playerID int64
	var _username string
	var bulletRating int64
	var blitzRating int64
	var rapidRating int64
	var classicalRating int64
	var row *sql.Row
	var err error

	if queryMode == qmUsername {
		sqlStmt += ` WHERE username = ?`
		err = QueryRowWithRetry(m.DB, sqlStmt, []any{username}, []any{&_playerID, &_username, &bulletRating, &blitzRating, &rapidRating, &classicalRating})
		app.rowsLog.Println(row)
	} else if queryMode == qmPlayerID {
		sqlStmt += ` WHERE player_id = ?`
		err = QueryRowWithRetry(m.DB, sqlStmt, []any{playerID}, []any{&_playerID, &_username, &bulletRating, &blitzRating, &rapidRating, &classicalRating})
		app.rowsLog.Println(row)
	} else {
		return UserRatings{}, errors.New("queryMode unknown")
	}

	if err != nil {
		app.errorLog.Printf("Error getting user_ratings: %s\n", err.Error())
		return UserRatings{}, err
	}
	return UserRatings{
		PlayerID:        _playerID,
		Username:        _username,
		BulletRating:    bulletRating,
		BlitzRating:     blitzRating,
		RapidRating:     rapidRating,
		ClassicalRating: classicalRating,
	}, nil
}

func (m *UserRatingsModel) GetRatingFromUsername(username string) (UserRatings, error) {
	return m.getRating(username, 0, qmUsername)
}

func (m *UserRatingsModel) GetRatingFromPlayerID(playerID int64) (UserRatings, error) {
	return m.getRating("", playerID, qmPlayerID)
}

func (m *UserRatingsModel) updateRating(username string, playerID int64, ratingType RatingType, newRating int64, queryMode QueryMode) error {
	app.infoLog.Printf("Updating rating to %v\n", newRating)

	sqlStmt := `
	UPDATE user_ratings
	`

	switch ratingType {
	case bullet:
		sqlStmt += " SET bullet_rating = ?"
	case blitz:
		sqlStmt += " SET blitz_rating = ?"
	case rapid:
		sqlStmt += " SET rapid_rating = ?"
	case classical:
		sqlStmt += " SET classical_rating = ?"
	}
	var err error

	if queryMode == qmUsername {
		sqlStmt += ` WHERE username = ?`
	} else if queryMode == qmPlayerID {
		sqlStmt += ` WHERE player_id = ?`
	} else {
		return errors.New("queryMode unknown")
	}

	tx, err := m.DB.Begin()
	if err != nil {
		app.errorLog.Printf("Error starting transaction: %v\n", err)
		return err
	}

	stmtOne, err := tx.Prepare(sqlStmt)
	if err != nil {
		app.errorLog.Printf("Error preparing statement: %v\n", err)
		return err
	}
	defer stmtOne.Close()

	if queryMode == qmUsername {
		_, err = ExecStatementWithRetry(stmtOne, newRating, username)
	} else if queryMode == qmPlayerID {
		_, err = ExecStatementWithRetry(stmtOne, newRating, playerID)
	}

	if err != nil {
		app.errorLog.Printf("Error executing statement: %v\n", err)
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			app.errorLog.Printf("updateRating: unable to rollback: %v", rollbackErr)
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		app.errorLog.Printf("Error commiting transaction in updateRating: %v\n", err)
		return err
	}

	return err
}

func (m *UserRatingsModel) UpdateRatingFromUsername(username string, ratingType RatingType, newRating int64) error {
	return m.updateRating(username, 0, ratingType, newRating, qmUsername)
}

func (m *UserRatingsModel) UpdateRatingFromPlayerID(playerID int64, ratingType RatingType, newRating int64) error {
	return m.updateRating("", playerID, ratingType, newRating, qmPlayerID)
}

func (m *UserRatingsModel) LogAll() {
	app.infoLog.Println("UserRatings:")

	rows, err := QueryWithRetry(m.DB, "select * from user_ratings;")
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
