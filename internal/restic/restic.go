package restic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Restic executable.
type Restic struct {
	exe     string
	Updates chan []byte
}

// New restic instance.
func New(exePath string) *Restic {
	return &Restic{
		exe:     exePath,
		Updates: make(chan []byte, 100),
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

	go func() {
		if r.Updates == nil {
			return
		}

		for {
			data, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}
			data = bytes.Replace(data[5:], []byte("\n"), []byte(""), -1)

			var vData map[string]interface{}
			err = json.Unmarshal(data, &vData)
			if err != nil {
				panic(err)
			}

			var msg = make(map[string]interface{})
			msg["type"] = "jobProgress"
			msg["job"] = jobID
			msg["time"] = time.Now().Unix()
			msg["msg"] = vData

			msgJSON, err := json.Marshal(&msg)
			if err != nil {
				panic(err)
			}

			select {
			case r.Updates <- msgJSON:
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
