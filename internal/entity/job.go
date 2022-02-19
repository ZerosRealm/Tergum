package entity

import (
	"encoding/json"
	"time"
)

type Job struct {
	ID      string `json:"id"`
	Done    bool   `json:"done"`
	Aborted bool   `json:"aborted"`
	// Packet   *JobPacket      `json:"-"`
	Request  interface{}     `json:"-"`
	Progress json.RawMessage `json:"progress"`

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// JobRequest to send to agents.
type JobRequest struct {
	ID string
	// Repo  *Repo
	Agent *Agent

	Type    string
	Request interface{}
}

// // BackupJob to send to an agent.
// type BackupJob struct {
// 	JobPacket
// 	Backup *Backup
// }

// // RestoreJob to send to an agent.
// type RestoreJob struct {
// 	JobPacket
// 	Snapshot string
// 	Target   string
// 	Include  []string
// 	Exclude  []string
// }

// // StopJob to send to an agent.
// type StopJob struct {
// 	JobPacket
// 	ID string
// }
