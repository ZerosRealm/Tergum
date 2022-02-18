package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"zerosrealm.xyz/tergum/internal/entity"
)

func (api *API) GetForget() http.HandlerFunc {
	type request struct{}
	type response struct {
		Forget *entity.Forget `json:"forget"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		forgetID := vars["id"]

		forget, err := api.services.ForgetSvc.Get([]byte(forgetID))
		if err != nil {
			api.error(w, r, "Could not get forget policy.", err, http.StatusInternalServerError)
			return
		}

		if forget == nil {
			api.error(w, r, "No forget policy found with that ID.", fmt.Errorf("no forget found with that ID"), http.StatusNotFound)
			return
		}

		api.respond(w, r, response{Forget: forget}, http.StatusOK)
	}
}

func (api *API) UpdateForget() http.HandlerFunc {
	type request struct {
		entity.Forget
	}
	type response struct {
		Forget *entity.Forget `json:"forget"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		forgetID := vars["id"]

		var req request
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		forget, err := api.services.ForgetSvc.Get([]byte(forgetID))
		if err != nil {
			api.error(w, r, "Could not get forget policy.", err, http.StatusInternalServerError)
			return
		}

		if forget == nil {
			api.error(w, r, "No forget policy with that ID.", fmt.Errorf("no forget found with that ID"), http.StatusNotFound)
			return
		}

		forget.Enabled = req.Enabled
		forget.LastX = req.LastX
		forget.Hourly = req.Hourly
		forget.Daily = req.Daily
		forget.Weekly = req.Weekly
		forget.Monthly = req.Monthly
		forget.Yearly = req.Yearly

		forget, err = api.services.ForgetSvc.Update(forget)
		if err != nil {
			api.error(w, r, "Could not update forget policy.", err, http.StatusInternalServerError)
			return
		}

		api.respond(w, r, response{Forget: forget}, http.StatusOK)
	}
}
