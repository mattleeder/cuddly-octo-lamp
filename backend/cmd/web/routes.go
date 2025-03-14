package main

import (
	"net/http"
	"net/http/pprof"

	"github.com/alexedwards/scs/v2"
)

func wrapWithSessionManager(sm *scs.SessionManager, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sm.LoadAndSave(handler).ServeHTTP(w, r)
	}
}

func (app *application) routes() *http.ServeMux {

	mux := http.NewServeMux()

	mux.HandleFunc("/", wrapWithSessionManager(app.sessionManager, rootHandler))
	mux.HandleFunc("/getMoves", wrapWithSessionManager(app.sessionManager, getChessMovesHandler))
	mux.HandleFunc("/joinQueue", wrapWithSessionManager(app.sessionManager, joinQueueHandler))
	mux.HandleFunc("/listenformatch", wrapWithSessionManager(app.sessionManager, matchFoundSSEHandler))
	mux.HandleFunc("/matchroom/{matchID}/ws", wrapWithSessionManager(app.sessionManager, serveMatchroomWs))
	mux.HandleFunc("/getHighestEloMatch", wrapWithSessionManager(app.sessionManager, getHighestEloMatchHandler))

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
