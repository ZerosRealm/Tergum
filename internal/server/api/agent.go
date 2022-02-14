package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"zerosrealm.xyz/tergum/internal/entities"
)

func (api *API) GetAgents() http.HandlerFunc {
	type response struct {
		Agents []*entities.Agent `json:"agents"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		agents, err := api.services.AgentSvc.GetAll()
		if err != nil {
			api.error(w, r, "Could not get agents.", err, http.StatusInternalServerError)
			return
		}

		if agents == nil {
			agents = make([]*entities.Agent, 0)
		}

		api.respond(w, r, &response{Agents: agents}, http.StatusOK)
	}
}

func (api *API) CreateAgent() http.HandlerFunc {
	type request struct {
		Name string `json:"name"`
		PSK  string `json:"psk"`
		IP   string `json:"ip"`
		Port int    `json:"port"`
	}
	type response struct {
		Agent *entities.Agent `json:"agent"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		agent := &entities.Agent{
			Name: req.Name,
			PSK:  req.PSK,
			IP:   req.IP,
			Port: req.Port,
		}

		agent, err = api.services.AgentSvc.Create(agent)
		if err != nil {
			api.error(w, r, "Could not create agent.", err, http.StatusInternalServerError)
			return
		}

		r.Header.Add("Location", fmt.Sprintf("/agent/%d", agent.ID))
		api.respond(w, r, response{Agent: agent}, http.StatusCreated)
	}
}

func (api *API) UpdateAgent() http.HandlerFunc {
	type request struct {
		Name string `json:"name"`
		PSK  string `json:"psk"`
		IP   string `json:"ip"`
		Port int    `json:"port"`
	}
	type response struct {
		Agent *entities.Agent `json:"agent"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		agentID := vars["id"]

		var req request
		err := api.decode(w, r, &req)
		if err != nil {
			api.error(w, r, msgDecodeError, err, http.StatusBadRequest)
			return
		}

		agent, err := api.services.AgentSvc.Get([]byte(agentID))
		if err != nil {
			api.error(w, r, "Could not get agent.", err, http.StatusInternalServerError)
			return
		}

		if agent == nil {
			api.error(w, r, "No agent found with that ID.", fmt.Errorf("no agent with that ID"), http.StatusNotFound)
			return
		}

		status := http.StatusOK
		// TODO: Create an agent with the given ID if it does not exist
		// if foundAgent == nil {
		// 	foundAgent = &entities.Agent{
		// 		ID: id,
		// 	}
		// 	status = http.StatusCreated
		// 	savedData.AgentIncrement++
		// 	savedData.Agents = append(savedData.Agents, foundAgent)
		// }

		agent.Name = req.Name
		agent.PSK = req.PSK
		agent.IP = req.IP
		agent.Port = req.Port

		agent, err = api.services.AgentSvc.Update(agent)
		if err != nil {
			api.error(w, r, "Could not update agent.", err, http.StatusInternalServerError)
			return
		}

		api.respond(w, r, response{Agent: agent}, status)
	}
}

func (api *API) DeleteAgent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		agentID := vars["id"]

		agent, err := api.services.AgentSvc.Get([]byte(agentID))
		if err != nil {
			api.error(w, r, "Could not get agent.", err, http.StatusInternalServerError)
			return
		}

		if agent == nil {
			api.error(w, r, "No agent found with that ID.", fmt.Errorf("no agent with that ID"), http.StatusNotFound)
			return
		}

		err = api.services.AgentSvc.Delete([]byte(agentID))
		if err != nil {
			api.error(w, r, "Could not delete agent.", err, http.StatusInternalServerError)
			return
		}

		api.respond(w, r, nil, http.StatusNoContent)
	}
}
