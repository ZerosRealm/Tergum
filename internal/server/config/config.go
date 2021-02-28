package config

import (
	"github.com/jinzhu/configor"
)

// Config for storing settings.
type Config struct {
	Listen struct {
		IP   string
		Port int
	}
	Restic string
}

// Load config.
func Load() Config {
	var conf Config
	configor.Load(&conf, "config.yml")

	if conf.Listen.IP == "" {
		conf.Listen.IP = "127.0.0.1"
	}
	if conf.Listen.Port == 0 {
		conf.Listen.Port = 80
	}

	return conf
}
