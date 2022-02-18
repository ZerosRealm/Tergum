package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"zerosrealm.xyz/tergum/internal/entity"
)

func (api *API) GetRepos() http.HandlerFunc {
	type response struct {
		Repos []*entity.Repo `json:"repos"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		repos, err := api.services.RepoSvc.GetAll()
		if err != nil {
			api.error(w, r, "Could not get repositories.", err, http.StatusInternalServerError)
			return
		}

		if repos == nil {
			repos = make([]*entity.Repo, 0)
		}

		api.respond(w, r, &response{Repos: repos}, http.StatusOK)
	}
}

func (api *API) CreateRepo() http.HandlerFunc {
	type request struct {
		Name     string   `json:"name"`
		Repo     string   `json:"repo"`
		Password string   `json:"password"`
		Settings []string `json:"settings"`
	}
	type response struct {
		Repo *entity.Repo `json:"repo"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		repo := &entity.Repo{
			Name:     req.Name,
			Repo:     req.Repo,
			Password: req.Password,
			Settings: req.Settings,
		}

		repo, err = api.services.RepoSvc.Create(repo)
		if err != nil {
			api.error(w, r, "Could not create repository.", err, http.StatusInternalServerError)
			return
		}

		r.Header.Add("Location", fmt.Sprintf("/repo/%d", repo.ID))
		api.respond(w, r, response{Repo: repo}, http.StatusCreated)
	}
}

func (api *API) UpdateRepo() http.HandlerFunc {
	type request struct {
		Name     string   `json:"name"`
		Repo     string   `json:"repo"`
		Password string   `json:"password"`
		Settings []string `json:"settings"`
	}
	type response struct {
		Repo *entity.Repo `json:"repo"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		repoID := vars["id"]

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
			api.error(w, r, "No repository with that ID.", fmt.Errorf("no repo with that ID"), http.StatusNotFound)
			return
		}

		status := http.StatusOK
		// TODO: Create a repo with the given ID if it does not exist
		// if foundRepo == nil {
		// 	foundRepo = &entity.Repo{
		// 		ID: id,
		// 	}
		// 	status = http.StatusCreated
		// 	savedData.RepoIncrement++
		// 	savedData.Repos = append(savedData.Repos, foundRepo)
		// }

		repo.Name = req.Name
		repo.Repo = req.Repo
		repo.Password = req.Password
		repo.Settings = req.Settings

		repo, err = api.services.RepoSvc.Update(repo)
		if err != nil {
			api.error(w, r, "Could not update repository.", err, http.StatusInternalServerError)
			return
		}

		api.respond(w, r, response{Repo: repo}, status)
	}
}

func (api *API) DeleteRepo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		repoID := vars["id"]

		repo, err := api.services.RepoSvc.Get([]byte(repoID))
		if err != nil {
			api.error(w, r, "Could not get repository.", err, http.StatusInternalServerError)
			return
		}

		if repo == nil {
			api.error(w, r, "No repository found with that ID.", fmt.Errorf("no repo with that ID"), http.StatusNotFound)
			return
		}

		err = api.services.RepoSvc.Delete([]byte(repoID))
		if err != nil {
			api.error(w, r, "Could not delete repository.", err, http.StatusInternalServerError)
			return
		}

		api.respond(w, r, nil, http.StatusNoContent)
	}
}
