package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"zerosrealm.xyz/tergum/internal/entities"
	"zerosrealm.xyz/tergum/internal/restic"
	manager "zerosrealm.xyz/tergum/internal/server/manager"
)

func (api *API) GetSnapshots(resticExe *restic.Restic) http.HandlerFunc {
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

		snapshots, err := resticExe.Snapshots(repo.Repo, repo.Password, repo.Settings...)
		if err != nil {
			api.error(w, r, "Could not get snapshots.", err, http.StatusInternalServerError)
			return
		}
		for i, snapshot := range snapshots {
			if len(snapshot.Tags) == 0 {
				snapshots[i].Tags = make([]string, 0)
			}
		}

		api.respond(w, r, response{Snapshots: snapshots}, http.StatusOK)
	}
}

func (api *API) DeleteSnapshot(resticExe *restic.Restic) http.HandlerFunc {
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

		out, err := resticExe.Forget(repo.Repo, repo.Password, snapshot, nil, repo.Settings...)
		if err != nil {
			api.error(w, r, "Running forget policy returned errors.", fmt.Errorf("%s: %s", string(out), err), http.StatusInternalServerError)
			return
		}

		api.respond(w, r, nil, http.StatusNoContent)
	}
}

func (api *API) RestoreSnapshot(manager *manager.Manager) http.HandlerFunc {
	type request struct {
		Agent   int    `json:"agent"`
		Dest    string `json:"destination"`
		Include string `json:"include"`
		Exclude string `json:"exclude"`
	}
	type response struct {
		Job *entities.Job `json:"job"`
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

		jobPacket := &entities.JobPacket{
			Type:  "restore",
			Repo:  repo,
			Agent: agent,
		}

		restoreJob := &entities.RestoreJob{
			Snapshot: snapshot,
			Target:   req.Dest,
			Include:  req.Include,
			Exclude:  req.Exclude,
		}

		job, err := manager.NewJob(jobPacket, restoreJob)
		if err != nil {
			api.error(w, r, "Could not create job.", err, http.StatusInternalServerError)
			return
		}
		api.log.Debug("Enqueuing job %s for %s", job.ID, agent.Name)

		api.respond(w, r, response{Job: job}, http.StatusOK)
	}
}

func (api *API) ListSnapshot(resticExe *restic.Restic) http.HandlerFunc {
	type file struct {
		*restic.FileNode
		Files []*file `json:"files"`
	}

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

		nodes, err := resticExe.List(repo.Repo, repo.Password, snapshot, repo.Settings...)
		if err != nil {
			api.error(w, r, "Could not list files inside snapshot.", err, http.StatusInternalServerError)
			return
		}

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

		api.respond(w, r, response{Directories: rootDirs}, http.StatusOK)
	}
}
