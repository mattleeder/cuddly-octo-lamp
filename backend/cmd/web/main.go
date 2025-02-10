package main

import (
	"burrchess/internal/models"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
)

type application struct {
	errorLog         *log.Logger
	infoLog          *log.Logger
	perfLog          *log.Logger
	secretKey        []byte
	liveMatches      *models.LiveMatchModel
	matchmakingQueue *models.MatchmakingQueueModel
}

var app *application

func main() {
	addr := flag.String("addr", ":8080", "HTTPS network address")
	dbDriverName := flag.String("db", "sqlite", "Database Driver Name")
	dbDataSourceName := flag.String("dsn", "./chess_site.db", "Database Data Source Name")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Llongfile)
	perfLog := log.New(os.Stdout, "PERF\t", log.Lshortfile)

	models.InitDatabase(*dbDriverName, *dbDataSourceName)
	db, err := sql.Open(*dbDriverName, *dbDataSourceName)
	defer db.Close()
	if err != nil {
		errorLog.Fatal(err)
	}

	app = &application{
		errorLog:         errorLog,
		infoLog:          infoLog,
		perfLog:          perfLog,
		secretKey:        []byte("}\xa4\xc3\x85D\x89\xb75\xf0\xe6\xcf\xcaZ\x00k\x88\xe4\x8f\xd0\xd6\x95\x0e\xa6\xf9\xc2;!\xa2\xc4[\xca\x91"),
		liveMatches:      &models.LiveMatchModel{DB: db},
		matchmakingQueue: &models.MatchmakingQueueModel{DB: db},
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	go matchmakingService()

	app.infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServeTLS("cmd/web/localhost.crt", "cmd/web/localhost.key")
	if err != nil {
		errorLog.Fatal(err)
	}
}
