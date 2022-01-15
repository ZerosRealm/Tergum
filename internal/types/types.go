package types

import (
	"encoding/json"
	"time"
)

// Repo for storing repository information.
type Repo struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Repo     string   `json:"repo"`
	Password string   `json:"password"`
	Settings []string `json:"settings"`
}

// Backup for a certain source to the target repository.
type Backup struct {
	ID       int       `json:"id"`
	Target   int       `json:"target"`
	Source   string    `json:"source"`
	Schedule string    `json:"schedule"`
	Exclude  []string  `json:"exclude"`
	LastRun  time.Time `json:"last_run"`
}

// Agent to send jobs to.
type Agent struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
	PSK  string `json:"psk"`
}

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
	Data interface{}
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
