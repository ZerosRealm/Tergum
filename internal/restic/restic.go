package restic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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

type BackupMsg struct {
	Type string `json:"message_type"`
}

type BackupStatus struct {
	BackupMsg
	Percent      float64  `json:"percent_done"`
	TotalFiles   int      `json:"total_files"`
	TotalBytes   int      `json:"total_bytes"`
	BytesDone    int      `json:"bytes_done"`
	CurrentFiles []string `json:"current_files"`
}

type BackupSummary struct {
	BackupMsg
	Snapshot string `json:"snapshot_id"`
}

// Backup source to target repo.
func (r *Restic) Backup(repo, source, password string, jobID string, updates chan []byte, env ...string) ([]byte, error) {
	args := []string{
		"backup",
		"--json",
		"--repo",
		repo,
		source,
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

	// TODO: Integrate the summary. Use map of interfaces and then cast based on type?
	// {"message_type":"summary","files_new":0,"files_changed":0,"files_unmodified":1,"dirs_new":0,"dirs_changed":0,"dirs_unmodified":4,"data_blobs":0,"tree_blobs":0,"data_added":0,"total_files_processed":1,"total_bytes_processed":10485760000,"total_duration":0.2987189,"snapshot_id":"3ddedf4e"}
	go func() {
		for {
			data, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}
			// data = strings.Replace(data, "\n", "", -1)
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

			// fmt.Println(string(data))
			// spew.Dump(msg)
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
