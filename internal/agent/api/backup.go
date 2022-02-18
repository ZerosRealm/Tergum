package api

import (
	"net/http"

	"zerosrealm.xyz/tergum/internal/agent/api/request"
)

func (api *API) Backup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req *request.Backup
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		go api.manager.Backup(req.Job, req.Repo, req.Backup)

		api.respond(w, r, nil, http.StatusNoContent)
	}
}
