package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	agentRequest "zerosrealm.xyz/tergum/internal/agent/api/request"
	"zerosrealm.xyz/tergum/internal/entity"
	"zerosrealm.xyz/tergum/internal/restic"
	manager "zerosrealm.xyz/tergum/internal/server/manager"
)

func (api *API) GetJobs(man *manager.Manager) http.HandlerFunc {
	type response struct {
		Jobs map[string]*entity.Job `json:"jobs"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		respJobs := make(map[string]*entity.Job, 0)
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

		backupRequest := job.Request.(agentRequest.Backup)
		if backupRequest.ID == "" {
			api.error(w, r, "No backup found with that ID.", fmt.Errorf("no backup found with that ID"), http.StatusNotFound)
			return
		}

		agents, err := api.services.BackupSubSvc.Get([]byte(strconv.Itoa(backupRequest.Backup.ID)))
		if err != nil {
			api.error(w, r, "Could not get backup subscriptions.", err, http.StatusInternalServerError)
			return
		}

		if agents == nil || len(agents.AgentIDs) == 0 {
			api.error(w, r, "No agents subscribed to that backup.", fmt.Errorf("no agents subscribed to that backup"), http.StatusNotFound)
			return
		}

		for _, agentID := range agents.AgentIDs {
			agent, err := api.services.AgentSvc.Get([]byte(strconv.Itoa(agentID)))
			if err != nil {
				api.error(w, r, "Could not get agent to send request to.", err, http.StatusInternalServerError)
				return
			}

			if agent == nil {
				api.error(w, r, "No agent found with that ID.", fmt.Errorf("no agent found with the ID '%d'", agentID), http.StatusNotFound)
				return
			}

			stopReq := &agentRequest.Stop{
				Job: agentRequest.Job{
					ID: job.ID,
				},
			}
			jobRequest := &entity.JobRequest{
				Type:  "stop",
				Agent: agent,

				Request: stopReq,
			}

			_, err = man.SendRequest(jobRequest, agent)
			if err != nil {
				api.error(w, r, "Could not stop job.", err, http.StatusInternalServerError)
				return
			}
		}

		api.respond(w, r, nil, http.StatusNoContent)
	}
}

func (api *API) JobProgress(man *manager.Manager, resticExe *restic.Restic) http.HandlerFunc {
	type request struct {
		Msg json.RawMessage `json:"msg"`
	}

	type wsResponse struct {
		Type string      `json:"type"`
		Job  *entity.Job `json:"job"`
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

		var authedAgent *entity.Agent
		for _, agent := range agents {
			if agent.PSK == psk {
				access = true
				authedAgent = agent
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

			backupRequest := job.Request.(agentRequest.Backup)
			if backupRequest.ID == "" {
				api.error(w, r, "Running forget policy failed.", fmt.Errorf("Backup request is invalid"), http.StatusBadRequest)
				return
			}

			forgetReq := &agentRequest.Forget{
				Repo:   backupRequest.Repo,
				Policy: forgetPolicy,
			}
			jobRequest := &entity.JobRequest{
				Type:  "forget",
				Agent: authedAgent,

				Request: forgetReq,
			}

			_, err = man.SendRequest(jobRequest, authedAgent)
			if err != nil {
				api.error(w, r, "Could not run forget policy.", err, http.StatusInternalServerError)
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
		Jobs []*entity.Job `json:"jobs"`
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
			api.respond(w, r, response{Jobs: make([]*entity.Job, 0)}, http.StatusOK)
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
