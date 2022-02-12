package restic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
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

	errReader := new(bytes.Buffer)
	cmd.Stderr = errReader

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
					err = cmd.Process.Signal(os.Kill)
					if err != nil {
						panic(fmt.Errorf("restic.Backup: failed to kill process: %w", err))
					}
				}

				return
			default:
			}

			data, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					log.Println("restic.Backup: failed to read update from reader:", err)
				}
				break
			}

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
		out, readErr := io.ReadAll(errReader)
		if readErr != nil {
			return nil, fmt.Errorf("restic.Backup cmd.Start(): could not read stderr: %w", readErr)
		}
		return out, err
	}
	if err := cmd.Wait(); err != nil {
		out, readErr := io.ReadAll(errReader)
		if readErr != nil {
			return nil, fmt.Errorf("restic.Backup cmd.Wait(): could not read stderr: %w", readErr)
		}
		return out, err
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

type ForgetOptions struct {
	LastX   int
	Hourly  int
	Daily   int
	Weekly  int
	Monthly int
	Yearly  int
}

// Forget a snapshot.
func (r *Restic) Forget(repo, password, snapshot string, options *ForgetOptions, env ...string) ([]byte, error) {
	args := []string{
		"forget",
	}

	if snapshot != "" {
		args = append(args, snapshot)
	}

	args = append(args, "--repo")
	args = append(args, repo)

	if options.LastX > 0 {
		args = append(args, "--keep-last")
		args = append(args, strconv.Itoa(options.LastX))
	}

	if options.Hourly > 0 {
		args = append(args, "--keep-hourly")
		args = append(args, strconv.Itoa(options.Hourly))
	}

	if options.Daily > 0 {
		args = append(args, "--keep-daily")
		args = append(args, strconv.Itoa(options.Daily))
	}

	if options.Weekly > 0 {
		args = append(args, "--keep-weekly")
		args = append(args, strconv.Itoa(options.Weekly))
	}

	if options.Monthly > 0 {
		args = append(args, "--keep-monthly")
		args = append(args, strconv.Itoa(options.Monthly))
	}

	if options.Yearly > 0 {
		args = append(args, "--keep-yearly")
		args = append(args, strconv.Itoa(options.Yearly))
	}

	cmd := exec.Command(r.exe, args...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RESTIC_PASSWORD="+password)
	cmd.Env = append(cmd.Env, env...)

	return cmd.CombinedOutput()
}
