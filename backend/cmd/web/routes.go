package main

import (
	"net/http"
	"net/http/pprof"
)

func (app *application) routes() *http.ServeMux {

	mux := http.NewServeMux()

	mux.HandleFunc("/", chessMoveValidationHandler)
	mux.HandleFunc("/getMoves", getChessMovesHandler)
	mux.HandleFunc("/makeMove", postChessMoveHandler)
	mux.HandleFunc("/joinQueue", joinQueueHandler)
	mux.HandleFunc("/listenformatch", matchFoundSSEHandler)
	mux.HandleFunc("/matchroom/{matchID}", getMatchStateHandler)
	mux.HandleFunc("/matchroom/{matchID}/ws", serveMatchroomWs)

	// Add the pprof routes
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

	return mux
}
