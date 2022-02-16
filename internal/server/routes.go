package server

import (
	"net/http"

	"github.com/gorilla/mux"

	"zerosrealm.xyz/tergum/internal/server/api"
)

func corsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		next.ServeHTTP(w, r)
	})
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv.router.ServeHTTP(w, r)
}

func (srv *Server) routes() {
	srv.router.NewRoute().HandlerFunc(corsHandler).Methods("OPTIONS")
	srv.router.StrictSlash(true)

	api := api.New(srv.log.WithFields("component", "api"), srv.services)

	apiRoute := srv.router.PathPrefix("/api/").Subrouter()

	apiRoute.Handle("/backup", api.GetBackups()).Methods("GET")
	apiRoute.Handle("/backup", api.CreateBackup(srv.manager)).Methods("POST")
	// apiRoute.Handle("/backup/{id}", srv.getBackup()).Methods("GET")
	apiRoute.Handle("/backup/{id}", api.UpdateBackup(srv.manager)).Methods("PUT")
	apiRoute.Handle("/backup/{id}", api.DeleteBackup()).Methods("DELETE")
	apiRoute.Handle("/backup/{id}/agent", api.GetBackupAgents()).Methods("GET")
	apiRoute.Handle("/backup/{id}/agent", api.UpdateBackupAgents()).Methods("PUT")

	apiRoute.Handle("/agent", api.GetAgents()).Methods("GET")
	apiRoute.Handle("/agent", api.CreateAgent()).Methods("POST")
	// apiRoute.Handle("/agent/{id}", srv.getAgent()).Methods("GET")
	apiRoute.Handle("/agent/{id}", api.UpdateAgent()).Methods("PUT")
	apiRoute.Handle("/agent/{id}", api.DeleteAgent()).Methods("DELETE")

	apiRoute.Handle("/repo", api.GetRepos()).Methods("GET")
	apiRoute.Handle("/repo", api.CreateRepo()).Methods("POST")
	// apiRoute.Handle("/repo/{id}", srv.getRepo()).Methods("GET")
	apiRoute.Handle("/repo/{id}", api.UpdateRepo()).Methods("PUT")
	apiRoute.Handle("/repo/{id}", api.DeleteRepo()).Methods("DELETE")
	apiRoute.Handle("/repo/{id}/snapshot", api.GetSnapshots(resticExe)).Methods("GET")
	apiRoute.Handle("/repo/{id}/snapshot/{snapshot}", api.DeleteSnapshot(resticExe)).Methods("DELETE")
	apiRoute.Handle("/repo/{id}/snapshot/{snapshot}/restore", api.RestoreSnapshot(srv.manager)).Methods("POST")

	apiRoute.Handle("/job", api.GetJobs(srv.manager)).Methods("GET")
	apiRoute.Handle("/job", api.CreateJob(srv.manager)).Methods("POST")
	// apiRoute.Handle("/job/{id}", srv.getJob()).Methods("GET")
	apiRoute.Handle("/job/{id}", api.StopJob(srv.manager)).Methods("DELETE")
	apiRoute.Handle("/job/{id}/progress", api.JobProgress(srv.manager, resticExe)).Methods("POST")
	apiRoute.Handle("/job/{id}/error", api.JobError(srv.manager)).Methods("POST")

	// apiRoute.Handle("/forget", srv.getJobs()).Methods("GET")
	// apiRoute.Handle("/forget", srv.createForget()).Methods("POST")
	apiRoute.Handle("/forget/{id}", api.GetForget()).Methods("GET")
	apiRoute.Handle("/forget/{id}", api.UpdateForget()).Methods("PUT")

	apiRoute.Handle("/setting/logging", api.SettingsLoggingGet()).Methods("GET")
	apiRoute.Handle("/setting/logging", api.SettingsLoggingSet()).Methods("PUT")

	apiRoute.Handle("/log", api.GetLogs()).Methods("GET")

	srv.router.Use(mux.CORSMethodMiddleware(srv.router))
	srv.router.Use(cors)
}
