package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

type application struct {
	errorLog  *log.Logger
	infoLog   *log.Logger
	secretKey []byte
}

var app *application

func main() {
	addr := flag.String("addr", ":8080", "HTTPS network address")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Llongfile)

	app = &application{
		errorLog:  errorLog,
		infoLog:   infoLog,
		secretKey: []byte("}\xa4\xc3\x85D\x89\xb75\xf0\xe6\xcf\xcaZ\x00k\x88\xe4\x8f\xd0\xd6\x95\x0e\xa6\xf9\xc2;!\xa2\xc4[\xca\x91"),
	}

	os.Remove("./chess_site.db")

	db := initDatabase()

	defer db.Close()

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	go matchmakingService()

	app.infoLog.Printf("Starting server on %s", *addr)
	err := srv.ListenAndServeTLS("cmd/web/localhost.crt", "cmd/web/localhost.key")
	if err != nil {
		errorLog.Fatal(err)
	}
}
