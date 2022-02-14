package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"zerosrealm.xyz/tergum/internal/entities"
	"zerosrealm.xyz/tergum/internal/restic"
	manager "zerosrealm.xyz/tergum/internal/server/manager"
)

func (api *API) GetJobs(man *manager.Manager) http.HandlerFunc {
	type response struct {
		Jobs map[string]*entities.Job `json:"jobs"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		respJobs := make(map[string]*entities.Job, 0)
		jobs, err := api.services.JobSvc.GetAll()
		if err != nil {
			api.error(w, r, "Could not get jobs.", err, http.StatusInternalServerError)
			return
		}

		if jobs == nil {
			api.respond(w, r, &response{respJobs}, http.StatusOK)
			return
		}

		for _, job := range jobs {
			respJobs[job.ID] = job
		}

		api.respond(w, r, &response{respJobs}, http.StatusOK)
	}
}

func (api *API) StopJob(man *manager.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		jobID := vars["id"]

		job, err := api.services.JobSvc.Get([]byte(jobID))
		if err != nil {
			api.error(w, r, "Could not get job.", err, http.StatusInternalServerError)
			return
		}

		if job == nil {
			api.error(w, r, "No job found with that ID.", fmt.Errorf("no job found with that ID"), http.StatusNotFound)
			return
		}

		err = man.StopJob(job)
		if err != nil {
			api.error(w, r, "Could not stop job.", err, http.StatusInternalServerError)
			return
		}

		api.respond(w, r, nil, http.StatusNoContent)
	}
}

func (api *API) JobProgress(man *manager.Manager, resticExe *restic.Restic) http.HandlerFunc {
	type request struct {
		Msg json.RawMessage `json:"msg"`
	}

	type wsResponse struct {
		Type string        `json:"type"`
		Job  *entities.Job `json:"job"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		auth := strings.SplitN(authHeader, " ", 2)
		if len(auth) != 2 || strings.ToLower(auth[0]) != "psk" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		psk := auth[1]
		access := false

		agents, err := api.services.AgentSvc.GetAll()
		if err != nil {
			api.error(w, r, "Could not get agents.", err, http.StatusInternalServerError)
			return
		}

		if agents == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		for _, agent := range agents {
			if agent.PSK == psk {
				access = true
				break
			}
		}

		if !access {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		vars := mux.Vars(r)
		jobID := vars["id"]

		job, err := api.services.JobSvc.Get([]byte(jobID))
		if err != nil {
			api.error(w, r, "Could not get job.", err, http.StatusInternalServerError)
			return
		}

		if job == nil {
			api.error(w, r, "No job found with that ID.", fmt.Errorf("no job found with that ID"), http.StatusNotFound)
			return
		}

		var req request
		err = api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		man.UpdateJobProgress(job, req.Msg)

		wsResponse := wsResponse{
			Type: "job_progress",
			Job:  job,
		}

		jobJSON, err := json.Marshal(wsResponse)
		if err != nil {
			api.error(w, r, "Could not encode websocket message.", err, http.StatusInternalServerError)
			return
		}

		man.WriteWS([]byte(jobJSON))

		if job.Done {
			forgetPolicy, err := api.services.ForgetSvc.Get([]byte("0"))
			if err != nil {
				api.error(w, r, "Could not get forget policy.", err, http.StatusInternalServerError)
				return
			}

			if !forgetPolicy.Enabled {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			options := &restic.ForgetOptions{
				LastX:   forgetPolicy.LastX,
				Hourly:  forgetPolicy.Hourly,
				Daily:   forgetPolicy.Daily,
				Weekly:  forgetPolicy.Weekly,
				Monthly: forgetPolicy.Monthly,
				Yearly:  forgetPolicy.Yearly,
			}

			out, err := resticExe.Forget(job.Packet.Repo.Repo, job.Packet.Repo.Password, "", options)
			if err != nil {
				api.error(w, r, "Running forget policy returned errors.", fmt.Errorf("%s: %s", string(out), err), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (api *API) CreateJob(man *manager.Manager) http.HandlerFunc {
	type request struct {
		Backup int `json:"backup"`
	}
	type response struct {
		Jobs []*entities.Job `json:"jobs"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		// TODO: Does this mean it can only find the schedule of the first backup?
		// So a backup can only have one schedule?
		schedule := manager.GetSchedule(req.Backup)
		if schedule == nil {
			api.error(w, r, "Could not get schedule for backup.", fmt.Errorf("no schedule found for that backup"), http.StatusNotFound)
			return
		}

		jobs, err := schedule.Start()
		if err != nil {
			api.error(w, r, "Could not start backup.", err, http.StatusInternalServerError)
			return
		}

		if jobs == nil {
			api.respond(w, r, response{Jobs: make([]*entities.Job, 0)}, http.StatusOK)
			return
		}

		api.respond(w, r, response{Jobs: jobs}, http.StatusOK)
	}
}

func (api *API) JobError(man *manager.Manager) http.HandlerFunc {
	type wsResponse struct {
		Type  string `json:"type"`
		Error string `json:"error"`
		Msg   string `json:"msg"`
	}

	type request struct {
		Msg   string `json:"msg"`
		Error string `json:"error"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authorization")
		auth := strings.SplitN(authHeader, " ", 2)
		if len(auth) != 2 || strings.ToLower(auth[0]) != "psk" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		psk := auth[1]
		access := false

		agents, err := api.services.AgentSvc.GetAll()
		if err != nil {
			api.error(w, r, "Could not get agents.", err, http.StatusInternalServerError)
			return
		}

		if agents == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		for _, agent := range agents {
			if agent.PSK == psk {
				access = true
				break
			}
		}

		if !access {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		vars := mux.Vars(r)
		jobID := vars["id"]

		job, err := api.services.JobSvc.Get([]byte(jobID))
		if err != nil {
			api.error(w, r, "Could not get job.", err, http.StatusInternalServerError)
			return
		}

		if job == nil {
			api.error(w, r, "No job found with that ID.", fmt.Errorf("no job found with that ID"), http.StatusNotFound)
			return
		}

		job.Aborted = true

		var req request
		err = api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		wsResponse := wsResponse{
			Type:  "job_error",
			Error: req.Error,
			Msg:   req.Msg,
		}

		jobJSON, err := json.Marshal(wsResponse)
		if err != nil {
			api.error(w, r, "Could not encode websocket message.", err, http.StatusInternalServerError)
			return
		}

		man.WriteWS([]byte(jobJSON))

		w.WriteHeader(http.StatusOK)
	}
}
