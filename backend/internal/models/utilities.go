package models

import (
	"database/sql"
	"errors"
	"time"
)

func ExecStatementWithRetry(stmt *sql.Stmt, args ...any) (sql.Result, error) {

	delay := 50 * time.Millisecond
	maxRetries := 5

	var result sql.Result
	var err error

	for range maxRetries {
		result, err = stmt.Exec(args...)
		if err == nil {
			return result, nil
		} else if err.Error() == "database is locked (5) (SQLITE_BUSY)" {
			app.errorLog.Printf("%v, sleeping for %s\n", err.Error(), delay)
			time.Sleep(delay)
			continue
		} else {
			return nil, err
		}
	}

	return nil, errors.New("execStatementWithRetry: max retries exceeded")
}

func QueryWithRetry(DB *sql.DB, query string, args ...any) (*sql.Rows, error) {

	delay := 50 * time.Millisecond
	maxRetries := 5

	var rows *sql.Rows
	var err error

	for range maxRetries {
		rows, err = DB.Query(query, args...)
		if err == nil {
			return rows, nil
		} else if err.Error() == "database is locked (5) (SQLITE_BUSY)" {
			app.errorLog.Printf("%v, sleeping for %s\n", err.Error(), delay)
			time.Sleep(delay)
			continue
		} else {
			return nil, err
		}
	}

	return nil, errors.New("execStatementWithRetry: max retries exceeded")
}
