package api

import (
	"net/http"

	"zerosrealm.xyz/tergum/internal/agent/api/request"
)

func (api *API) Stop() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var req *request.Stop
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		go api.manager.Stop(req.Job)

		api.respond(w, r, nil, http.StatusNoContent)
	}
}
