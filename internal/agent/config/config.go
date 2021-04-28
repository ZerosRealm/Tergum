package config

import (
	_ "embed"
	"log"
	"os"

	"github.com/jinzhu/configor"
)

// Config for storing settings.
type Config struct {
	Listen struct {
		IP   string
		Port int
	}
	PSK    string
	Restic string
	Server string
}

//go:embed agent.yml
var defaultConfig []byte

// Load config.
func Load() Config {
	if _, err := os.Stat("agent.yml"); os.IsNotExist(err) {
		f, err := os.Create("agent.yml")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		f.Write(defaultConfig)
	}
	var conf Config
	configor.Load(&conf, "agent.yml")

	if conf.Listen.IP == "" {
		conf.Listen.IP = "127.0.0.1"
	}
	if conf.Listen.Port == 0 {
		conf.Listen.Port = 666
	}

	return conf
}
