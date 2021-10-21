package types

import "time"

// Repo for storing repository information.
type Repo struct {
	ID       int
	Name     string
	Repo     string
	Password string
	Settings []string
}

// Backup for a certain source to the target repository.
type Backup struct {
	ID       int
	Target   int
	Source   string
	Schedule string
	Exclude  []string
	LastRun  time.Time
}

// Agent to send jobs to.
type Agent struct {
	ID   int
	Name string
	IP   string
	Port int
	PSK  string
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
