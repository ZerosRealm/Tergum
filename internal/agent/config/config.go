package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jinzhu/configor"
	"zerosrealm.xyz/tergum/internal/log"
)

// Config for storing settings.
type Config struct {
	Listen struct {
		IP   string `default:"127.0.0.1"`
		Port int    `default:"666"`
	}
	PSK          string
	Registration string
	Restic       string
	Server       string
	Log          log.Config
}

// Load config.
func Load() (*Config, error) {
	var conf Config
	if _, err := os.Stat("agent.yml"); !os.IsNotExist(err) {
		configor.Load(&conf, "agent.yml")
	}

	ip := os.Getenv("TERGUM_IP")
	psk := os.Getenv("TERGUM_PSK")
	port := os.Getenv("TERGUM_PORT")
	server := os.Getenv("TERGUM_SERVER")
	restic := os.Getenv("TERGUM_RESTIC")
	regToken := os.Getenv("TERGUM_REGISTRATION")

	if ip != "" {
		conf.Listen.IP = ip
	}
	if psk != "" {
		conf.PSK = psk
	}
	if port != "" {
		num, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("TERGUM_PORT is not an integer")
		}
		conf.Listen.Port = num
	}
	if server != "" {
		conf.Server = server
	}
	if restic != "" {
		conf.Restic = restic
	}
	if regToken != "" {
		conf.Registration = regToken
	}

	if conf.Listen.IP == "" {
		conf.Listen.IP = "127.0.0.1"
	}
	if conf.Listen.Port == 0 {
		conf.Listen.Port = 666
	}

	return &conf, nil
}
