package config

import (
	"log"
	"os"
	"strconv"

	"github.com/jinzhu/configor"
)

type dbConfig struct {
	Driver         string `default:"memory"`
	DataSourceName string `default:""`
}

// Config for storing settings.
type Config struct {
	Listen struct {
		IP   string
		Port int
	}
	Restic   string
	Cache    string
	Database dbConfig
}

// Load config.
func Load() Config {
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
			log.Fatal("TERGUM_PORT is not an integer")
		}
		conf.Listen.Port = num
	}
	if restic != "" {
		conf.Restic = restic
	}

	if conf.Listen.IP == "" {
		conf.Listen.IP = "127.0.0.1"
	}
	if conf.Listen.Port == 0 {
		conf.Listen.Port = 8080
	}

	return conf
}
