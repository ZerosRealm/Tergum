package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"zerosrealm.xyz/tergum/internal/entity"
)

func (api *API) GetSettings() http.HandlerFunc {
	type response struct {
		Settings []*entity.Setting `json:"settings"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		settings, err := api.services.SettingSvc.GetAll()
		if err != nil {
			api.error(w, r, "Could not get settings.", err, http.StatusInternalServerError)
			return
		}

		if settings == nil {
			settings = make([]*entity.Setting, 0)
		}

		api.respond(w, r, &response{Settings: settings}, http.StatusOK)
	}
}

func (api *API) GetSetting() http.HandlerFunc {
	type response struct {
		*entity.Setting
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		settingKey := vars["id"]

		setting, err := api.services.SettingSvc.Get([]byte(settingKey))
		if err != nil {
			api.error(w, r, "Could not get setting.", err, http.StatusInternalServerError)
			return
		}

		if setting == nil {
			api.error(w, r, "No setting found with that ID.", fmt.Errorf("no setting with that ID"), http.StatusNotFound)
			return
		}

		api.respond(w, r, &response{setting}, http.StatusOK)
	}
}

func (api *API) CreateSetting() http.HandlerFunc {
	type request struct {
		*entity.Setting
	}
	type response struct {
		*entity.Setting
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		setting := &entity.Setting{
			Key:   req.Key,
			Value: req.Value,
		}

		setting, err = api.services.SettingSvc.Create(setting)
		if err != nil {
			api.error(w, r, "Could not create setting.", err, http.StatusInternalServerError)
			return
		}

		r.Header.Add("Location", fmt.Sprintf("/setting/%d", setting.Key))
		api.respond(w, r, response{setting}, http.StatusCreated)
	}
}

func (api *API) UpdateSetting() http.HandlerFunc {
	type request struct {
		*entity.Setting
	}
	type response struct {
		*entity.Setting
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		settingKey := vars["id"]

		var req request
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		setting, err := api.services.SettingSvc.Get([]byte(settingKey))
		if err != nil {
			api.error(w, r, "Could not get setting.", err, http.StatusInternalServerError)
			return
		}

		if setting == nil {
			api.error(w, r, "No setting found with that ID.", fmt.Errorf("no setting with that ID"), http.StatusNotFound)
			return
		}

		status := http.StatusOK
		// TODO: Create an setting with the given ID if it does not exist
		// if foundSetting == nil {
		// 	foundSetting = &entity.Setting{
		// 		ID: id,
		// 	}
		// 	status = http.StatusCreated
		// 	savedData.SettingIncrement++
		// 	savedData.Settings = append(savedData.Settings, foundSetting)
		// }

		setting.Value = req.Value

		setting, err = api.services.SettingSvc.Update(setting)
		if err != nil {
			api.error(w, r, "Could not update setting.", err, http.StatusInternalServerError)
			return
		}

		api.respond(w, r, response{setting}, status)
	}
}

func (api *API) DeleteSetting() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		settingKey := vars["id"]

		setting, err := api.services.SettingSvc.Get([]byte(settingKey))
		if err != nil {
			api.error(w, r, "Could not get setting.", err, http.StatusInternalServerError)
			return
		}

		if setting == nil {
			api.error(w, r, "No setting found with that ID.", fmt.Errorf("no setting with that ID"), http.StatusNotFound)
			return
		}

		err = api.services.SettingSvc.Delete([]byte(settingKey))
		if err != nil {
			api.error(w, r, "Could not delete setting.", err, http.StatusInternalServerError)
			return
		}

		api.respond(w, r, nil, http.StatusNoContent)
	}
}
