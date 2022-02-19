package api

import (
	"fmt"
	"net/http"
	"strings"

	"zerosrealm.xyz/tergum/internal/agent/api/request"
	"zerosrealm.xyz/tergum/internal/restic"
)

func (api *API) GetSnapshots() http.HandlerFunc {
	type response struct {
		Snapshots []*restic.Snapshot `json:"snapshots"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request.GetSnapshots
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		snapshots, err := api.manager.GetSnapshots(req.Repo)
		if err != nil {
			api.error(w, r, "Could not get snapshots.", err, http.StatusInternalServerError)
			return
		}

		api.respond(w, r, response{Snapshots: snapshots}, http.StatusOK)
	}
}

func (api *API) DeleteSnapshot() http.HandlerFunc {
	type response struct {
		Output string `json:"output"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request.DeleteSnapshot
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		out, err := api.manager.Forget(req.Repo, req.Snapshots, nil)
		if err != nil {
			api.error(w, r, "Could not delete snapshots.", fmt.Errorf("%s: %s", err, out), http.StatusInternalServerError)
			return
		}

		api.respond(w, r, response{Output: string(out)}, http.StatusOK)
	}
}

func (api *API) Forget() http.HandlerFunc {
	type response struct {
		Output string `json:"output"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request.Forget
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		options := &restic.ForgetOptions{
			LastX:   req.Policy.LastX,
			Hourly:  req.Policy.Hourly,
			Daily:   req.Policy.Daily,
			Weekly:  req.Policy.Weekly,
			Monthly: req.Policy.Monthly,
			Yearly:  req.Policy.Yearly,
		}

		out, err := api.manager.Forget(req.Repo, nil, options)
		if err != nil {
			api.error(w, r, "Could not forget snapshots.", fmt.Errorf("%s: %s", err, out), http.StatusInternalServerError)
			return
		}

		api.respond(w, r, response{Output: string(out)}, http.StatusOK)
	}
}

func (api *API) ListSnapshot() http.HandlerFunc {
	type file struct {
		*restic.FileNode
		Files []*file `json:"files"`
	}

	type response struct {
		Directories []*file `json:"directories"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request.List
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		nodes, err := api.manager.List(req.Repo, req.Snapshot)
		if err != nil {
			api.error(w, r, "Could not list snapshot.", err, http.StatusInternalServerError)
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
