package models

import (
	"database/sql"
	"errors"
	"math/rand"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const PASSWORD_COST = 16

type QueryMode int

const (
	qmUsername = iota
	qmPlayerID
)

type QueryConsumer int

const (
	client = iota
	server
)

type NewUserOptions struct {
	email *string
}

type NewUserInfo struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Email      string `json:"email"`
	RememberMe bool   `json:"rememberMe"`
}

type UserLoginInfo struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

type UserModel struct {
	DB *sql.DB
}

type UserServerSide struct {
	PlayerID int64  `json:"playerID"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	JoinDate int64  `json:"joinDate"`
	LastSeen int64  `json:"lastSeen"`
}

type UserClientSide struct {
	PlayerID int64  `json:"playerID"`
	Username string `json:"username"`
	JoinDate int64  `json:"joinDate"`
	LastSeen int64  `json:"lastSeen"`
}

type Ratings struct {
	BulletRating    int64 `json:"bullet"`
	BlitzRating     int64 `json:"blitz"`
	RapidRating     int64 `json:"rapid"`
	ClassicalRating int64 `json:"classical"`
}

type UserTileInfo struct {
	PlayerID      int64   `json:"playerID"`
	Username      string  `json:"username"`
	PingStatus    bool    `json:"pingStatus"`
	JoinDate      int64   `json:"joinDate"`
	LastSeen      int64   `json:"lastSeen"`
	Ratings       Ratings `json:"ratings"`
	NumberOfGames int64   `json:"numberOfGames"`
}

func hashPassword(password string) (string, error) {
	var passwordBytes = []byte(password)
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword(passwordBytes, PASSWORD_COST)
	return string(hashedPasswordBytes), err
}

func doesPasswordMatch(plaintextPassword string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plaintextPassword))
	return err == nil
}

func CreateNewUserOptions(newUser NewUserInfo) (options NewUserOptions) {
	if newUser.Email != "" {
		options.email = &newUser.Email
	}

	return options
}

func (m *UserModel) InsertNew(username string, password string, options *NewUserOptions) (int64, error) {

	app.infoLog.Printf("Inserting new user: %v\n", username)

	if username == "" {
		app.errorLog.Printf("Username not given")
		return 0, errors.New("username not given")
	}

	if username == "" {
		app.errorLog.Printf("Username not given")
		return 0, errors.New("password not given")
	}

	var playerID = rand.Int63()
	hashedPassword, err := hashPassword(password)
	password = "" // Overwrite password to avoid accidental usage
	if err != nil {
		app.errorLog.Printf("Error generating password hash: %v\n", err.Error())
		return 0, err
	}
	var email = sql.NullString{Valid: false}
	if options.email != nil {
		email = sql.NullString{String: *options.email, Valid: true}
	}

	stepOne := `
	insert into users (player_id, username, password, email) VALUES (?, ?, ?, ?);
	`
	stepTwo := `
	insert into user_ratings (player_id, username) VALUES (?, ?);
	`
	var stmtOne, stmtTwo *sql.Stmt

	tx, err := m.DB.Begin()
	if err != nil {
		app.errorLog.Printf("Error starting transaction: %v\n", err)
		return 0, err
	}

	stmtOne, err = tx.Prepare(stepOne)
	if err != nil {
		app.errorLog.Printf("Error preparing first statement: %v\n", err)
		return 0, err
	}
	defer stmtOne.Close()

	stmtTwo, err = tx.Prepare(stepTwo)
	if err != nil {
		app.errorLog.Printf("Error preparing second statement: %v\n", err)
		return 0, err
	}
	defer stmtTwo.Close()

	_, err = ExecStatementWithRetry(stmtOne, playerID, username, hashedPassword, email)

	if err != nil {
		app.errorLog.Printf("Error inserting new user: %s\n", err.Error())
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			app.errorLog.Printf("insert users: unable to rollback: %s", rollbackErr)
		}
		return 0, err
	}

	_, err = ExecStatementWithRetry(stmtTwo, playerID, username)

	if err != nil {
		app.errorLog.Printf("Error executing second statement: %v\n", err)
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			app.errorLog.Printf("insert user_ratings: unable to rollback: %v", rollbackErr)
		}
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		app.errorLog.Printf("Error commiting transaction: %v\n", err)
		return 0, err
	}

	app.infoLog.Printf("Inserted player_id: %v\n", playerID)

	return playerID, nil
}

func (m *UserModel) Authenticate(username string, password string) (playerID int64, authorized bool) {
	app.infoLog.Printf("Checking password for user: %v\n", username)

	sqlStmt := `
	select player_id, password from users where username = ?
	`
	var hashedPassword string
	err := QueryRowWithRetry(m.DB, sqlStmt, []any{password}, []any{&playerID, &hashedPassword})
	if err != nil {
		app.errorLog.Printf("Error getting password for user: %v\n", err.Error())
		return 0, false
	}

	authorized = doesPasswordMatch(password, hashedPassword)

	if !authorized {
		return 0, false
	}

	return playerID, true
}

func (m *UserModel) LogAll() {
	app.infoLog.Println("Users:")

	rows, err := QueryWithRetry(m.DB, "select * from users;")
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

func (m *UserModel) getUserClientSide(username string, playerID int64, queryMode QueryMode) (UserClientSide, error) {
	// queryMode:
	// 0:	username
	// 1:	playerID
	sqlStmt := `
	SELECT player_id,
	       username,
		   join_date,
		   last_seen
	  FROM users
	`

	var _playerID int64
	var _username string
	var join_date int64
	var last_seen int64
	var err error

	if queryMode == qmUsername {
		sqlStmt += ` WHERE username = ?`
		err = QueryRowWithRetry(m.DB, sqlStmt, []any{username}, []any{&_playerID, &_username, &join_date, &last_seen})
	} else if queryMode == qmPlayerID {
		sqlStmt += ` WHERE player_id = ?`
		err = QueryRowWithRetry(m.DB, sqlStmt, []any{playerID}, []any{&_playerID, &_username, &join_date, &last_seen})
	} else {
		return UserClientSide{}, errors.New("queryMode unknown")
	}

	if err != nil {
		app.errorLog.Printf("Error getting user: %v\n", err.Error())
		return UserClientSide{}, err
	}
	return UserClientSide{
		PlayerID: _playerID,
		Username: _username,
		JoinDate: join_date,
		LastSeen: last_seen,
	}, nil
}

func (m *UserModel) GetUserClientSideFromUsername(username string) (UserClientSide, error) {
	return m.getUserClientSide(username, 0, qmUsername)
}

func (m *UserModel) GetUserClientSideFromPlayerID(playerID int64) (UserClientSide, error) {
	return m.getUserClientSide("", playerID, qmPlayerID)
}

func (m *UserModel) getUserServerSide(username string, playerID int64, queryMode QueryMode) (UserServerSide, error) {
	// queryMode:
	// 0:	username
	// 1:	playerID
	sqlStmt := `
	SELECT player_id,
	       username,
		   password,
		   email,
		   join_date,
		   last_seen
	  FROM users
	`

	var _playerID int64
	var _username string
	var password string
	var email string
	var join_date int64
	var last_seen int64
	var err error

	if queryMode == qmUsername {
		sqlStmt += ` WHERE username = ?`
		err = QueryRowWithRetry(m.DB, sqlStmt, []any{username}, []any{&_playerID, &_username, &password, &email, &join_date, &last_seen})
	} else if queryMode == qmPlayerID {
		sqlStmt += ` WHERE player_id = ?`
		err = QueryRowWithRetry(m.DB, sqlStmt, []any{playerID}, []any{&_playerID, &_username, &password, &email, &join_date, &last_seen})

	} else {
		return UserServerSide{}, errors.New("queryMode unknown")
	}

	if err != nil {
		app.errorLog.Printf("Error getting user: %v\n", err.Error())
		return UserServerSide{}, err
	}
	return UserServerSide{
		PlayerID: _playerID,
		Username: _username,
		Password: password,
		Email:    email,
		JoinDate: join_date,
		LastSeen: last_seen,
	}, nil
}

func (m *UserModel) GetUserFromUsername(username string) (UserClientSide, error) {
	return m.getUserClientSide(username, 0, qmUsername)
}

func (m *UserModel) GetUserFromPlayerID(playerID int64) (UserClientSide, error) {
	return m.getUserClientSide("", playerID, qmPlayerID)
}

func (m *UserModel) updateLastSeen(username string, playerID int64, queryMode QueryMode) error {
	// queryMode:
	// 0:	username
	// 1:	playerID
	sqlStmt := `
	UPDATE users
	   SET last_seen = unixepoch('now')
	`

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

	updateStmt, err := tx.Prepare(sqlStmt)
	if err != nil {
		app.errorLog.Printf("Error preparing statement: %v\n", err)
		return err
	}
	defer updateStmt.Close()

	if queryMode == qmUsername {
		_, err = ExecStatementWithRetry(updateStmt, username)
	} else if queryMode == qmPlayerID {
		_, err = ExecStatementWithRetry(updateStmt, playerID)
	}

	if err != nil {
		app.errorLog.Printf("Error executing statement: %v\n", err)
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			app.errorLog.Printf("insert updateLastSeen: unable to rollback: %v", rollbackErr)
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

func (m *UserModel) UpdateLastSeenFromUsername(username string) error {
	return m.updateLastSeen(username, 0, qmUsername)
}

func (m *UserModel) UpdateLastSeenFromPlayerID(playerID int64) error {
	return m.updateLastSeen("", playerID, qmPlayerID)
}

func (m *UserModel) updateEmail(email string, username string, playerID int64, queryMode QueryMode) error {
	// queryMode:
	// 0:	username
	// 1:	playerID
	sqlStmt := `
	UPDATE users
	   SET email = ?
	`

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

	updateStmt, err := tx.Prepare(sqlStmt)
	if err != nil {
		app.errorLog.Printf("Error preparing statement: %v\n", err)
		return err
	}
	defer updateStmt.Close()

	if queryMode == qmUsername {
		_, err = ExecStatementWithRetry(updateStmt, username, email)
	} else if queryMode == qmPlayerID {
		_, err = ExecStatementWithRetry(updateStmt, playerID, email)
	}

	if err != nil {
		app.errorLog.Printf("Error executing statement: %v\n", err)
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			app.errorLog.Printf("insert updateRating: unable to rollback: %v", rollbackErr)
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		app.errorLog.Printf("Error commiting transaction in updateEmail: %v\n", err)
		return err
	}

	return err
}

func (m *UserModel) UpdateEmailFromUsername(username string, email string) error {
	return m.updateEmail(email, username, 0, qmUsername)
}

func (m *UserModel) UpdateEmailFromPlayerID(playerID int64, email string) error {
	return m.updateEmail(email, "", playerID, qmPlayerID)
}

func (m *UserModel) SearchForUsers(searchString string) ([]UserClientSide, error) {
	sqlStmt := `
	SELECT player_id, username, join_date, last_seen
	  FROM users
	 WHERE UPPER(username) GLOB ?
	`

	var output []UserClientSide
	var playerID int64
	var username string
	var joinDate int64
	var lastSeen int64

	rows, err := QueryWithRetry(m.DB, sqlStmt, strings.ToUpper(searchString))
	if err != nil {
		app.errorLog.Printf("Error in SearchForUsers: %s\n", err.Error())
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&playerID, &username, &joinDate, &lastSeen)

		if err != nil {
			app.errorLog.Printf("Error in SearchForUsers: %s\n", err.Error())
			return nil, err
		}

		output = append(output, UserClientSide{
			PlayerID: playerID,
			Username: username,
			JoinDate: joinDate,
			LastSeen: lastSeen,
		})
	}

	return output, nil
}

func (m *UserModel) GetTileInfoFromUsername(username string) (*UserTileInfo, error) {
	sqlStmt := `
	SELECT users.player_id, join_date, last_seen, bullet_rating, blitz_rating, rapid_rating, classical_rating
	  FROM users
	 INNER JOIN user_ratings
	    ON users.player_id = user_ratings.player_id  
	 WHERE users.username = ?
	`

	var playerID int64
	var joinDate int64
	var lastSeen int64
	var bullet_rating int64
	var blitz_rating int64
	var rapid_rating int64
	var classical_rating int64

	err := QueryRowWithRetry(m.DB, sqlStmt, []any{username}, []any{&playerID, &joinDate, &lastSeen, &bullet_rating, &blitz_rating, &rapid_rating, &classical_rating})
	if err != nil {
		app.errorLog.Printf("Error in GetTileInfoFromPlayerID: %s\n", err.Error())
		return nil, err
	}

	sqlStmt = `
	SELECT Count(past_matches.match_id) as number_of_games
	  FROM users
	 INNER JOIN past_matches
	    ON past_matches.white_player_id = users.player_id
		OR past_matches.black_player_id = users.player_id
	 WHERE users.username = ?
	`

	var numberOfGames int64

	err = QueryRowWithRetry(m.DB, sqlStmt, []any{username}, []any{&numberOfGames})
	if err != nil {
		app.errorLog.Printf("Error in GetTileInfoFromPlayerID gameCount: %s\n", err.Error())
		return nil, err
	}

	pingStatus := time.Since(time.Unix(lastSeen, 0)) < 10*time.Second

	return &UserTileInfo{
		PlayerID:   playerID,
		Username:   username,
		PingStatus: pingStatus,
		JoinDate:   joinDate,
		LastSeen:   lastSeen,
		Ratings: Ratings{
			BulletRating:    bullet_rating,
			BlitzRating:     blitz_rating,
			RapidRating:     rapid_rating,
			ClassicalRating: classical_rating,
		},
		NumberOfGames: numberOfGames,
	}, nil

}
