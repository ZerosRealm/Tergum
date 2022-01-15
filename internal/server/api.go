package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"zerosrealm.xyz/tergum/internal/restic"
	"zerosrealm.xyz/tergum/internal/types"
)

func (srv *Server) getBackups() http.HandlerFunc {
	type response struct {
		Backups []*types.Backup `json:"backups"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		resp := response{
			Backups: make([]*types.Backup, 0),
		}

		resp.Backups = append(resp.Backups, savedData.Backups...)

		srv.respond(w, r, resp, 200)
	}
}

func (srv *Server) createBackup() http.HandlerFunc {
	type request struct {
		Target   int    `json:"target"`
		Source   string `json:"source"`
		Schedule string `json:"schedule"`
	}
	type response struct {
		Backup *types.Backup `json:"backup"`
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

		savedData.BackupIncrement++
		backup := &types.Backup{
			ID:       savedData.BackupIncrement,
			Target:   req.Target,
			Source:   req.Source,
			Schedule: req.Schedule,
			Exclude:  []string{},
		}

		savedData.Backups = append(savedData.Backups, backup)
		addSchedule(req.Schedule, srv.manager, backup)

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
		Backup *types.Backup `json:"backup"`
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

		id, err := strconv.Atoi(backupID)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		var foundBackup *types.Backup
		for _, backup := range savedData.Backups {
			if backup.ID == id {
				foundBackup = backup
				break
			}
		}

		status := http.StatusOK
		if foundBackup == nil {
			foundBackup = &types.Backup{
				ID: id,
			}
			status = http.StatusCreated
			savedData.BackupIncrement++
			savedData.Backups = append(savedData.Backups, foundBackup)
		}

		foundBackup.Target = req.Target
		foundBackup.Source = req.Source
		foundBackup.Schedule = req.Schedule
		foundBackup.Exclude = req.Exclude

		// TODO: validate schedule/cron syntax
		schedule := getSchedule(foundBackup.ID)
		if schedule == nil {
			addSchedule(foundBackup.Schedule, srv.manager, foundBackup)
		} else {
			schedule.newScheduler(foundBackup.Schedule)
		}

		srv.respond(w, r, response{Backup: foundBackup}, status)
	}
}

func (srv *Server) deleteBackup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		vars := mux.Vars(r)
		backupID := vars["id"]

		id, err := strconv.Atoi(backupID)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		index := -1
		for i, backup := range savedData.Backups {
			if backup.ID == id {
				index = i
				break
			}
		}

		if index == -1 {
			srv.error(w, r, fmt.Errorf("no backup with that ID"), http.StatusNotFound)
			return
		}

		savedData.Backups = append(savedData.Backups[:index], savedData.Backups[index+1:]...)
		removeSchedule(id)

		srv.respond(w, r, nil, http.StatusNoContent)
	}
}

func (srv *Server) getRepos() http.HandlerFunc {
	type response struct {
		Repos []*types.Repo `json:"repos"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		resp := response{
			Repos: make([]*types.Repo, 0),
		}

		resp.Repos = append(resp.Repos, savedData.Repos...)

		srv.respond(w, r, resp, http.StatusOK)
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
		Repo *types.Repo `json:"repo"`
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

		savedData.RepoIncrement++
		repo := &types.Repo{
			ID:       savedData.RepoIncrement,
			Name:     req.Name,
			Repo:     req.Repo,
			Password: req.Password,
			Settings: req.Settings,
		}
		savedData.Repos = append(savedData.Repos, repo)

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
		Repo *types.Repo `json:"repo"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		vars := mux.Vars(r)
		repoID := vars["id"]

		var req request
		err := srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(repoID)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		var foundRepo *types.Repo
		for _, repo := range savedData.Repos {
			if repo.ID == id {
				foundRepo = repo
				break
			}
		}

		status := http.StatusOK
		if foundRepo == nil {
			foundRepo = &types.Repo{
				ID: id,
			}
			status = http.StatusCreated
			savedData.RepoIncrement++
			savedData.Repos = append(savedData.Repos, foundRepo)
		}

		foundRepo.Name = req.Name
		foundRepo.Repo = req.Repo
		foundRepo.Password = req.Password
		foundRepo.Settings = req.Settings

		srv.respond(w, r, response{Repo: foundRepo}, status)
	}
}

func (srv *Server) deleteRepo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		vars := mux.Vars(r)
		repoID := vars["id"]

		id, err := strconv.Atoi(repoID)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		index := -1
		for i, repo := range savedData.Repos {
			if repo.ID == id {
				index = i
				break
			}
		}

		if index == -1 {
			srv.error(w, r, fmt.Errorf("no repo with that ID"), http.StatusNotFound)
			return
		}

		savedData.Repos = append(savedData.Repos[:index], savedData.Repos[index+1:]...)

		srv.respond(w, r, nil, http.StatusNoContent)
	}
}

func (srv *Server) getAgents() http.HandlerFunc {
	type response struct {
		Agents []*types.Agent `json:"agents"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		resp := response{
			Agents: make([]*types.Agent, 0),
		}

		resp.Agents = append(resp.Agents, savedData.Agents...)

		srv.respond(w, r, resp, http.StatusOK)
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
		Agent *types.Agent `json:"agent"`
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

		savedData.AgentIncrement++
		agent := &types.Agent{
			ID:   savedData.AgentIncrement,
			Name: req.Name,
			PSK:  req.PSK,
			IP:   req.IP,
			Port: req.Port,
		}
		savedData.Agents = append(savedData.Agents, agent)

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
		Agent *types.Agent `json:"agent"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		vars := mux.Vars(r)
		agentID := vars["id"]

		var req request
		err := srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(agentID)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		var foundAgent *types.Agent
		for _, agent := range savedData.Agents {
			if agent.ID == id {
				foundAgent = agent
				break
			}
		}

		status := http.StatusOK
		if foundAgent == nil {
			foundAgent = &types.Agent{
				ID: id,
			}
			status = http.StatusCreated
			savedData.AgentIncrement++
			savedData.Agents = append(savedData.Agents, foundAgent)
		}

		foundAgent.Name = req.Name
		foundAgent.PSK = req.PSK
		foundAgent.IP = req.IP
		foundAgent.Port = req.Port

		srv.respond(w, r, response{Agent: foundAgent}, status)
	}
}

func (srv *Server) deleteAgent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		vars := mux.Vars(r)
		agentID := vars["id"]

		id, err := strconv.Atoi(agentID)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		index := -1
		for i, agent := range savedData.Agents {
			if agent.ID == id {
				index = i
				break
			}
		}

		if index == -1 {
			srv.error(w, r, fmt.Errorf("no agent found with that ID"), http.StatusNotFound)
			return
		}

		savedData.Agents = append(savedData.Agents[:index], savedData.Agents[index+1:]...)

		srv.respond(w, r, nil, http.StatusNoContent)
	}
}

func (srv *Server) deleteSnapshot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		vars := mux.Vars(r)
		repoID := vars["id"]
		snapshot := vars["snapshot"]

		id, err := strconv.Atoi(repoID)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		var foundRepo *types.Repo
		for _, repo := range savedData.Repos {
			if repo.ID == id {
				foundRepo = repo
				break
			}
		}

		if foundRepo == nil {
			srv.error(w, r, fmt.Errorf("no repo found with that ID"), http.StatusNotFound)
			return
		}

		out, err := resticExe.Forget(foundRepo.Repo, foundRepo.Password, snapshot, nil, foundRepo.Settings...)
		if err != nil {
			srv.error(w, r, fmt.Errorf(string(out)), http.StatusInternalServerError)
			return
		}

		srv.respond(w, r, nil, http.StatusNoContent)
	}
}

func (srv *Server) getBackupAgents() http.HandlerFunc {
	type response struct {
		Agents []*types.Agent `json:"agents"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		vars := mux.Vars(r)
		backupID := vars["id"]

		id, err := strconv.Atoi(backupID)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		index := -1
		for i, backup := range savedData.Backups {
			if backup.ID == id {
				index = i
				break
			}
		}

		if index == -1 {
			srv.error(w, r, fmt.Errorf("no backup found with that ID"), http.StatusNotFound)
			return
		}

		agents, ok := savedData.BackupSubscribers[id]
		if !ok || agents == nil {
			srv.respond(w, r, response{Agents: make([]*types.Agent, 0)}, http.StatusOK)
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
		Agents []*types.Agent `json:"agents"`
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

		id, err := strconv.Atoi(backupID)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		var foundBackup *types.Backup
		for _, backup := range savedData.Backups {
			if backup.ID == id {
				foundBackup = backup
				break
			}
		}

		if foundBackup == nil {
			srv.error(w, r, fmt.Errorf("no backup found with that ID"), http.StatusNotFound)
			return
		}

		agents, ok := savedData.BackupSubscribers[id]
		if !ok || len(req.Agents) == 0 {
			agents = make([]*types.Agent, 0)
		}

		for _, agentID := range req.Agents {
			var foundAgent *types.Agent
			for _, agent := range savedData.Agents {
				if agent.ID == agentID {
					foundAgent = agent
					break
				}
			}
			if foundAgent == nil {
				srv.error(w, r, fmt.Errorf("no agent found with that ID"), http.StatusNotFound)
				return
			}

			agents = append(agents, foundAgent)
		}

		savedData.BackupSubscribers[id] = agents

		srv.respond(w, r, response{Agents: agents}, http.StatusCreated)
	}
}

func (srv *Server) getJobs() http.HandlerFunc {
	type response struct {
		Jobs map[string]*types.Job `json:"jobs"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		resp := response{Jobs: make(map[string]*types.Job)}
		for _, job := range srv.manager.Jobs {
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
		Type string     `json:"type"`
		Job  *types.Job `json:"job"`
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
		for _, agent := range savedData.Agents {
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
		err := srv.decode(w, r, &req)
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

		w.WriteHeader(http.StatusOK)
	}
}

func (srv *Server) getSnapshots() http.HandlerFunc {
	type response struct {
		Snapshots []restic.Snapshot `json:"snapshots"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		vars := mux.Vars(r)
		repoID := vars["id"]

		id, err := strconv.Atoi(repoID)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		var foundRepo *types.Repo
		for _, repo := range savedData.Repos {
			if repo.ID == id {
				foundRepo = repo
				break
			}
		}

		if foundRepo == nil {
			srv.error(w, r, fmt.Errorf("no repo found with that ID"), http.StatusNotFound)
			return
		}

		snapshots, err := resticExe.Snapshots(foundRepo.Repo, foundRepo.Password, foundRepo.Settings...)
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
		Job string `json:"job"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

		vars := mux.Vars(r)
		repoID := vars["id"]
		snapshot := vars["snapshot"]

		var req request
		err := srv.decode(w, r, &req)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(repoID)
		if err != nil {
			srv.error(w, r, err, http.StatusBadRequest)
			return
		}

		var foundRepo *types.Repo
		for _, repo := range savedData.Repos {
			if repo.ID == id {
				foundRepo = repo
				break
			}
		}

		if foundRepo == nil {
			srv.error(w, r, fmt.Errorf("no repo found with that ID"), http.StatusNotFound)
			return
		}

		var foundAgent *types.Agent
		for _, agent := range savedData.Agents {
			if agent.ID == req.Agent {
				foundAgent = agent
				break
			}
		}

		if foundAgent == nil {
			srv.error(w, r, fmt.Errorf("no agent found with that ID"), http.StatusNotFound)
			return
		}

		job := types.JobPacket{
			Type:  "restore",
			Agent: foundAgent,
			Repo:  foundRepo,
		}

		restoreJob := types.RestoreJob{
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
		log.Printf("enqueuing job %s for %s\n", jobID, foundAgent.Name)

		srv.respond(w, r, response{Job: jobID}, http.StatusOK)
	}
}

func (srv *Server) template() http.HandlerFunc {
	type request struct{}
	type response struct{}
	return func(w http.ResponseWriter, r *http.Request) {
		savedData.Mutex.Lock()
		defer savedData.Mutex.Unlock()

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
