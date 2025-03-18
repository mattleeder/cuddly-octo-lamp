package models

import (
	"database/sql"
	"errors"
	"math/rand"

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
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type UserLoginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
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

func (m *UserModel) InsertNew(username string, password string, options *NewUserOptions) error {
	defer m.LogAll()
	app.infoLog.Printf("Inserting new user: %v\n", username)

	if username == "" {
		app.errorLog.Printf("Username not given")
		return errors.New("username not given")
	}

	if username == "" {
		app.errorLog.Printf("Username not given")
		return errors.New("password not given")
	}

	var playerID = rand.Int63()
	hashedPassword, err := hashPassword(password)
	password = "" // Overwrite password to avoid accidental usage
	if err != nil {
		app.errorLog.Printf("Error generating password hash: %v\n", err.Error())
		return err
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

	_, err = stmtOne.Exec(playerID, username, hashedPassword, email)
	if err != nil {
		app.errorLog.Printf("Error inserting new user: %s\n", err.Error())
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			app.errorLog.Printf("insert users: unable to rollback: %s", rollbackErr)
		}
		return err
	}

	_, err = stmtTwo.Exec(playerID, username)
	if err != nil {
		app.errorLog.Printf("Error executing second statement: %v\n", err)
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			app.errorLog.Printf("insert user_ratings: unable to rollback: %v", rollbackErr)
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

func (m *UserModel) Authenticate(username string, password string) (playerID int64, authorized bool) {
	app.infoLog.Printf("Checking password for user: %v\n", username)

	sqlStmt := `
	select player_id, password from users where username = ?
	`
	row := m.DB.QueryRow(sqlStmt, password)
	var hashedPassword string
	err := row.Scan(&playerID, &hashedPassword)
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

	rows, err := m.DB.Query("select * from users;")
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
	var row *sql.Row

	if queryMode == qmUsername {
		sqlStmt += ` WHERE username = ?`
		row = m.DB.QueryRow(sqlStmt, username)
	} else if queryMode == qmPlayerID {
		sqlStmt += ` WHERE player_id = ?`
		row = m.DB.QueryRow(sqlStmt, playerID)
	} else {
		return UserClientSide{}, errors.New("queryMode unknown")
	}

	err := row.Scan(&_playerID, &_username, &join_date, &last_seen)
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
	var row *sql.Row

	if queryMode == qmUsername {
		sqlStmt += ` WHERE username = ?`
		row = m.DB.QueryRow(sqlStmt, username)
	} else if queryMode == qmPlayerID {
		sqlStmt += ` WHERE player_id = ?`
		row = m.DB.QueryRow(sqlStmt, playerID)
	} else {
		return UserServerSide{}, errors.New("queryMode unknown")
	}

	err := row.Scan(&_playerID, &_username, &password, &email, &join_date, &last_seen)
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
	var err error

	if queryMode == qmUsername {
		sqlStmt += ` WHERE username = ?`
		_, err = m.DB.Exec(sqlStmt, username)
	} else if queryMode == qmPlayerID {
		sqlStmt += ` WHERE player_id = ?`
		_, err = m.DB.Exec(sqlStmt, playerID)
	} else {
		return errors.New("queryMode unknown")
	}

	if err != nil {
		app.errorLog.Printf("Error updating last_seen: %v\n", err.Error())
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
	var err error

	if queryMode == qmUsername {
		sqlStmt += ` WHERE username = ?`
		_, err = m.DB.Exec(sqlStmt, username, email)
	} else if queryMode == qmPlayerID {
		sqlStmt += ` WHERE player_id = ?`
		_, err = m.DB.Exec(sqlStmt, playerID, email)
	} else {
		return errors.New("queryMode unknown")
	}

	if err != nil {
		app.errorLog.Printf("Error updating email: %v\n", err.Error())
	}

	return err
}

func (m *UserModel) UpdateEmailFromUsername(username string, email string) error {
	return m.updateEmail(email, username, 0, qmUsername)
}

func (m *UserModel) UpdateEmailFromPlayerID(playerID int64, email string) error {
	return m.updateEmail(email, "", playerID, qmPlayerID)
}
