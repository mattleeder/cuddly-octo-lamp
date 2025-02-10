package models

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	rowsLog  *log.Logger
}

var app *application

func init() {

	infoLog := log.New(os.Stdout, "DB INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "DB ERROR\t", log.Ldate|log.Ltime|log.Llongfile)
	rowsLog := log.New(os.Stdout, "DB ROW\t", 0)

	app = &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		rowsLog:  rowsLog,
	}

	app.infoLog.Println("RAN MODELS INIT")
}

func InitDatabase(driverName string, dataSourceName string) {
	os.Remove("./chess_site.db")

	db, err := sql.Open(driverName, dataSourceName)
	defer db.Close()
	if err != nil {
		app.errorLog.Fatal(err)
	}

	schemaPath := filepath.Join("internal", "models", "schema.sql")
	c, ioErr := os.ReadFile(schemaPath)
	if ioErr != nil {
		app.errorLog.Fatalf("%s", ioErr.Error())
	}
	sqlStmt := string(c)

	_, err = db.Exec(sqlStmt)
	if err != nil {
		app.errorLog.Fatalf("%q: %s\n", err, sqlStmt)
	}
}
