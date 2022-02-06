package server

import (
	"fmt"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/robfig/cron/v3"
	"zerosrealm.xyz/tergum/internal/types"
)

type schedule struct {
	BackupID  int
	Schedule  string
	Scheduler *cron.Cron

	manager *Manager
}

var schedules = []*schedule{}

func (man *Manager) buildSchedules() {
	man.log.Debug("Building schedules")
	backups, err := man.services.backupSvc.GetAll()
	if err != nil {
		man.log.Error("buildSchedules: could not get backups", err)
		return
	}

	for _, backup := range backups {
		man.log.Debug("Adding schedule for backup", fmt.Sprintf("#%d", backup.ID))
		man.addSchedule(backup.Schedule, backup.ID)
	}
}

func (schedule *schedule) start() ([]string, error) {
	backup, err := schedule.manager.services.backupSvc.Get([]byte(strconv.Itoa(schedule.BackupID)))
	if err != nil {
		return nil, err
	}

	schedule.manager.log.WithFields("backup", backup.ID).Debug("Starting backup")

	// subscribers := schedule.manager.services.backupSvc.

	spew.Dump(backup)

	if len(savedData.BackupSubscribers[backup.ID]) == 0 {
		schedule.manager.log.WithFields("backup", backup.ID).Debug("No subscribers, skipping backup")
		return nil, nil
	}

	jobs := []string{}
	for _, agent := range savedData.BackupSubscribers[backup.ID] {
		target := strconv.Itoa(backup.Target)
		repo, err := schedule.manager.services.repoSvc.Get([]byte(target))
		if err != nil {
			schedule.manager.log.WithFields("backup", backup.ID).Error("schedule.Start: could not get repos", err)
			continue
		}

		if repo == nil {
			// log.Println("No repo found with ID defined in backup target")
			schedule.manager.log.WithFields("backup", backup.ID).Error("schedule.Start: no repo found with ID defined in backup target")
			break
		}
		job := types.JobPacket{
			Type:  "backup",
			Agent: agent,
			Repo:  repo,
		}

		backupJob := types.BackupJob{
			Backup: backup,
		}

		id, err := schedule.manager.NewJob(&job, &backupJob)
		if err != nil {
			schedule.manager.log.WithFields("backup", backup.ID).Error("schedule.Start: job could not be enqueued", err)
			return nil, err
		}
		schedule.manager.log.WithFields("backup", backup.ID).Debug("Enqueuing job", id, "for agent", agent.Name)
		jobs = append(jobs, id)
	}

	return jobs, nil
}

func getSchedule(backupID int) *schedule {
	for _, sch := range schedules {
		if sch.BackupID == backupID {
			return sch
		}
	}
	return nil
}

func getSchedules(cronSchedule string) []*schedule {
	matches := []*schedule{}
	for _, sch := range schedules {
		if sch.Schedule == cronSchedule {
			matches = append(matches, sch)
		}
	}
	return matches
}

func (man *Manager) addSchedule(cronSchedule string, backupID int) *schedule {
	schedule := schedule{
		BackupID: backupID,
		manager:  man,
	}

	schedule.newScheduler(cronSchedule)
	schedules = append(schedules, &schedule)

	return &schedule
}

func (sch *schedule) newScheduler(cronSchedule string) {
	if sch.Scheduler != nil {
		sch.Scheduler.Stop()
	}
	sch.Schedule = cronSchedule

	scheduler := cron.New()
	sch.Scheduler = scheduler

	scheduler.AddFunc(sch.Schedule, func() {
		sch.start()
	})

	scheduler.Start()
}

func stopSchedulers() {
	for _, schedule := range schedules {
		schedule.Scheduler.Stop()
	}
}

func removeSchedule(backupID int) {
	for i, schedule := range schedules {
		if schedule.BackupID == backupID {
			schedule.Scheduler.Stop()
			schedules = append(schedules[:i], schedules[i+1:]...)
		}
	}
}
