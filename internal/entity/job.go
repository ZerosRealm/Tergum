package entity

import (
	"encoding/json"
	"time"
)

type Job struct {
	ID       string          `json:"id"`
	Done     bool            `json:"done"`
	Aborted  bool            `json:"aborted"`
	Packet   *JobPacket      `json:"-"`
	Progress json.RawMessage `json:"progress"`

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// JobPacket to send to agents.
type JobPacket struct {
	ID    string
	Repo  *Repo
	Agent *Agent

	Type string
	Data []byte
}

// BackupJob to send to an agent.
type BackupJob struct {
	Backup *Backup
}

// RestoreJob to send to an agent.
type RestoreJob struct {
	Snapshot string
	Target   string
	Include  string
	Exclude  string
}

// StopJob to send to an agent.
type StopJob struct {
	ID string
}
