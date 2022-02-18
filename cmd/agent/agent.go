package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/davecgh/go-spew/spew"
	"zerosrealm.xyz/tergum/internal/agent"
	"zerosrealm.xyz/tergum/internal/agent/config"
	"zerosrealm.xyz/tergum/internal/entity"
	"zerosrealm.xyz/tergum/internal/log"
)

func registerAgent(log *log.Logger, conf *config.Config) error {
	if conf.Registration == "" {
		log.WithFields("function", "registerAgent").Debug("No registration token, skipping.")
		return nil
	}

	log.WithFields("function", "registerAgent").Debug("Registering agent")

	type registrationData struct {
		Hostname string `json:"hostname"`
		Port     int    `json:"port"`
		Token    string `json:"token"`
	}
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("registerAgent(): error getting hostname: %w", err)
	}

	msg, err := json.Marshal(registrationData{
		Hostname: hostname,
		Port:     conf.Listen.Port,
		Token:    conf.Registration,
	})
	if err != nil {
		return fmt.Errorf("registerAgent(): error marshalling registration data: %w", err)
	}

	req, err := http.NewRequest("POST", conf.Server+"/api/register", bytes.NewReader(msg))
	if err != nil {
		return fmt.Errorf("registerAgent(): error creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("registerAgent(): error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.WithFields("function", "registerAgent").Error("non-200 status:", resp.Status, "body read error:", err)
			return nil
		}
		log.WithFields("function", "registerAgent").Info("non-200 status:", resp.Status, "body:", string(body))
		return fmt.Errorf("registerAgent(): non-200 status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("registerAgent(): error reading response body: %w", err)
	}

	var agent *entity.Agent
	err = json.Unmarshal(body, &agent)
	if err != nil {
		return fmt.Errorf("registerAgent(): error unmarshalling response body: %w", err)
	}
	log.WithFields("function", "registerAgent").Debug("Body returned:", spew.Sdump(string(body)))

	conf.PSK = agent.PSK

	log.WithFields("function", "registerAgent").Debug("Agent registered:", spew.Sdump(agent))

	return nil
}

func main() {
	conf, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := log.New(&conf.Log)
	if err != nil {
		fmt.Println("Error creating logger:", err)
		return
	}

	log.Info("Starting agent")

	if conf.PSK == "" && conf.Registration == "" {
		log.Info("no PSK defined - exiting")
		return
	}

	if conf.Restic == "" {
		log.Info("no path to restic defined - exiting")
		return
	}

	if conf.Server == "" {
		log.Info("no server defined - exiting")
		return
	}

	err = registerAgent(log, conf)
	if err != nil {
		log.Error("error registering agent:", err)
		return
	}

	server, err := agent.NewServer(conf)
	if err != nil {
		log.Error("error creating server:", err)
		return
	}

	server.Start()
}
