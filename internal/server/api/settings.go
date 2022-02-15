package api

import (
	"fmt"
	"net/http"
)

func (api *API) SettingsLoggingGet() http.HandlerFunc {
	type response struct {
		Level string `json:"level"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		resp := response{
			Level: api.log.GetLevel(),
		}

		api.respond(w, r, resp, http.StatusOK)
	}
}

func (api *API) SettingsLoggingSet() http.HandlerFunc {
	type request struct {
		Level string `json:"level"`
	}
	type response struct {
		Level string `json:"level"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		if !api.log.SetLevel(req.Level) {
			api.error(w, r, "Invalid level selected.", fmt.Errorf("invalid setting 'level', value '%s'", req.Level), http.StatusBadRequest)
			return
		}

		resp := response{
			Level: req.Level,
		}

		api.respond(w, r, resp, http.StatusOK)
	}
}
