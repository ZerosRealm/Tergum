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
	Restic string
}

//go:embed config.yml
var defaultConfig []byte

// Load config.
func Load() Config {
	if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
		f, err := os.Create("config.yml")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		f.Write(defaultConfig)
	}

	var conf Config
	configor.Load(&conf, "config.yml")

	if conf.Listen.IP == "" {
		conf.Listen.IP = "127.0.0.1"
	}
	if conf.Listen.Port == 0 {
		conf.Listen.Port = 8080
	}

	return conf
}
