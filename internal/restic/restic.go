package restic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
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

// TODO: Add cancelation?
// Backup source to target repo.
func (r *Restic) Backup(repo, source, password string, exclude []string, jobID string, updates chan []byte, env ...string) ([]byte, error) {
	args := []string{
		"backup",
		"--json",
		"--repo",
		repo,
		source,
	}

	if exclude != nil && len(exclude) != 0 {
		for _, val := range exclude {
			args = append(args, "--exclude")
			args = append(args, val)
		}
	}

	// defer cancel()

	cmd := exec.Command(r.exe, args...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RESTIC_PASSWORD="+password)
	for _, env := range env {
		cmd.Env = append(cmd.Env, env)
	}

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(cmdReader)

	go func() {
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
			msg["msg"] = vData

			msgJSON, err := json.Marshal(&msg)
			if err != nil {
				panic(err)
			}

			select {
			case updates <- msgJSON:
				log.Println("sent update")
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

	return []byte("Done!!!!!"), nil
}

// Backup source to target repo.
func (r *Restic) oldBackup(repo, source, password string, env ...string) ([]byte, error) {
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
	for _, env := range env {
		cmd.Env = append(cmd.Env, env)
	}

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

	if policies != nil && len(policies) > 0 {
		for _, policy := range policies {
			split := strings.SplitN(policy, " ", 2)
			for _, str := range split {
				args = append(args, str)
			}
		}
	}

	cmd := exec.Command(r.exe, args...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RESTIC_PASSWORD="+password)
	for _, env := range env {
		cmd.Env = append(cmd.Env, env)
	}

	return cmd.CombinedOutput()
}
