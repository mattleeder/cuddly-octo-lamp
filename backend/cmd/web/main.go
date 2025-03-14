package main

import (
	"burrchess/internal/models"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/sqlite3store" //Sqlite3 ?
	"github.com/alexedwards/scs/v2"
)

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	perfLog        *log.Logger
	debugLog       *log.Logger
	secretKey      []byte
	liveMatches    *models.LiveMatchModel
	pastMatches    *models.PastMatchModel
	users          *models.UserModel
	userRatings    *models.UserRatingsModel
	dbTaskQueue    *models.TaskQueue
	sessionManager *scs.SessionManager
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
	debugLog := log.New(os.Stdout, "DEBUG\t", log.Lshortfile)

	models.InitDatabase(*dbDriverName, *dbDataSourceName)
	db, err := sql.Open(*dbDriverName, *dbDataSourceName)
	defer db.Close()
	if err != nil {
		errorLog.Fatal(err)
	}

	sessionManager := scs.New()
	sessionManager.Store = sqlite3store.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.IdleTimeout = 1 * time.Hour
	sessionManager.HashTokenInStore = true
	sessionManager.Cookie = scs.SessionCookie{
		Name:     "id",
		HttpOnly: true,
		Persist:  false,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	}

	app = &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		perfLog:        perfLog,
		debugLog:       debugLog,
		secretKey:      []byte("}\xa4\xc3\x85D\x89\xb75\xf0\xe6\xcf\xcaZ\x00k\x88\xe4\x8f\xd0\xd6\x95\x0e\xa6\xf9\xc2;!\xa2\xc4[\xca\x91"),
		liveMatches:    &models.LiveMatchModel{DB: db},
		pastMatches:    &models.PastMatchModel{DB: db},
		users:          &models.UserModel{DB: db},
		userRatings:    &models.UserRatingsModel{DB: db},
		dbTaskQueue:    models.DBTaskQueue,
		sessionManager: sessionManager,
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
