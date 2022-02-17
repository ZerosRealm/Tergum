package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"zerosrealm.xyz/tergum/internal/entities"
)

func generatePSK(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (api *API) RegisterAgent() http.HandlerFunc {
	type request struct {
		Hostname string `json:"hostname"`
		Port     int    `json:"port"`
		Token    string `json:"token"`
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

		// Check if agent registration is enabled.
		setting, err := api.services.SettingSvc.Get([]byte("registration-enabled"))
		if err != nil {
			api.error(w, r, "Could not get registration-enabled setting.", err, http.StatusInternalServerError)
			return
		}

		if setting == nil {
			api.error(w, r, "registration-enabled setting not found.", fmt.Errorf("registration-enabled setting not found"), http.StatusInternalServerError)
			return
		}

		enabled := false
		err = json.Unmarshal(setting.Value, &enabled)
		if err != nil {
			api.error(w, r, "Could not parse registration-enabled setting.", err, http.StatusInternalServerError)
			return
		}

		if !enabled {
			api.error(w, r, "registration is disabled.", fmt.Errorf("registration is disabled"), http.StatusForbidden)
			return
		}

		// Get the token.
		setting, err = api.services.SettingSvc.Get([]byte("registration-token"))
		if err != nil {
			api.error(w, r, "Could not get registration-token setting.", err, http.StatusInternalServerError)
			return
		}

		if setting == nil {
			api.error(w, r, "registration-token setting not found.", fmt.Errorf("registration-token setting not found"), http.StatusInternalServerError)
			return
		}

		var token string
		err = json.Unmarshal(setting.Value, &token)
		if err != nil {
			api.error(w, r, "Could not parse registration-token setting.", err, http.StatusInternalServerError)
			return
		}

		// Check if the token is valid.
		if token != req.Token {
			api.error(w, r, "Invalid token.", fmt.Errorf("invalid token"), http.StatusForbidden)
			return
		}

		// TODO: Optimize with filters.
		agents, err := api.services.AgentSvc.GetAll()
		if err != nil {
			api.error(w, r, "Could not retrieve agents.", err, http.StatusInternalServerError)
			return
		}

		ip := strings.Split(r.RemoteAddr, ":")[0]

		for _, agent := range agents {
			if agent.Name == req.Hostname && agent.IP == ip && agent.Port == req.Port {
				api.respond(w, r, &response{Agent: agent}, http.StatusOK)
				return
			}
		}

		psk, err := generatePSK(64)
		if err != nil {
			api.error(w, r, "Could not generate PSK.", err, http.StatusInternalServerError)
			return
		}

		agent := &entities.Agent{
			Name: req.Hostname,
			IP:   ip,
			Port: req.Port,
			PSK:  psk,
		}

		agent, err = api.services.AgentSvc.Create(agent)
		if err != nil {
			api.error(w, r, "Could not create agent.", err, http.StatusInternalServerError)
			return
		}

		api.respond(w, r, response{Agent: agent}, http.StatusOK)
	}
}
