package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"zerosrealm.xyz/tergum/internal/agent/manager"
	"zerosrealm.xyz/tergum/internal/log"
	"zerosrealm.xyz/tergum/internal/restic"
)

const msgDecodeError = "Could not decode request."

type API struct {
	log     *log.Logger
	restic  *restic.Restic
	manager *manager.Manager
	PSK     string
}

func New(logger *log.Logger, restic *restic.Restic, man *manager.Manager, PSK string) *API {
	return &API{
		log:     logger,
		restic:  restic,
		manager: man,
		PSK:     PSK,
	}
}

type errorResponse struct {
	Code    int    `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (api *API) error(w http.ResponseWriter, r *http.Request, msg string, err error, status int) {
	data := errorResponse{
		Code:    status,
		Error:   err.Error(),
		Message: msg,
	}

	api.log.WithFields("method", r.Method, "path", r.URL.Path, "status", status, "src", r.RemoteAddr).Error(err)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.log.Error("error: got error encoding response", err)
	}
}

func (api *API) respond(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	w.Header().Add("Content-Type", "application/json")

	if status != 200 {
		w.WriteHeader(status)
	}

	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			api.error(w, r, "Could not retrieve response!", err, http.StatusInternalServerError)
			return
		}
	}

	api.log.WithFields("method", r.Method, "path", r.URL.Path, "status", status, "src", r.RemoteAddr).Debug(fmt.Sprintf("%s %s - %d", r.Method, r.URL.Path, status))
}

func (api *API) decode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (api *API) Authenticate() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-PSK") != api.PSK {
				api.error(w, r, "Forbidden", fmt.Errorf("incorrect PSK"), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (api *API) template() http.HandlerFunc {
	type request struct{}
	type response struct{}
	return func(w http.ResponseWriter, r *http.Request) {
		// vars := mux.Vars(r)
		// name := vars["project"]

		// var req request
		// err := api.decode(w, r, &req)
		// if err != nil {
		// 	api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
		// 	return
		// }
	}
}
