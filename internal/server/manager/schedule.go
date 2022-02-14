package server

import (
	"fmt"
	"strconv"

	"github.com/robfig/cron/v3"
	"zerosrealm.xyz/tergum/internal/entities"
)

type schedule struct {
	BackupID  int
	Schedule  string
	Scheduler *cron.Cron

	manager *Manager
}

var schedules = []*schedule{}

func (man *Manager) BuildSchedules() {
	man.log.Debug("Building schedules")
	backups, err := man.services.BackupSvc.GetAll()
	if err != nil {
		man.log.Error("buildSchedules: could not get backups", err)
		return
	}

	for _, backup := range backups {
		man.log.Debug("Adding schedule for backup", fmt.Sprintf("#%d", backup.ID))
		man.AddSchedule(backup.Schedule, backup.ID)
	}
}

func (schedule *schedule) Start() ([]*entities.Job, error) {
	backup, err := schedule.manager.services.BackupSvc.Get([]byte(strconv.Itoa(schedule.BackupID)))
	if err != nil {
		return nil, err
	}

	schedule.manager.log.WithFields("backup", backup.ID).Debug("Starting backup")

	subcribers, err := schedule.manager.services.BackupSubSvc.Get([]byte(strconv.Itoa(schedule.BackupID)))
	if err != nil {
		return nil, err
	}

	if subcribers == nil || len(subcribers.AgentIDs) == 0 {
		schedule.manager.log.WithFields("backup", backup.ID).Debug("No subscribers, skipping backup")
		return nil, nil
	}

	agents := make([]*entities.Agent, 0)
	for _, agentID := range subcribers.AgentIDs {
		agent, err := schedule.manager.services.AgentSvc.Get([]byte(strconv.Itoa(agentID)))
		if err != nil {
			schedule.manager.log.WithFields("backup", backup.ID).Error("schedule.Start: could not get agent", err)
			continue
		}

		if agent == nil {
			schedule.manager.log.WithFields("backup", backup.ID).Error("schedule.Start: no agent found with ID defined as backup subscriber")
			continue
		}

		agents = append(agents, agent)
	}

	jobs := []*entities.Job{}
	for _, agent := range agents {
		target := strconv.Itoa(backup.Target)
		repo, err := schedule.manager.services.RepoSvc.Get([]byte(target))
		if err != nil {
			schedule.manager.log.WithFields("backup", backup.ID).Error("schedule.Start: could not get repos", err)
			continue
		}

		if repo == nil {
			// log.Println("No repo found with ID defined in backup target")
			schedule.manager.log.WithFields("backup", backup.ID).Error("schedule.Start: no repo found with ID defined in backup target")
			break
		}

		jobPacket := &entities.JobPacket{
			Type:  "backup",
			Agent: agent,
			Repo:  repo,
		}

		backupJob := &entities.BackupJob{
			Backup: backup,
		}

		job, err := schedule.manager.NewJob(jobPacket, backupJob)
		if err != nil {
			schedule.manager.log.WithFields("backup", backup.ID).Error("schedule.Start: could not create new job", err)
			return nil, err
		}
		schedule.manager.log.WithFields("backup", backup.ID).Debug("Enqueuing job", job.ID, "for agent", agent.Name)
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func GetSchedule(backupID int) *schedule {
	for _, sch := range schedules {
		if sch.BackupID == backupID {
			return sch
		}
	}
	return nil
}

func GetSchedules(cronSchedule string) []*schedule {
	matches := []*schedule{}
	for _, sch := range schedules {
		if sch.Schedule == cronSchedule {
			matches = append(matches, sch)
		}
	}
	return matches
}

func (man *Manager) AddSchedule(cronSchedule string, backupID int) *schedule {
	schedule := schedule{
		BackupID: backupID,
		manager:  man,
	}

	schedule.NewScheduler(cronSchedule)
	schedules = append(schedules, &schedule)

	return &schedule
}

func (sch *schedule) NewScheduler(cronSchedule string) {
	if sch.Scheduler != nil {
		sch.Scheduler.Stop()
	}
	sch.Schedule = cronSchedule

	scheduler := cron.New()
	sch.Scheduler = scheduler

	scheduler.AddFunc(sch.Schedule, func() {
		sch.Start()
	})

	scheduler.Start()
}

func StopSchedulers() {
	for _, schedule := range schedules {
		schedule.Scheduler.Stop()
	}
}

func RemoveSchedule(backupID int) {
	for i, schedule := range schedules {
		if schedule.BackupID == backupID {
			schedule.Scheduler.Stop()
			schedules = append(schedules[:i], schedules[i+1:]...)
		}
	}
}
