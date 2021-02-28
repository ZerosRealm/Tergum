package restic

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Restic executable.
type Restic struct {
	exe string
}

// New restic instance.
func New(exePath string) *Restic {
	return &Restic{
		exe: exePath,
	}
}

// Backup source to target repo.
func (r *Restic) Backup(repo, source, password string, env ...string) ([]byte, error) {
	args := []string{
		"backup",
		"--repo",
		repo,
		source,
	}

	cmd := exec.Command(r.exe, args...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RESTIC_PASSWORD="+password)
	for _, env := range env {
		cmd.Env = append(cmd.Env, env)
	}

	return cmd.CombinedOutput()
}

// Snapshot instance from repo.
type Snapshot struct {
	ID    string
	Time  string
	Host  string
	Tags  string
	Paths string
}

// Snapshots from repo.
func (r *Restic) Snapshots(repo, password string, env ...string) ([]Snapshot, error) {
	args := []string{
		"snapshots",
		"--repo",
		repo,
	}

	cmd := exec.Command(r.exe, args...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RESTIC_PASSWORD="+password)
	for _, env := range env {
		cmd.Env = append(cmd.Env, env)
	}

	out, err := cmd.CombinedOutput()

	if err != nil {
		if len(out) == 0 {
			return nil, err
		}

		return nil, fmt.Errorf("%s", string(out))
	}

	re := regexp.MustCompile(`(?m)(\w+)  (\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})  (\w+)     (.+)  (.+)`)
	matches := re.FindAllStringSubmatch(string(out), -1)

	snapshots := make([]Snapshot, 0)
	for _, match := range matches {
		snapshots = append(snapshots, Snapshot{
			ID:    match[1],
			Time:  match[2],
			Host:  match[3],
			Tags:  strings.ReplaceAll(match[4], " ", ""),
			Paths: match[5],
		})
	}

	return snapshots, err
}
