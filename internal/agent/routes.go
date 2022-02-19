package agent

import (
	"net/http"

	"github.com/gorilla/mux"
	"zerosrealm.xyz/tergum/internal/agent/api"
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

	api := api.New(srv.log.WithFields("component", "api"), srv.restic, srv.manager, srv.conf.PSK)

	apiRoute := srv.router.PathPrefix("/api/").Subrouter()
	apiRoute.Use(api.Authenticate())

	apiRoute.Handle("/backup", api.Backup()).Methods("POST")
	apiRoute.Handle("/stop", api.Stop()).Methods("POST")
	apiRoute.Handle("/snapshot", api.GetSnapshots()).Methods("POST")
	apiRoute.Handle("/snapshot", api.DeleteSnapshot()).Methods("DELETE")
	apiRoute.Handle("/snapshot/list", api.ListSnapshot()).Methods("POST")
	apiRoute.Handle("/snapshot/forget", api.Forget()).Methods("POST")
	apiRoute.Handle("/snapshot/restore", api.Restore()).Methods("POST")

	srv.router.Use(mux.CORSMethodMiddleware(srv.router))
	srv.router.Use(cors)
}
