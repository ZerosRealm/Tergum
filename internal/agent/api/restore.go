package api

import (
	"net/http"

	"zerosrealm.xyz/tergum/internal/agent/api/request"
)

func (api *API) Restore() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req *request.Restore
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		go api.manager.Restore(req.Job, req.Repo, req.Snapshot, req.Target, req.Include, req.Exclude)

		api.respond(w, r, nil, http.StatusNoContent)
	}
}
