package entity

import "time"

// Backup for a certain source to the target repository.
type Backup struct {
	ID       int       `json:"id"`
	Target   int       `json:"target"`
	Source   string    `json:"source"`
	Schedule string    `json:"schedule"`
	Exclude  []string  `json:"exclude"`
	LastRun  time.Time `json:"last_run"`
}

type BackupSubscribers struct {
	BackupID int   `json:"backup_id"`
	AgentIDs []int `json:"agent_ids"`
}
