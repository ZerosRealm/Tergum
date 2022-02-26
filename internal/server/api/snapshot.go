package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	agentRequest "zerosrealm.xyz/tergum/internal/agent/api/request"
	"zerosrealm.xyz/tergum/internal/entity"
	"zerosrealm.xyz/tergum/internal/restic"
	manager "zerosrealm.xyz/tergum/internal/server/manager"
)

func (api *API) GetSnapshots(man *manager.Manager, resticExe *restic.Restic) http.HandlerFunc {
	type response struct {
		Snapshots []*restic.Snapshot `json:"snapshots"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		repoID := vars["id"]

		repo, err := api.services.RepoSvc.Get([]byte(repoID))
		if err != nil {
			api.error(w, r, "Could not get repository.", err, http.StatusInternalServerError)
			return
		}

		if repo == nil {
			api.error(w, r, "No repository found with that ID.", fmt.Errorf("no repo found with that ID"), http.StatusNotFound)
			return
		}

		log := api.log.WithFields("method", r.Method, "path", r.URL.Path, "src", r.RemoteAddr)

		if resticExe != nil {
			snapshots, err := resticExe.Snapshots(repo.Repo, repo.Password, repo.Settings...)
			if err == nil {
				api.respond(w, r, response{Snapshots: snapshots}, http.StatusOK)
				return
			}
			log.Debug("Server could not get snapshots:", err)
		}

		agents, err := api.services.AgentSvc.GetAll()
		if err != nil {
			api.error(w, r, "Could not get agents.", err, http.StatusInternalServerError)
			return
		}

		if agents == nil && len(agents) == 0 {
			api.error(w, r, "No agents found to send request to.", fmt.Errorf("no agents found"), http.StatusNotFound)
			return
		}

		for _, agent := range agents {
			log.Debug("Sending request to agent", agent.Name)
			snapshotsReq := &agentRequest.GetSnapshots{
				Repo: repo,
			}
			jobRequest := &entity.JobRequest{
				Type:  "getsnapshots",
				Agent: agent,

				Data: snapshotsReq,
			}

			body, err := man.SendRequest(jobRequest, agent)
			if err != nil {
				log.Debug("Agent returned error:", err)
				continue
			}
			log.Debug("Returned body:", string(body))

			var resp response
			err = json.Unmarshal(body, &resp)
			if err != nil {
				api.error(w, r, "Could not unmarshal agent response.", err, http.StatusInternalServerError)
				return
			}

			api.respond(w, r, resp, http.StatusOK)
			return
		}

		api.error(w, r, "No agents could get snapshots.", fmt.Errorf("no agents could get snapshots, check debug logs."), http.StatusNotFound)
	}
}

func (api *API) DeleteSnapshot(man *manager.Manager, resticExe *restic.Restic) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		repoID := vars["id"]
		snapshot := vars["snapshot"]

		repo, err := api.services.RepoSvc.Get([]byte(repoID))
		if err != nil {
			api.error(w, r, "Could not get repository.", err, http.StatusInternalServerError)
			return
		}

		if repo == nil {
			api.error(w, r, "No repository with that ID.", fmt.Errorf("no repo with that ID"), http.StatusNotFound)
			return
		}

		log := api.log.WithFields("method", r.Method, "path", r.URL.Path, "src", r.RemoteAddr)

		if resticExe != nil {
			out, err := resticExe.Forget(repo.Repo, repo.Password, []string{snapshot}, nil, repo.Settings...)
			if err == nil {
				api.respond(w, r, nil, http.StatusNoContent)
				return
			}
			log.Debug("Server could not delete snapshot:", out)
		}

		agents, err := api.services.AgentSvc.GetAll()
		if err != nil {
			api.error(w, r, "Could not get agents.", err, http.StatusInternalServerError)
			return
		}

		if agents == nil && len(agents) == 0 {
			api.error(w, r, "No agents found to send request to.", fmt.Errorf("no agents found"), http.StatusNotFound)
			return
		}

		for _, agent := range agents {
			log.Debug("Sending request to agent", agent.Name)
			snapshotsReq := &agentRequest.DeleteSnapshot{
				Repo: repo,
				Snapshots: []string{
					snapshot,
				},
			}
			jobRequest := &entity.JobRequest{
				Type:  "deletesnapshot",
				Agent: agent,

				Data: snapshotsReq,
			}

			_, err := man.SendRequest(jobRequest, agent)
			if err != nil {
				log.Debug("Agent returned error:", err)
				continue
			}

			api.respond(w, r, nil, http.StatusNoContent)
			return
		}

		api.error(w, r, "No agents could delete snapshot.", fmt.Errorf("no agents could delete snapshot, check debug logs."), http.StatusNotFound)
	}
}

func (api *API) RestoreSnapshot(manager *manager.Manager) http.HandlerFunc {
	type request struct {
		Agent   int      `json:"agent"`
		Dest    string   `json:"destination"`
		Include []string `json:"include"`
		Exclude []string `json:"exclude"`
	}
	type response struct {
		Job *entity.Job `json:"job"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		repoID := vars["id"]
		snapshot := vars["snapshot"]

		var req request
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		repo, err := api.services.RepoSvc.Get([]byte(repoID))
		if err != nil {
			api.error(w, r, "Could not get repository.", err, http.StatusInternalServerError)
			return
		}

		if repo == nil {
			api.error(w, r, "No repository found with that ID.", fmt.Errorf("no repo found with that ID"), http.StatusNotFound)
			return
		}

		agent, err := api.services.AgentSvc.Get([]byte(strconv.Itoa(req.Agent)))
		if err != nil {
			api.error(w, r, "Could not get agent.", err, http.StatusInternalServerError)
			return
		}

		restoreReq := &agentRequest.Restore{
			Repo:     repo,
			Snapshot: snapshot,
			Target:   req.Dest,
			Include:  req.Include,
			Exclude:  req.Exclude,
		}
		jobRequest := &entity.JobRequest{
			Type:  "restore",
			Agent: agent,
			// Repo:  repo,

			Data: restoreReq,
		}

		job, err := manager.NewJob(jobRequest)
		if err != nil {
			api.error(w, r, "Could not create job.", err, http.StatusInternalServerError)
			return
		}
		api.log.Debug("Enqueuing job %s for %s", job.ID, agent.Name)

		api.respond(w, r, response{Job: job}, http.StatusOK)
	}
}

type file struct {
	*restic.FileNode
	Files []*file `json:"files"`
}

func generateDirectories(nodes []*restic.FileNode) []*file {
	dirIndex := make([]*file, 0)
	rootDirs := make([]*file, 0)
	for _, node := range nodes {
		if node.Type == "dir" {
			dir := &file{
				FileNode: node,
				Files:    make([]*file, 0),
			}
			dirIndex = append(dirIndex, dir)

			found := false
			for _, parentDir := range dirIndex {
				parentPath := parentDir.FileNode.Path

				temp := strings.Split(dir.Path, "/")
				dirPath := temp[0 : len(temp)-1]

				if parentPath == strings.Join(dirPath, "/") {
					parentDir.Files = append(parentDir.Files, dir)
					found = true
					break
				}
			}

			if !found {
				rootDirs = append(rootDirs, dir)
			}
		}

		if node.Type == "file" {
			for _, dir := range dirIndex {
				dirPath := dir.FileNode.Path

				temp := strings.Split(node.Path, "/")
				filePath := temp[0 : len(temp)-1]

				if dirPath == strings.Join(filePath, "/") {
					dir.Files = append(dir.Files, &file{
						FileNode: node,
						Files:    make([]*file, 0),
					})
				}
			}
		}
	}
	return rootDirs
}

func (api *API) ListSnapshot(man *manager.Manager, resticExe *restic.Restic) http.HandlerFunc {
	type response struct {
		Directories []*file `json:"directories"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		repoID := vars["id"]
		snapshot := vars["snapshot"]

		repo, err := api.services.RepoSvc.Get([]byte(repoID))
		if err != nil {
			api.error(w, r, "Could not get repository.", err, http.StatusInternalServerError)
			return
		}

		if repo == nil {
			api.error(w, r, "No repository with that ID.", fmt.Errorf("no repo with that ID"), http.StatusNotFound)
			return
		}

		log := api.log.WithFields("method", r.Method, "path", r.URL.Path, "src", r.RemoteAddr)

		if resticExe != nil {
			nodes, err := resticExe.List(repo.Repo, repo.Password, snapshot, repo.Settings...)
			if err == nil {
				api.respond(w, r, response{Directories: generateDirectories(nodes)}, http.StatusOK)
				return
			}
			log.Debug("Server could not list snapshot:", err)
		}

		agents, err := api.services.AgentSvc.GetAll()
		if err != nil {
			api.error(w, r, "Could not get agents.", err, http.StatusInternalServerError)
			return
		}

		if agents == nil && len(agents) == 0 {
			api.error(w, r, "No agents found to send request to.", fmt.Errorf("no agents found"), http.StatusNotFound)
			return
		}

		for _, agent := range agents {
			log.Debug("Sending request to agent", agent.Name)
			snapshotsReq := &agentRequest.List{
				Repo:     repo,
				Snapshot: snapshot,
			}
			jobRequest := &entity.JobRequest{
				Type:  "list",
				Agent: agent,

				Data: snapshotsReq,
			}

			body, err := man.SendRequest(jobRequest, agent)
			if err != nil {
				log.Debug("Agent returned error:", err)
				continue
			}

			var resp response
			err = json.Unmarshal(body, &resp)
			if err != nil {
				api.error(w, r, "Could not unmarshal agent response.", err, http.StatusInternalServerError)
				return
			}

			api.respond(w, r, resp, http.StatusOK)
			return
		}

		api.error(w, r, "No agents could list snapshot.", fmt.Errorf("no agents could list snapshot, check debug logs."), http.StatusNotFound)
	}
}
