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

	var _playerID int64
	var _username string
	var bulletRating int64
	var blitzRating int64
	var rapidRating int64
	var classicalRating int64
	var row *sql.Row

	if queryMode == qmUsername {
		sqlStmt += ` WHERE username = ?`
		row = m.DB.QueryRow(sqlStmt, username)
	} else if queryMode == qmPlayerID {
		sqlStmt += ` WHERE player_id = ?`
		row = m.DB.QueryRow(sqlStmt, playerID)
	} else {
		return UserRatings{}, errors.New("queryMode unknown")
	}

	err := row.Scan(&_playerID, &_username, &bulletRating, &blitzRating, &rapidRating, &classicalRating)
	if err != nil {
		app.errorLog.Printf("Error getting user_ratings: %v\n", err.Error())
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
	return m.getRating(username, 0, 0)
}

func (m *UserRatingsModel) GetRatingFromPlayerID(playerID int64) (UserRatings, error) {
	return m.getRating("", playerID, 0)
}

func (m *UserRatingsModel) updateRating(username string, playerID int64, ratingType RatingType, newRating float64, queryMode QueryMode) error {
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
		_, err = m.DB.Exec(sqlStmt, newRating, username)
	} else if queryMode == qmPlayerID {
		sqlStmt += ` WHERE player_id = ?`
		_, err = m.DB.Exec(sqlStmt, newRating, playerID)
	} else {
		return errors.New("queryMode unknown")
	}

	if err != nil {
		app.errorLog.Printf("Error updating rating: %v\n", err.Error())
	}

	return err
}

func (m *UserRatingsModel) UpdateRatingFromUsername(username string, ratingType RatingType, newRating float64) error {
	return m.updateRating(username, 0, ratingType, newRating, qmUsername)
}

func (m *UserRatingsModel) UpdateRatingFromPlayerID(playerID int64, ratingType RatingType, newRating float64) error {
	return m.updateRating("", playerID, ratingType, newRating, qmPlayerID)
}
