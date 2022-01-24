package server

import (
	"log"

	"github.com/robfig/cron/v3"
	"zerosrealm.xyz/tergum/internal/types"
)

type schedule struct {
	Backup    *types.Backup
	Schedule  string
	Scheduler *cron.Cron

	manager *Manager
}

var schedules = []*schedule{}

func buildSchedules(man *Manager) {
	backups, err := man.services.backupSvc.GetAll()
	if err != nil {
		// TODO: Implement proper logging.
		log.Println(err)
		return
	}

	for _, backup := range backups {
		addSchedule(backup.Schedule, man, backup)
	}
}

func (schedule *schedule) start() {
	for _, agent := range savedData.BackupSubscribers[schedule.Backup.ID] {
		job := types.JobPacket{
			Type:  "backup",
			Agent: agent,
		}

		repos, err := schedule.manager.services.repoSvc.GetAll()
		if err != nil {
			// TODO: Implement proper logging.
			log.Println(err)
			continue
		}

		// TODO: Use filtering.
		var foundRepo *types.Repo
		for _, repo := range repos {
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

		id, err := schedule.manager.NewJob(&job, &backupJob)
		if err != nil {
			log.Printf("job %s could not be enqueued\n", id)
			return
		}
		log.Printf("enqueuing job %s for %s\n", id, agent.Name)
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

func addSchedule(cronSchedule string, manager *Manager, backup *types.Backup) *schedule {
	schedule := schedule{
		Backup:  backup,
		manager: manager,
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
		if schedule.Backup.ID == backupID {
			schedule.Scheduler.Stop()
			schedules = append(schedules[:i], schedules[i+1:]...)
		}
	}
}
