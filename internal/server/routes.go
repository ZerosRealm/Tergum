package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type errorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func (srv *Server) error(w http.ResponseWriter, r *http.Request, err error, status int) {
	data := errorResponse{
		Code:  status,
		Error: err.Error(),
	}

	srv.log.WithFields("method", r.Method, "path", r.URL.Path, "status", status, "src", r.RemoteAddr).Error(err)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		srv.log.Error("error: got error encoding response", err)
	}
}

func (srv *Server) respond(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	w.Header().Add("Content-Type", "application/json")

	if status != 200 {
		w.WriteHeader(status)
	}

	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}
	}

	srv.log.WithFields("method", r.Method, "path", r.URL.Path, "status", status, "src", r.RemoteAddr).Debug()
}

func (srv *Server) decode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

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

	api := srv.router.PathPrefix("/api/").Subrouter()

	api.Handle("/backup", srv.getBackups()).Methods("GET")
	api.Handle("/backup", srv.createBackup()).Methods("POST")
	// api.Handle("/backup/{id}", srv.getBackup()).Methods("GET")
	api.Handle("/backup/{id}", srv.updateBackup()).Methods("PUT")
	api.Handle("/backup/{id}", srv.deleteBackup()).Methods("DELETE")
	api.Handle("/backup/{id}/agent", srv.getBackupAgents()).Methods("GET")
	api.Handle("/backup/{id}/agent", srv.updateBackupAgents()).Methods("PUT")

	api.Handle("/agent", srv.getAgents()).Methods("GET")
	api.Handle("/agent", srv.createAgent()).Methods("POST")
	// api.Handle("/agent/{id}", srv.getAgent()).Methods("GET")
	api.Handle("/agent/{id}", srv.updateAgent()).Methods("PUT")
	api.Handle("/agent/{id}", srv.deleteAgent()).Methods("DELETE")

	api.Handle("/repo", srv.getRepos()).Methods("GET")
	api.Handle("/repo", srv.createRepo()).Methods("POST")
	// api.Handle("/repo/{id}", srv.getRepo()).Methods("GET")
	api.Handle("/repo/{id}", srv.updateRepo()).Methods("PUT")
	api.Handle("/repo/{id}", srv.deleteRepo()).Methods("DELETE")
	api.Handle("/repo/{id}/snapshot", srv.getSnapshots()).Methods("GET")
	api.Handle("/repo/{id}/snapshot/{snapshot}", srv.deleteSnapshot()).Methods("DELETE")
	api.Handle("/repo/{id}/snapshot/{snapshot}/restore", srv.restoreSnapshot()).Methods("POST")

	api.Handle("/job", srv.getJobs()).Methods("GET")
	api.Handle("/job", srv.createJob()).Methods("POST")
	// api.Handle("/job/{id}", srv.getJob()).Methods("GET")
	api.Handle("/job/{id}", srv.stopJob()).Methods("DELETE")
	api.Handle("/job/{id}/progress", srv.jobProgress()).Methods("POST")
	api.Handle("/job/{id}/error", srv.jobError()).Methods("POST")

	// api.Handle("/forget", srv.getJobs()).Methods("GET")
	// api.Handle("/forget", srv.createForget()).Methods("POST")
	api.Handle("/forget/{id}", srv.getForget()).Methods("GET")
	api.Handle("/forget/{id}", srv.updateForget()).Methods("PUT")

	srv.router.Use(mux.CORSMethodMiddleware(srv.router))
	srv.router.Use(cors)
}
