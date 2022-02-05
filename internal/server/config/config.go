package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jinzhu/configor"
	"zerosrealm.xyz/tergum/internal/log"
)

type dbConfig struct {
	Driver         string `default:"memory"`
	DataSourceName string `default:""`
}

// Config for storing settings.
type Config struct {
	Listen struct {
		IP   string `default:"127.0.0.1"`
		Port int    `default:"8080"`
	}
	Restic   string
	Cache    string
	Database dbConfig
	Log      log.Config
}

// Load config.
func Load() (*Config, error) {
	var conf Config
	if _, err := os.Stat("config.yml"); !os.IsNotExist(err) {
		configor.Load(&conf, "config.yml")
	}

	ip := os.Getenv("TERGUM_IP")
	port := os.Getenv("TERGUM_PORT")
	restic := os.Getenv("TERGUM_RESTIC")

	if ip != "" {
		conf.Listen.IP = ip
	}
	if port != "" {
		num, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("TERGUM_PORT is not an integer")
		}
		conf.Listen.Port = num
	}
	if restic != "" {
		conf.Restic = restic
	}

	return &conf, nil
}
