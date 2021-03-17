package server

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/rs/xid"
	"zerosrealm.xyz/tergum/internal/types"
)

func encodeError(msg string) []byte {
	var data = map[string]string{}
	data["message"] = msg
	data["type"] = "error"
	jsonMsg, _ := json.Marshal(data)
	return jsonMsg
}

func encodeMessage(msg string, msgType string) []byte {
	var data = map[string]string{}
	data["message"] = msg
	data["type"] = msgType
	jsonMsg, _ := json.Marshal(data)
	return jsonMsg
}

func getBackups() ([]byte, error) {
	backups := make(map[string]interface{})
	backups["type"] = "getbackups"
	backups["backups"] = []types.Backup{}

	for _, backup := range savedData.Backups {
		backups["backups"] = append(backups["backups"].([]types.Backup), *backup)
	}

	var buf = bytes.NewBufferString("")
	enc := json.NewEncoder(buf)
	err := enc.Encode(backups)
	if err != nil {
		return encodeError(err.Error()), err
	}

	return buf.Bytes(), nil
}

func getRepos() ([]byte, error) {
	repos := make(map[string]interface{})
	repos["type"] = "getrepos"
	repos["repos"] = []types.Repo{}

	for _, repo := range savedData.Repos {
		repos["repos"] = append(repos["repos"].([]types.Repo), *repo)
	}

	var buf = bytes.NewBufferString("")
	enc := json.NewEncoder(buf)
	err := enc.Encode(repos)
	if err != nil {
		return encodeError(err.Error()), err
	}

	return buf.Bytes(), nil
}

func getAgents() ([]byte, error) {
	agents := make(map[string]interface{})
	agents["type"] = "getagents"
	agents["agents"] = []types.Agent{}

	for _, agent := range savedData.Agents {
		agents["agents"] = append(agents["agents"].([]types.Agent), *agent)
	}

	var buf = bytes.NewBufferString("")
	enc := json.NewEncoder(buf)
	err := enc.Encode(agents)
	if err != nil {
		return encodeError(err.Error()), err
	}

	return buf.Bytes(), nil
}

func newBackup(data map[string]interface{}) ([]byte, error) {
	var target int
	switch v := data["target"].(type) {
	case int:
		target = v
	case float32:
		target = int(v)
	case float64:
		target = int(v)
	default:
		msg := "target was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var source string
	switch v := data["source"].(type) {
	case string:
		source = v
	default:
		msg := "source was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var schedule string
	switch v := data["schedule"].(type) {
	case string:
		schedule = v
	default:
		msg := "schedule was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	// TODO: Validate schedule - (@(annually|yearly|monthly|weekly|daily|hourly|reboot))|(@every (\d+(ns|us|µs|ms|s|m|h))+)|((((\d+,)+\d+|(\d+(\/|-)\d+)|\d+|\*) ?){5,7})

	savedData.BackupIncrement++
	backup := types.Backup{
		ID:       savedData.BackupIncrement,
		Target:   target,
		Source:   source,
		Schedule: schedule,
	}

	savedData.Backups = append(savedData.Backups, &backup)

	addSchedule(schedule, &backup)

	return encodeMessage("New backup added!", "success"), nil
}

func newAgent(data map[string]interface{}) ([]byte, error) {
	var name string
	switch v := data["name"].(type) {
	case string:
		name = v
	default:
		msg := "name was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var psk string
	switch v := data["psk"].(type) {
	case string:
		psk = v
	default:
		msg := "psk was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var ip string
	switch v := data["ip"].(type) {
	case string:
		ip = v
	default:
		msg := "ip was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var port int
	switch v := data["port"].(type) {
	case int:
		port = v
	case float32:
		port = int(v)
	case float64:
		port = int(v)
	default:
		msg := "port was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	savedData.AgentIncrement++
	agent := types.Agent{
		ID:   savedData.AgentIncrement,
		Name: name,
		PSK:  psk,
		IP:   ip,
		Port: port,
	}

	savedData.Agents = append(savedData.Agents, &agent)

	return encodeMessage("New agent added!", "success"), nil
}

func newRepo(data map[string]interface{}) ([]byte, error) {
	var name string
	switch v := data["name"].(type) {
	case string:
		name = v
	default:
		msg := "name was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var repo string
	switch v := data["repo"].(type) {
	case string:
		repo = v
	default:
		msg := "repo was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var password string
	switch v := data["password"].(type) {
	case string:
		password = v
	default:
		msg := "password was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var settings []string
	switch v := data["settings"].(type) {
	case []string:
		settings = v
	case []interface{}:
		settings = []string{}
		for _, val := range v {
			switch vVal := val.(type) {
			case string:
				settings = append(settings, vVal)
			default:
				msg := "settings was of invalid type"
				return encodeError(msg), fmt.Errorf(msg)
			}
		}
	default:
		msg := "settings was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	savedData.RepoIncrement++
	newRepo := types.Repo{
		ID:       savedData.RepoIncrement,
		Name:     name,
		Repo:     repo,
		Password: password,
		Settings: settings,
	}

	savedData.Repos = append(savedData.Repos, &newRepo)

	return encodeMessage("New repo added!", "success"), nil
}

func updateRepo(data map[string]interface{}) ([]byte, error) {
	var id int
	switch v := data["id"].(type) {
	case int:
		id = v
	case float32:
		id = int(v)
	case float64:
		id = int(v)
	default:
		msg := "id was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var index int
	var foundRepo *types.Repo
	for i, repo := range savedData.Repos {
		if repo.ID == id {
			index = i
			foundRepo = repo
			break
		}
	}

	if foundRepo == nil {
		msg := "no repo was found with that id"
		return encodeError(msg), fmt.Errorf(msg)
	}

	if _, ok := data["name"]; ok {
		switch v := data["name"].(type) {
		case string:
			foundRepo.Name = v
		default:
			msg := "name was of invalid type"
			return encodeError(msg), fmt.Errorf(msg)
		}
	}

	if _, ok := data["repo"]; ok {
		switch v := data["repo"].(type) {
		case string:
			foundRepo.Repo = v
		default:
			msg := "repo was of invalid type"
			return encodeError(msg), fmt.Errorf(msg)
		}
	}

	if _, ok := data["password"]; ok {
		switch v := data["password"].(type) {
		case string:
			foundRepo.Password = v
		default:
			msg := "password was of invalid type"
			return encodeError(msg), fmt.Errorf(msg)
		}
	}

	if _, ok := data["settings"]; ok {
		switch v := data["settings"].(type) {
		case []string:
			foundRepo.Settings = v
		case []interface{}:
			foundRepo.Settings = []string{}
			for _, val := range v {
				switch vVal := val.(type) {
				case string:
					foundRepo.Settings = append(foundRepo.Settings, vVal)
				default:
					msg := "settings was of invalid type"
					return encodeError(msg), fmt.Errorf(msg)
				}
			}
		default:
			msg := "settings was of invalid type"
			return encodeError(msg), fmt.Errorf(msg)
		}
	}

	savedData.Repos[index] = foundRepo

	return encodeMessage("Repo updated!", "success"), nil
}

func updateBackup(data map[string]interface{}) ([]byte, error) {
	var id int
	switch v := data["id"].(type) {
	case int:
		id = v
	case float32:
		id = int(v)
	case float64:
		id = int(v)
	default:
		msg := "id was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var index int
	var foundBackup *types.Backup
	for i, backup := range savedData.Backups {
		if backup.ID == id {
			index = i
			foundBackup = backup
			break
		}
	}

	if foundBackup == nil {
		msg := "no backup was found with that id"
		return encodeError(msg), fmt.Errorf(msg)
	}

	if _, ok := data["target"]; ok {
		switch v := data["target"].(type) {
		case int:
			foundBackup.Target = v
		case float32:
			foundBackup.Target = int(v)
		case float64:
			foundBackup.Target = int(v)
		default:
			msg := "target was of invalid type"
			return encodeError(msg), fmt.Errorf(msg)
		}
	}

	if _, ok := data["source"]; ok {
		switch v := data["source"].(type) {
		case string:
			foundBackup.Source = v
		default:
			msg := "source was of invalid type"
			return encodeError(msg), fmt.Errorf(msg)
		}
	}

	if _, ok := data["schedule"]; ok {
		switch v := data["schedule"].(type) {
		case string:
			// TODO: Validate cron.
			foundBackup.Schedule = v
			schedule := getSchedule(foundBackup.ID)

			if schedule == nil {
				addSchedule(v, foundBackup)
				break
			}
			schedule.newScheduler(v)
		default:
			msg := "schedule was of invalid type"
			return encodeError(msg), fmt.Errorf(msg)
		}
	}

	savedData.Backups[index] = foundBackup

	return encodeMessage("Backup updated!", "success"), nil
}

func updateAgent(data map[string]interface{}) ([]byte, error) {
	var id int
	switch v := data["id"].(type) {
	case int:
		id = v
	case float32:
		id = int(v)
	case float64:
		id = int(v)
	default:
		msg := "id was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var index int
	var foundAgent *types.Agent
	for i, agent := range savedData.Agents {
		if agent.ID == id {
			index = i
			foundAgent = agent
			break
		}
	}

	if foundAgent == nil {
		msg := "no agent was found with that id"
		return encodeError(msg), fmt.Errorf(msg)
	}

	if _, ok := data["name"]; ok {
		switch v := data["name"].(type) {
		case string:
			foundAgent.Name = v
		default:
			msg := "name was of invalid type"
			return encodeError(msg), fmt.Errorf(msg)
		}
	}

	if _, ok := data["psk"]; ok {
		switch v := data["psk"].(type) {
		case string:
			foundAgent.PSK = v
		default:
			msg := "psk was of invalid type"
			return encodeError(msg), fmt.Errorf(msg)
		}
	}

	if _, ok := data["ip"]; ok {
		switch v := data["ip"].(type) {
		case string:
			foundAgent.IP = v
		default:
			msg := "ip was of invalid type"
			return encodeError(msg), fmt.Errorf(msg)
		}
	}

	if _, ok := data["port"]; ok {
		switch v := data["port"].(type) {
		case int:
			foundAgent.Port = v
		case float32:
			foundAgent.Port = int(v)
		case float64:
			foundAgent.Port = int(v)
		default:
			msg := "port was of invalid type"
			return encodeError(msg), fmt.Errorf(msg)
		}
	}

	savedData.Agents[index] = foundAgent

	return encodeMessage("Agent updated!", "success"), nil
}

func deleteBackup(data map[string]interface{}) ([]byte, error) {
	var id int
	switch v := data["id"].(type) {
	case int:
		id = v
	case float32:
		id = int(v)
	case float64:
		id = int(v)
	default:
		msg := "id was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var index int
	var found bool
	for i, backup := range savedData.Backups {
		if backup.ID == id {
			index = i
			found = true
			break
		}
	}

	if !found {
		msg := "no backup was found with that id"
		return encodeError(msg), fmt.Errorf(msg)
	}

	savedData.Backups = append(savedData.Backups[:index], savedData.Backups[index+1:]...)

	removeSchedule(id)

	return encodeMessage("Backup deleted!", "success"), nil
}

func deleteRepo(data map[string]interface{}) ([]byte, error) {
	var id int
	switch v := data["id"].(type) {
	case int:
		id = v
	case float32:
		id = int(v)
	case float64:
		id = int(v)
	default:
		msg := "id was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var index int
	var found bool
	for i, repo := range savedData.Repos {
		if repo.ID == id {
			index = i
			found = true
			break
		}
	}

	if !found {
		msg := "no repo was found with that id"
		return encodeError(msg), fmt.Errorf(msg)
	}

	savedData.Repos = append(savedData.Repos[:index], savedData.Repos[index+1:]...)

	return encodeMessage("Repo deleted!", "success"), nil
}

func deleteAgent(data map[string]interface{}) ([]byte, error) {
	var id int
	switch v := data["id"].(type) {
	case int:
		id = v
	case float32:
		id = int(v)
	case float64:
		id = int(v)
	default:
		msg := "id was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var index int
	var found bool
	for i, agent := range savedData.Agents {
		if agent.ID == id {
			index = i
			found = true
			break
		}
	}

	if !found {
		msg := "no agent was found with that id"
		return encodeError(msg), fmt.Errorf(msg)
	}

	savedData.Agents = append(savedData.Agents[:index], savedData.Agents[index+1:]...)

	return encodeMessage("Agent deleted!", "success"), nil
}

func getSubscribers() ([]byte, error) {
	data := make(map[string]interface{})
	data["type"] = "getsubscribers"
	data["subscribers"] = make(map[int][]types.Agent)

	for key, agents := range savedData.BackupSubscribers {
		data["subscribers"].(map[int][]types.Agent)[key] = make([]types.Agent, 0)
		for _, agent := range agents {
			data["subscribers"].(map[int][]types.Agent)[key] = append(data["subscribers"].(map[int][]types.Agent)[key], *agent)
		}
	}

	var buf = bytes.NewBufferString("")
	enc := json.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return encodeError(err.Error()), err
	}

	return buf.Bytes(), nil
}

func updateSubscribers(data map[string]interface{}) ([]byte, error) {
	var backupID int
	switch v := data["backup"].(type) {
	case int:
		backupID = v
	case float32:
		backupID = int(v)
	case float64:
		backupID = int(v)
	default:
		msg := "backup id was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var agentIDs []int
	switch v := data["agents"].(type) {
	case []int:
		agentIDs = v
	case []interface{}:
		agentIDs = []int{}
		for _, val := range v {
			switch vVal := val.(type) {
			case int:
				agentIDs = append(agentIDs, vVal)
			case float32:
				agentIDs = append(agentIDs, int(vVal))
			case float64:
				agentIDs = append(agentIDs, int(vVal))
			default:
				msg := "agents was of invalid type"
				return encodeError(msg), fmt.Errorf(msg)
			}
		}
	default:
		msg := "agents was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var found bool
	for _, backup := range savedData.Backups {
		if backup.ID == backupID {
			found = true
			break
		}
	}

	if !found {
		msg := "no backup was found with that id"
		return encodeError(msg), fmt.Errorf(msg)
	}

	if len(agentIDs) == 0 {
		savedData.BackupSubscribers[backupID] = make([]*types.Agent, 0)
		return encodeMessage("Subscribed to backup!", "updateSubscribers"), nil
	}

	var foundAgents []*types.Agent = make([]*types.Agent, 0)
	for _, agentID := range agentIDs {
		for _, agent := range savedData.Agents {
			if agent.ID == agentID {
				foundAgents = append(foundAgents, agent)
				break
			}
		}
	}

	if savedData.BackupSubscribers[backupID] == nil {
		savedData.BackupSubscribers[backupID] = make([]*types.Agent, 0)
	}

	savedData.BackupSubscribers[backupID] = foundAgents

	return encodeMessage("Subscribed to backup!", "updateSubscribers"), nil
}

func getJobs() ([]byte, error) {
	jobs := make(map[string]interface{})
	jobs["type"] = "getjobs"
	jobs["jobs"] = make(map[string][]byte)

	for job, data := range savedData.Jobs {
		jobs["jobs"].(map[string][]byte)[job] = data
	}

	var buf = bytes.NewBufferString("")
	enc := json.NewEncoder(buf)
	err := enc.Encode(jobs)
	if err != nil {
		return encodeError(err.Error()), err
	}

	return buf.Bytes(), nil
}

func getSnapshots(data map[string]interface{}) ([]byte, error) {
	var repoID int
	switch v := data["id"].(type) {
	case int:
		repoID = v
	case float32:
		repoID = int(v)
	case float64:
		repoID = int(v)
	default:
		msg := "repo id was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var foundRepo *types.Repo
	for _, repo := range savedData.Repos {
		if repo.ID == repoID {
			foundRepo = repo
			break
		}
	}

	if foundRepo == nil {
		msg := "no repo was found with that id"
		return encodeError(msg), fmt.Errorf(msg)
	}

	snapshots, err := resticExe.Snapshots(foundRepo.Repo, foundRepo.Password, foundRepo.Settings...)
	if err != nil {
		return encodeError(err.Error()), err
	}

	resp := make(map[string]interface{})
	resp["type"] = "getsnapshots"
	resp["repo"] = repoID
	resp["snapshots"] = snapshots

	var buf = bytes.NewBufferString("")
	enc := json.NewEncoder(buf)
	err = enc.Encode(resp)
	if err != nil {
		return encodeError(err.Error()), err
	}

	return buf.Bytes(), nil
}

func restoreSnapshot(data map[string]interface{}) ([]byte, error) {
	var repoID int
	switch v := data["repo"].(type) {
	case int:
		repoID = v
	case float32:
		repoID = int(v)
	case float64:
		repoID = int(v)
	default:
		msg := "repo id was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var foundRepo *types.Repo
	for _, repo := range savedData.Repos {
		if repo.ID == repoID {
			foundRepo = repo
			break
		}
	}

	if foundRepo == nil {
		msg := "no repo was found with that id"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var agentID int
	switch v := data["agent"].(type) {
	case int:
		agentID = v
	case float32:
		agentID = int(v)
	case float64:
		agentID = int(v)
	default:
		msg := "agent id was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var foundAgent *types.Agent
	for _, agent := range savedData.Agents {
		if agent.ID == agentID {
			foundAgent = agent
			break
		}
	}

	if foundAgent == nil {
		msg := "no agent was found with that id"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var snapshot string
	switch v := data["snapshot"].(type) {
	case string:
		snapshot = v
	default:
		msg := "snapshot was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var target string
	switch v := data["target"].(type) {
	case string:
		target = v
	default:
		msg := "target was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var include string
	switch v := data["include"].(type) {
	case string:
		include = v
	default:
		msg := "include was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	var exclude string
	switch v := data["exclude"].(type) {
	case string:
		exclude = v
	default:
		msg := "exclude was of invalid type"
		return encodeError(msg), fmt.Errorf(msg)
	}

	id := xid.New().String()

	job := types.JobPacket{}
	job.ID = id
	job.Type = "restore"
	job.Agent = foundAgent
	job.Repo = foundRepo

	restoreJob := types.RestoreJob{
		Snapshot: snapshot,
		Target:   target,
		Include:  include,
		Exclude:  exclude,
	}

	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(restoreJob)

	if err != nil {
		spew.Dump(restoreJob)
		panic(err)
	}
	job.Job = buf.Bytes()

	log.Printf("enqueuing job %s for %s\n", id, foundAgent.Name)
	ok := enqueue(job)
	if !ok {
		msg := fmt.Sprintf("job %s could not be enqueued\n", id)
		return encodeError(msg), fmt.Errorf(msg)
	}

	return encodeMessage("Restore job sent to agent.", "success"), nil
}
