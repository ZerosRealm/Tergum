package types

// Job to send to an agent.
type Job struct {
	ID     string
	Backup *Backup
	Agent  *Agent
}

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
	Job  Job
	Repo Repo
}
