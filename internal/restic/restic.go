package restic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Restic executable.
type Restic struct {
	exe     string
	ctx     context.Context
	Jobs    chan *Job
	Updates chan JobUpdate
}

type Job struct {
	ID     string
	ctx    context.Context
	Cancel context.CancelFunc
}

type JobUpdate struct {
	ID  string
	Msg json.RawMessage
}

// New restic instance.
func New(ctx context.Context, exePath string) *Restic {
	return &Restic{
		exe:     exePath,
		ctx:     ctx,
		Jobs:    make(chan *Job, 100),
		Updates: make(chan JobUpdate, 100),
	}
}

// Restic JSON struct
// https://github.com/restic/restic/blob/master/internal/ui/backup/json.go#L198

// TODO: Add cancelation?
// Backup source to target repo.
func (r *Restic) Backup(repo, source, password string, exclude []string, jobID string, env ...string) ([]byte, error) {
	args := []string{
		"backup",
		"--json",
		"--repo",
		repo,
		source,
	}

	if len(exclude) != 0 {
		for _, val := range exclude {
			args = append(args, "--exclude")
			args = append(args, val)
		}
	}

	// defer cancel()

	cmd := exec.Command(r.exe, args...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RESTIC_PASSWORD="+password)
	cmd.Env = append(cmd.Env, env...)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(cmdReader)

	// Create a job.
	job := &Job{
		ID: jobID,
	}
	ctx, cancel := context.WithCancel(r.ctx)
	defer cancel()
	job.ctx = ctx
	job.Cancel = cancel

	r.Jobs <- job

	go func() {
		if r.Updates == nil {
			return
		}

		for {
			select {
			case <-job.ctx.Done():
				err := cmd.Process.Signal(os.Interrupt)
				if err != nil {
					cmd.Process.Signal(os.Kill)
				}

				return
			default:
			}

			data, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}
			data = bytes.Replace(data[5:], []byte("\n"), []byte(""), -1)

			update := JobUpdate{
				ID:  jobID,
				Msg: json.RawMessage(data),
			}

			select {
			case r.Updates <- update:
			default:
			}
		}
	}()
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return []byte("Done"), nil
}

// Restore snapshot to target.
func (r *Restic) Restore(repo, password, snapshot, target, include, exclude string, env ...string) ([]byte, error) {
	args := []string{
		"restore",
		snapshot,
		"--target",
		target,
		"--repo",
		repo,
	}

	if include != "" {
		args = append(args, "--include")
		args = append(args, include)
	}

	if exclude != "" {
		args = append(args, "--exclude")
		args = append(args, exclude)
	}

	cmd := exec.Command(r.exe, args...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RESTIC_PASSWORD="+password)
	cmd.Env = append(cmd.Env, env...)

	return cmd.CombinedOutput()
}

// Snapshot instance from repo.
type Snapshot struct {
	ID    string   `json:"id"`
	Time  string   `json:"time"`
	Host  string   `json:"hostname"`
	User  string   `json:"username"`
	Tags  []string `json:"tags"`
	Paths []string `json:"paths"`
}

// Snapshots from repo.
func (r *Restic) Snapshots(repo, password string, env ...string) ([]Snapshot, error) {
	args := []string{
		"snapshots",
		"--json",
		"--repo",
		repo,
	}

	cmd := exec.Command(r.exe, args...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RESTIC_PASSWORD="+password)
	cmd.Env = append(cmd.Env, env...)

	out, err := cmd.CombinedOutput()

	if err != nil {
		if len(out) == 0 {
			return nil, err
		}

		return nil, fmt.Errorf("%s", string(out))
	}

	var snapshots []Snapshot
	err = json.Unmarshal(out, &snapshots)
	if err != nil {
		return nil, err
	}

	return snapshots, err
}

// Forget a snapshot.
func (r *Restic) Forget(repo, password, snapshot string, policies []string, env ...string) ([]byte, error) {
	args := []string{
		"forget",
	}

	if snapshot != "" {
		args = append(args, snapshot)
	}

	args = append(args, "--repo")
	args = append(args, repo)

	if len(policies) > 0 {
		for _, policy := range policies {
			split := strings.SplitN(policy, " ", 2)
			args = append(args, split...)
		}
	}

	cmd := exec.Command(r.exe, args...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RESTIC_PASSWORD="+password)
	cmd.Env = append(cmd.Env, env...)

	return cmd.CombinedOutput()
}
