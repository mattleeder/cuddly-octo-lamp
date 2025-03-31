package main

import (
	"net/http"
	"net/http/pprof"

	"github.com/alexedwards/scs/v2"
)

func wrapWithSessionManager(sm *scs.SessionManager, handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// sessionID := sm.Token(r.Context())
		// app.infoLog.Println("Session ID:", sessionID)
		sm.LoadAndSave(handler).ServeHTTP(w, r)
	}
}

func withLogSessionSecureCorsChain(handlerFunc http.HandlerFunc) http.Handler {
	return app.logRequest(app.recoverPanic(wrapWithSessionManager(app.sessionManager, secureHeaders(corsHeaders(http.HandlerFunc(handlerFunc))))))
}

func withLogSecureCorsChain(handlerFunc http.HandlerFunc) http.Handler {
	return app.logRequest(app.recoverPanic(secureHeaders(corsHeaders(http.HandlerFunc(handlerFunc)))))
}

func (app *application) routes() http.Handler {

	mux := http.NewServeMux()

	mux.Handle("/", withLogSessionSecureCorsChain(rootHandler))
	mux.Handle("/getMoves", withLogSessionSecureCorsChain(getChessMovesHandler))
	mux.Handle("/joinQueue", withLogSessionSecureCorsChain(joinQueueHandler))
	mux.Handle("/matchroom/{matchID}/ws", withLogSessionSecureCorsChain(serveMatchroomWs))
	mux.Handle("/getHighestEloMatch", withLogSessionSecureCorsChain(getHighestEloMatchHandler))
	mux.Handle("/register", withLogSessionSecureCorsChain(registerUserHandler))
	mux.Handle("/login", withLogSessionSecureCorsChain(loginHandler))
	mux.Handle("/logout", withLogSessionSecureCorsChain(logoutHandler))
	mux.Handle("/validateSession", withLogSessionSecureCorsChain(validateSessionHandler))
	mux.Handle("/getAccountSettings", withLogSessionSecureCorsChain(getUserAccountSettingsHandler))

	mux.Handle("/userSearch", withLogSecureCorsChain(userSearchHandler))
	mux.Handle("/getTileInfo", withLogSecureCorsChain(getTileInfoHandler))
	mux.Handle("/getPastMatches", withLogSecureCorsChain(getPastMatchesListHandler))

	mux.Handle("/listenformatch", app.logRequest(app.recoverPanic(http.HandlerFunc(matchFoundSSEHandler))))

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
