package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
	"zerosrealm.xyz/tergum/internal/entities"
	manager "zerosrealm.xyz/tergum/internal/server/manager"
)

func (api *API) GetBackups() http.HandlerFunc {
	type response struct {
		Backups []*entities.Backup `json:"backups"`
	}
	return func(w http.ResponseWriter, r *http.Request) {

		backups, err := api.services.BackupSvc.GetAll()
		if err != nil {
			api.error(w, r, "Could not get backups.", err, http.StatusInternalServerError)
			return
		}

		if backups == nil {
			backups = make([]*entities.Backup, 0)
		}

		api.respond(w, r, &response{Backups: backups}, 200)
	}
}

func (api *API) CreateBackup(man *manager.Manager) http.HandlerFunc {
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
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		_, err = cron.ParseStandard(req.Schedule)
		if err != nil {
			api.error(w, r, "Invalid cron schedule.", err, http.StatusBadRequest)
			return
		}

		backup := &entities.Backup{
			Target:   req.Target,
			Source:   req.Source,
			Schedule: req.Schedule,
			Exclude:  []string{},
		}

		backup, err = api.services.BackupSvc.Create(backup)
		if err != nil {
			api.error(w, r, "Could not create backup.", err, http.StatusInternalServerError)
			return
		}

		man.AddSchedule(req.Schedule, backup.ID)

		r.Header.Add("Location", fmt.Sprintf("/backup/%d", backup.ID))
		api.respond(w, r, response{Backup: backup}, http.StatusCreated)
	}
}

func (api *API) UpdateBackup(man *manager.Manager) http.HandlerFunc {
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
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		backup, err := api.services.BackupSvc.Get([]byte(backupID))
		if err != nil {
			api.error(w, r, "Could not get backup.", err, http.StatusInternalServerError)
			return
		}

		if backup == nil {
			api.error(w, r, "No backup with that ID.", fmt.Errorf("no backup with that ID"), http.StatusNotFound)
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
			api.error(w, r, "Invalid cron schedule.", err, http.StatusBadRequest)
			return
		}

		backup.Target = req.Target
		backup.Source = req.Source
		backup.Schedule = req.Schedule
		backup.Exclude = req.Exclude

		backup, err = api.services.BackupSvc.Update(backup)
		if err != nil {
			api.error(w, r, "Could not update backup.", err, http.StatusInternalServerError)
			return
		}

		schedule := manager.GetSchedule(backup.ID)
		if schedule == nil {
			man.AddSchedule(backup.Schedule, backup.ID)
		} else {
			schedule.NewScheduler(backup.Schedule)
		}

		api.respond(w, r, response{Backup: backup}, status)
	}
}

func (api *API) DeleteBackup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		backupID := vars["id"]

		backup, err := api.services.BackupSvc.Get([]byte(backupID))
		if err != nil {
			api.error(w, r, "Could not get backup.", err, http.StatusInternalServerError)
			return
		}

		if backup == nil {
			api.error(w, r, "No backup with that ID.", fmt.Errorf("no backup with that ID"), http.StatusNotFound)
			return
		}

		err = api.services.BackupSvc.Delete([]byte(backupID))
		if err != nil {
			api.error(w, r, "Could not delete backup.", err, http.StatusInternalServerError)
			return
		}
		manager.RemoveSchedule(backup.ID)

		api.respond(w, r, nil, http.StatusNoContent)
	}
}

func (api *API) GetBackupAgents() http.HandlerFunc {
	type response struct {
		Agents []*entities.Agent `json:"agents"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		backupID := vars["id"]

		backup, err := api.services.BackupSvc.Get([]byte(backupID))
		if err != nil {
			api.error(w, r, "Could not get backup.", err, http.StatusInternalServerError)
			return
		}

		if backup == nil {
			api.error(w, r, "No backup found with that ID.", fmt.Errorf("no backup with that ID"), http.StatusNotFound)
			return
		}

		// TODO: Use a different service for this.
		// agents, ok := savedData.BackupSubscribers[backup.ID]
		// if !ok || agents == nil {
		// 	api.respond(w, r, response{Agents: make([]*entities.Agent, 0)}, http.StatusOK)
		// 	return
		// }

		subscribers, err := api.services.BackupSubSvc.Get([]byte(backupID))
		if err != nil {
			api.error(w, r, "Could not get backup subscribers.", err, http.StatusInternalServerError)
			return
		}

		if subscribers == nil || len(subscribers.AgentIDs) == 0 {
			api.respond(w, r, response{Agents: make([]*entities.Agent, 0)}, http.StatusOK)
			return
		}

		agents := make([]*entities.Agent, 0)
		for _, agentID := range subscribers.AgentIDs {
			agentID := strconv.Itoa(agentID)
			agent, err := api.services.AgentSvc.Get([]byte(agentID))
			if err != nil {
				api.error(w, r, "Could not get agent.", err, http.StatusInternalServerError)
				return
			}
			agents = append(agents, agent)
		}

		api.respond(w, r, response{Agents: agents}, http.StatusOK)
	}
}

func (api *API) UpdateBackupAgents() http.HandlerFunc {
	type request struct {
		Agents []int `json:"agents"`
	}
	type response struct {
		Agents []*entities.Agent `json:"agents"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		backupID := vars["id"]

		var req request
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		backup, err := api.services.BackupSvc.Get([]byte(backupID))
		if err != nil {
			api.error(w, r, "Could not get backup.", err, http.StatusInternalServerError)
			return
		}

		if backup == nil {
			api.error(w, r, "No backup found with that ID.", fmt.Errorf("no backup with that ID"), http.StatusNotFound)
			return
		}

		agentIDs := make([]int, 0)
		for _, agent := range req.Agents {
			agentIDs = append(agentIDs, agent)
		}

		backupSubscriber := &entities.BackupSubscribers{
			BackupID: backup.ID,
			AgentIDs: agentIDs,
		}

		backupSubscriber, err = api.services.BackupSubSvc.Update(backupSubscriber)
		if err != nil {
			api.error(w, r, "Could not update backup subscribers.", err, http.StatusInternalServerError)
			return
		}

		agents := make([]*entities.Agent, 0)
		for _, agentID := range backupSubscriber.AgentIDs {
			agent, err := api.services.AgentSvc.Get([]byte(strconv.Itoa(agentID)))
			if err != nil {
				api.error(w, r, "Could not get agent.", err, http.StatusInternalServerError)
				continue
			}

			if agent == nil {
				api.error(w, r, "No agent found with that ID.", fmt.Errorf("no agent with that ID"), http.StatusNotFound)
				continue
			}

			agents = append(agents, agent)
		}

		api.respond(w, r, response{Agents: agents}, http.StatusCreated)
	}
}
