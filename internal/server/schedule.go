package server

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/robfig/cron/v3"
	"github.com/rs/xid"
	"zerosrealm.xyz/tergum/internal/types"
)

type schedule struct {
	Backup    *types.Backup
	Schedule  string
	Scheduler *cron.Cron
}

var schedules = []*schedule{}

func buildSchedules() {
	for _, backup := range savedData.Backups {
		addSchedule(backup.Schedule, backup)
	}
}

func scheduleBackup(schedule *schedule) {
	for _, agent := range savedData.BackupSubscribers[schedule.Backup.ID] {
		id := xid.New().String()

		job := types.JobPacket{}
		job.ID = id
		job.Type = "backup"
		job.Agent = agent

		var foundRepo *types.Repo
		for _, repo := range savedData.Repos {
			if repo.ID == schedule.Backup.Target {
				foundRepo = repo
				break
			}
		}

		if foundRepo.ID == 0 {
			log.Println("No repo found with ID defined in backup target")
			break
		}
		job.Repo = foundRepo

		backupJob := types.BackupJob{
			Backup: schedule.Backup,
		}

		buf := bytes.NewBuffer(nil)
		enc := gob.NewEncoder(buf)
		err := enc.Encode(backupJob)

		if err != nil {
			spew.Dump(backupJob)
			panic(err)
		}
		job.Job = buf.Bytes()

		log.Printf("enqueuing job %s for %s\n", id, agent.Name)
		ok := enqueue(job)
		if !ok {
			log.Printf("job %s could not be enqueued\n", id)
		}
	}
}

func getSchedule(backupID int) *schedule {
	for _, sch := range schedules {
		if sch.Backup.ID == backupID {
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

func addSchedule(cronSchedule string, backup *types.Backup) *schedule {
	schedule := schedule{
		Backup: backup,
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

	scheduler.AddFunc(cronSchedule, func() {
		scheduleBackup(sch)
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
		if schedule.Backup.ID == backupID {
			schedule.Scheduler.Stop()
			schedules = append(schedules[:i], schedules[i+1:]...)
		}
	}
}
