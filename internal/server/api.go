package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
	"zerosrealm.xyz/tergum/internal/entities"
	"zerosrealm.xyz/tergum/internal/restic"
)

func (srv *Server) getBackups() http.HandlerFunc {
	type response struct {
		Backups []*entities.Backup `json:"backups"`
	}
	return func(w http.ResponseWriter, r *http.Request) {

		backups, err := srv.services.backupSvc.GetAll()
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if backups == nil {
			backups = make([]*entities.Backup, 0)
		}

		srv.respond(w, r, &response{Backups: backups}, 200)
	}
}

func (srv *Server) createBackup() http.HandlerFunc {
	type request struct {
		Target   int    `json:"target"`
		Source   string `json:"source"`
		Schedule string `json:"schedule"`
	}
	type response struct {
		Backup *entities.Backup `json:"backup"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		err := srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		_, err = cron.ParseStandard(req.Schedule)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		backup := &entities.Backup{
			Target:   req.Target,
			Source:   req.Source,
			Schedule: req.Schedule,
			Exclude:  []string{},
		}

		backup, err = srv.services.backupSvc.Create(backup)
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		srv.manager.addSchedule(req.Schedule, backup.ID)

		r.Header.Add("Location", fmt.Sprintf("/backup/%d", backup.ID))
		srv.respond(w, r, response{Backup: backup}, http.StatusCreated)
	}
}

func (srv *Server) updateBackup() http.HandlerFunc {
	type request struct {
		Target   int      `json:"target"`
		Source   string   `json:"source"`
		Schedule string   `json:"schedule"`
		Exclude  []string `json:"exclude"`
	}
	type response struct {
		Backup *entities.Backup `json:"backup"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		backupID := vars["id"]

		var req request
		err := srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		backup, err := srv.services.backupSvc.Get([]byte(backupID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if backup == nil {
			srv.error(w, r, fmt.Errorf("no backup with that ID"), http.StatusNotFound)
			return
		}

		status := http.StatusOK
		// TODO: Create a backup with the given ID if it does not exist
		// if backup == nil {
		// 	backup = &entities.Backup{
		// 		ID: id,
		// 	}
		// 	status = http.StatusCreated
		// 	savedData.Backups = append(savedData.Backups, foundBackup)
		// }

		_, err = cron.ParseStandard(req.Schedule)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		backup.Target = req.Target
		backup.Source = req.Source
		backup.Schedule = req.Schedule
		backup.Exclude = req.Exclude

		backup, err = srv.services.backupSvc.Update(backup)
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		schedule := getSchedule(backup.ID)
		if schedule == nil {
			srv.manager.addSchedule(backup.Schedule, backup.ID)
		} else {
			schedule.newScheduler(backup.Schedule)
		}

		srv.respond(w, r, response{Backup: backup}, status)
	}
}

func (srv *Server) deleteBackup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		backupID := vars["id"]

		backup, err := srv.services.backupSvc.Get([]byte(backupID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if backup == nil {
			srv.error(w, r, fmt.Errorf("no backup with that ID"), http.StatusNotFound)
			return
		}

		err = srv.services.backupSvc.Delete([]byte(backupID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}
		removeSchedule(backup.ID)

		srv.respond(w, r, nil, http.StatusNoContent)
	}
}

func (srv *Server) getRepos() http.HandlerFunc {
	type response struct {
		Repos []*entities.Repo `json:"repos"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		repos, err := srv.services.repoSvc.GetAll()
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if repos == nil {
			repos = make([]*entities.Repo, 0)
		}

		srv.respond(w, r, &response{Repos: repos}, http.StatusOK)
	}
}

func (srv *Server) createRepo() http.HandlerFunc {
	type request struct {
		Name     string   `json:"name"`
		Repo     string   `json:"repo"`
		Password string   `json:"password"`
		Settings []string `json:"settings"`
	}
	type response struct {
		Repo *entities.Repo `json:"repo"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		err := srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		repo := &entities.Repo{
			Name:     req.Name,
			Repo:     req.Repo,
			Password: req.Password,
			Settings: req.Settings,
		}

		repo, err = srv.services.repoSvc.Create(repo)
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		r.Header.Add("Location", fmt.Sprintf("/repo/%d", repo.ID))
		srv.respond(w, r, response{Repo: repo}, http.StatusCreated)
	}
}

func (srv *Server) updateRepo() http.HandlerFunc {
	type request struct {
		Name     string   `json:"name"`
		Repo     string   `json:"repo"`
		Password string   `json:"password"`
		Settings []string `json:"settings"`
	}
	type response struct {
		Repo *entities.Repo `json:"repo"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		repoID := vars["id"]

		var req request
		err := srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		repo, err := srv.services.repoSvc.Get([]byte(repoID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if repo == nil {
			srv.error(w, r, fmt.Errorf("no repo with that ID"), http.StatusNotFound)
			return
		}

		status := http.StatusOK
		// TODO: Create a repo with the given ID if it does not exist
		// if foundRepo == nil {
		// 	foundRepo = &entities.Repo{
		// 		ID: id,
		// 	}
		// 	status = http.StatusCreated
		// 	savedData.RepoIncrement++
		// 	savedData.Repos = append(savedData.Repos, foundRepo)
		// }

		repo.Name = req.Name
		repo.Repo = req.Repo
		repo.Password = req.Password
		repo.Settings = req.Settings

		repo, err = srv.services.repoSvc.Update(repo)
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		srv.respond(w, r, response{Repo: repo}, status)
	}
}

func (srv *Server) deleteRepo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		repoID := vars["id"]

		repo, err := srv.services.repoSvc.Get([]byte(repoID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if repo == nil {
			srv.error(w, r, fmt.Errorf("no repo with that ID"), http.StatusNotFound)
			return
		}

		err = srv.services.repoSvc.Delete([]byte(repoID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		srv.respond(w, r, nil, http.StatusNoContent)
	}
}

func (srv *Server) getAgents() http.HandlerFunc {
	type response struct {
		Agents []*entities.Agent `json:"agents"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		agents, err := srv.services.agentSvc.GetAll()
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if agents == nil {
			agents = make([]*entities.Agent, 0)
		}

		srv.respond(w, r, &response{Agents: agents}, http.StatusOK)
	}
}

func (srv *Server) createAgent() http.HandlerFunc {
	type request struct {
		Name string `json:"name"`
		PSK  string `json:"psk"`
		IP   string `json:"ip"`
		Port int    `json:"port"`
	}
	type response struct {
		Agent *entities.Agent `json:"agent"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		err := srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		agent := &entities.Agent{
			Name: req.Name,
			PSK:  req.PSK,
			IP:   req.IP,
			Port: req.Port,
		}

		agent, err = srv.services.agentSvc.Create(agent)
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		r.Header.Add("Location", fmt.Sprintf("/agent/%d", agent.ID))
		srv.respond(w, r, response{Agent: agent}, http.StatusCreated)
	}
}

func (srv *Server) updateAgent() http.HandlerFunc {
	type request struct {
		Name string `json:"name"`
		PSK  string `json:"psk"`
		IP   string `json:"ip"`
		Port int    `json:"port"`
	}
	type response struct {
		Agent *entities.Agent `json:"agent"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		agentID := vars["id"]

		var req request
		err := srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		agent, err := srv.services.agentSvc.Get([]byte(agentID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if agent == nil {
			srv.error(w, r, fmt.Errorf("no agent with that ID"), http.StatusNotFound)
			return
		}

		status := http.StatusOK
		// TODO: Create an agent with the given ID if it does not exist
		// if foundAgent == nil {
		// 	foundAgent = &entities.Agent{
		// 		ID: id,
		// 	}
		// 	status = http.StatusCreated
		// 	savedData.AgentIncrement++
		// 	savedData.Agents = append(savedData.Agents, foundAgent)
		// }

		agent.Name = req.Name
		agent.PSK = req.PSK
		agent.IP = req.IP
		agent.Port = req.Port

		agent, err = srv.services.agentSvc.Update(agent)
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		srv.respond(w, r, response{Agent: agent}, status)
	}
}

func (srv *Server) deleteAgent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		agentID := vars["id"]

		agent, err := srv.services.agentSvc.Get([]byte(agentID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if agent == nil {
			srv.error(w, r, fmt.Errorf("no agent with that ID"), http.StatusNotFound)
			return
		}

		err = srv.services.agentSvc.Delete([]byte(agentID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		srv.respond(w, r, nil, http.StatusNoContent)
	}
}

func (srv *Server) deleteSnapshot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		repoID := vars["id"]
		snapshot := vars["snapshot"]

		repo, err := srv.services.repoSvc.Get([]byte(repoID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if repo == nil {
			srv.error(w, r, fmt.Errorf("no repo with that ID"), http.StatusNotFound)
			return
		}

		out, err := resticExe.Forget(repo.Repo, repo.Password, snapshot, nil, repo.Settings...)
		if err != nil {
			srv.error(w, r, fmt.Errorf(string(out)), http.StatusInternalServerError)
			return
		}

		srv.respond(w, r, nil, http.StatusNoContent)
	}
}

func (srv *Server) getBackupAgents() http.HandlerFunc {
	type response struct {
		Agents []*entities.Agent `json:"agents"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		vars := mux.Vars(r)
		backupID := vars["id"]

		backup, err := srv.services.backupSvc.Get([]byte(backupID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if backup == nil {
			srv.error(w, r, fmt.Errorf("no backup with that ID"), http.StatusNotFound)
			return
		}

		// TODO: Use a different service for this.
		agents, ok := savedData.BackupSubscribers[backup.ID]
		if !ok || agents == nil {
			srv.respond(w, r, response{Agents: make([]*entities.Agent, 0)}, http.StatusOK)
			return
		}

		srv.respond(w, r, response{Agents: agents}, http.StatusOK)
	}
}

func (srv *Server) updateBackupAgents() http.HandlerFunc {
	type request struct {
		Agents []int `json:"agents"`
	}
	type response struct {
		Agents []*entities.Agent `json:"agents"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		vars := mux.Vars(r)
		backupID := vars["id"]

		var req request
		err := srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		backup, err := srv.services.backupSvc.Get([]byte(backupID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if backup == nil {
			srv.error(w, r, fmt.Errorf("no backup with that ID"), http.StatusNotFound)
			return
		}
		subscribedAgents, ok := savedData.BackupSubscribers[backup.ID]
		if !ok || len(req.Agents) == 0 {
			subscribedAgents = make([]*entities.Agent, 0)
		}

		for _, agentID := range req.Agents {
			agent, err := srv.services.agentSvc.Get([]byte(strconv.Itoa(agentID)))
			if err != nil {
				srv.error(w, r, err, http.StatusInternalServerError)
				return
			}

			if agent == nil {
				srv.error(w, r, fmt.Errorf("no agent with the ID %d", agent.ID), http.StatusNotFound)
				return
			}

			subscribedAgents = append(subscribedAgents, agent)
		}

		savedData.BackupSubscribers[backup.ID] = subscribedAgents

		srv.respond(w, r, response{Agents: subscribedAgents}, http.StatusCreated)
	}
}

func (srv *Server) getJobs() http.HandlerFunc {
	type response struct {
		Jobs map[string]*entities.Job `json:"jobs"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		resp := response{Jobs: make(map[string]*entities.Job)}
		for _, job := range srv.manager.jobs {
			resp.Jobs[job.ID] = job
		}

		srv.respond(w, r, resp, http.StatusOK)
	}
}

func (srv *Server) stopJob() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		vars := mux.Vars(r)
		jobID := vars["id"]

		job := srv.manager.getJob(jobID)

		if job == nil {
			srv.error(w, r, fmt.Errorf("no job found with that ID"), http.StatusNotFound)
			return
		}

		err := srv.manager.stopJob(job)
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		srv.respond(w, r, nil, http.StatusNoContent)
	}
}

func (srv *Server) jobProgress() http.HandlerFunc {
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

		agents, err := srv.services.agentSvc.GetAll()
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
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

		job := srv.manager.getJob(jobID)

		if job == nil {
			srv.error(w, r, fmt.Errorf("no job found with that ID"), http.StatusNotFound)
			return
		}

		var req request
		err = srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		srv.manager.updateJobProgress(job, req.Msg)

		wsResponse := wsResponse{
			Type: "job_progress",
			Job:  job,
		}

		jobJSON, err := json.Marshal(wsResponse)
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		srv.manager.WriteWS([]byte(jobJSON))

		if job.Done {
			forgetPolicy, err := srv.services.forgetSvc.Get([]byte("0"))
			if err != nil {
				srv.error(w, r, err, http.StatusInternalServerError)
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

			_, err = resticExe.Forget(job.Packet.Repo.Repo, job.Packet.Repo.Password, "", options)
			if err != nil {
				srv.error(w, r, err, http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (srv *Server) getSnapshots() http.HandlerFunc {
	type response struct {
		Snapshots []restic.Snapshot `json:"snapshots"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		repoID := vars["id"]

		repo, err := srv.services.repoSvc.Get([]byte(repoID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if repo == nil {
			srv.error(w, r, fmt.Errorf("no repo found with that ID"), http.StatusNotFound)
			return
		}

		snapshots, err := resticExe.Snapshots(repo.Repo, repo.Password, repo.Settings...)
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}
		for i, snapshot := range snapshots {
			if len(snapshot.Tags) == 0 {
				snapshots[i].Tags = make([]string, 0)
			}
		}

		srv.respond(w, r, response{Snapshots: snapshots}, http.StatusOK)
	}
}

func (srv *Server) restoreSnapshot() http.HandlerFunc {
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
		err := srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		repo, err := srv.services.repoSvc.Get([]byte(repoID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if repo == nil {
			srv.error(w, r, fmt.Errorf("no repo found with that ID"), http.StatusNotFound)
			return
		}

		agent, err := srv.services.agentSvc.Get([]byte(strconv.Itoa(req.Agent)))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		job := entities.JobPacket{
			Type:  "restore",
			Repo:  repo,
			Agent: agent,
		}

		restoreJob := entities.RestoreJob{
			Snapshot: snapshot,
			Target:   req.Dest,
			Include:  req.Include,
			Exclude:  req.Exclude,
		}

		jobID, err := srv.manager.NewJob(&job, &restoreJob)
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}
		srv.log.Debug("Enqueuing job %s for %s", jobID, agent.Name)

		newJob := srv.manager.getJob(jobID)
		if newJob == nil {
			srv.error(w, r, fmt.Errorf("no job found with that ID"), http.StatusNotFound)
			return
		}

		srv.respond(w, r, response{Job: newJob}, http.StatusOK)
	}
}

func (srv *Server) createJob() http.HandlerFunc {
	type request struct {
		Backup int `json:"backup"`
	}
	type response struct {
		Jobs []*entities.Job `json:"jobs"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		var req request
		err := srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		schedule := getSchedule(req.Backup)
		if schedule == nil {
			srv.error(w, r, fmt.Errorf("no schedule found with that ID"), http.StatusNotFound)
			return
		}

		jobIDs, err := schedule.start()
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		jobs := make([]*entities.Job, len(jobIDs))
		for i, jobID := range jobIDs {
			jobs[i] = srv.manager.getJob(jobID)
		}

		srv.respond(w, r, response{Jobs: jobs}, http.StatusOK)
	}
}

func (srv *Server) jobError() http.HandlerFunc {
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

		agents, err := srv.services.agentSvc.GetAll()
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
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

		job := srv.manager.getJob(jobID)

		if job == nil {
			srv.error(w, r, fmt.Errorf("no job found with that ID"), http.StatusNotFound)
			return
		}

		job.Aborted = true

		var req request
		err = srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		wsResponse := wsResponse{
			Type:  "job_error",
			Error: req.Error,
			Msg:   req.Msg,
		}

		jobJSON, err := json.Marshal(wsResponse)
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		srv.manager.WriteWS([]byte(jobJSON))

		w.WriteHeader(http.StatusOK)
	}
}

func (srv *Server) getForget() http.HandlerFunc {
	type request struct{}
	type response struct {
		Forget *entities.Forget `json:"forget"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		forgetID := vars["id"]

		forget, err := srv.services.forgetSvc.Get([]byte(forgetID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if forget == nil {
			srv.error(w, r, fmt.Errorf("no forget found with that ID"), http.StatusNotFound)
			return
		}

		srv.respond(w, r, response{Forget: forget}, http.StatusOK)
	}
}

func (srv *Server) updateForget() http.HandlerFunc {
	type request struct {
		entities.Forget
	}
	type response struct {
		Forget *entities.Forget `json:"forget"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		forgetID := vars["id"]

		var req request
		err := srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		forget, err := srv.services.forgetSvc.Get([]byte(forgetID))
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		if forget == nil {
			srv.error(w, r, fmt.Errorf("no forget found with that ID"), http.StatusNotFound)
			return
		}

		forget.Enabled = req.Enabled
		forget.LastX = req.LastX
		forget.Hourly = req.Hourly
		forget.Daily = req.Daily
		forget.Weekly = req.Weekly
		forget.Monthly = req.Monthly
		forget.Yearly = req.Yearly

		forget, err = srv.services.forgetSvc.Update(forget)
		if err != nil {
			srv.error(w, r, err, http.StatusInternalServerError)
			return
		}

		srv.respond(w, r, response{Forget: forget}, http.StatusOK)
	}
}

func (srv *Server) template() http.HandlerFunc {
	type request struct{}
	type response struct{}
	return func(w http.ResponseWriter, r *http.Request) {
		// vars := mux.Vars(r)
		// name := vars["project"]

		// var req request
		// err := srv.decode(w, r, &req)
		// if err != nil {
		// 	srv.error(w, r, err, http.StatusBadRequest)
		// 	return
		// }
	}
}
