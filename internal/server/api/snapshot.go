package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"zerosrealm.xyz/tergum/internal/entities"
	"zerosrealm.xyz/tergum/internal/restic"
	manager "zerosrealm.xyz/tergum/internal/server/manager"
)

func (api *API) GetSnapshots(resticExe *restic.Restic) http.HandlerFunc {
	type response struct {
		Snapshots []restic.Snapshot `json:"snapshots"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		repoID := vars["id"]

		repo, err := api.services.RepoSvc.Get([]byte(repoID))
		if err != nil {
			api.error(w, r, "Could not get repository.", err, http.StatusInternalServerError)
			return
		}

		if repo == nil {
			api.error(w, r, "No repository found with that ID.", fmt.Errorf("no repo found with that ID"), http.StatusNotFound)
			return
		}

		snapshots, err := resticExe.Snapshots(repo.Repo, repo.Password, repo.Settings...)
		if err != nil {
			api.error(w, r, "Could not get snapshots.", err, http.StatusInternalServerError)
			return
		}
		for i, snapshot := range snapshots {
			if len(snapshot.Tags) == 0 {
				snapshots[i].Tags = make([]string, 0)
			}
		}

		api.respond(w, r, response{Snapshots: snapshots}, http.StatusOK)
	}
}

func (api *API) DeleteSnapshot(resticExe *restic.Restic) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		repoID := vars["id"]
		snapshot := vars["snapshot"]

		repo, err := api.services.RepoSvc.Get([]byte(repoID))
		if err != nil {
			api.error(w, r, "Could not get repository.", err, http.StatusInternalServerError)
			return
		}

		if repo == nil {
			api.error(w, r, "No repository with that ID.", fmt.Errorf("no repo with that ID"), http.StatusNotFound)
			return
		}

		out, err := resticExe.Forget(repo.Repo, repo.Password, snapshot, nil, repo.Settings...)
		if err != nil {
			api.error(w, r, "Running forget policy returned errors.", fmt.Errorf("%s: %s", string(out), err), http.StatusInternalServerError)
			return
		}

		api.respond(w, r, nil, http.StatusNoContent)
	}
}

func (api *API) RestoreSnapshot(manager *manager.Manager) http.HandlerFunc {
	type request struct {
		Agent   int    `json:"agent"`
		Dest    string `json:"destination"`
		Include string `json:"include"`
		Exclude string `json:"exclude"`
	}
	type response struct {
		Job *entities.Job `json:"job"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		repoID := vars["id"]
		snapshot := vars["snapshot"]

		var req request
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		repo, err := api.services.RepoSvc.Get([]byte(repoID))
		if err != nil {
			api.error(w, r, "Could not get repository.", err, http.StatusInternalServerError)
			return
		}

		if repo == nil {
			api.error(w, r, "No repository found with that ID.", fmt.Errorf("no repo found with that ID"), http.StatusNotFound)
			return
		}

		agent, err := api.services.AgentSvc.Get([]byte(strconv.Itoa(req.Agent)))
		if err != nil {
			api.error(w, r, "Could not get agent.", err, http.StatusInternalServerError)
			return
		}

		jobPacket := &entities.JobPacket{
			Type:  "restore",
			Repo:  repo,
			Agent: agent,
		}

		restoreJob := &entities.RestoreJob{
			Snapshot: snapshot,
			Target:   req.Dest,
			Include:  req.Include,
			Exclude:  req.Exclude,
		}

		job, err := manager.NewJob(jobPacket, restoreJob)
		if err != nil {
			api.error(w, r, "Could not create job.", err, http.StatusInternalServerError)
			return
		}
		api.log.Debug("Enqueuing job %s for %s", job.ID, agent.Name)

		api.respond(w, r, response{Job: job}, http.StatusOK)
	}
}
