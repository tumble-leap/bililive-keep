package main

import (
	"os"

	_ "embed"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var configFile = "config.yaml"

//go:embed config_example.yaml
var ExampleCfg string

var cfg Cfg

func LoadConfig() error {
	result, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	yaml.Unmarshal(result, &cfg)
	return nil
}

func CreateBlankConfigfile() {
	os.Create(configFile)
	f, err := os.OpenFile(configFile, os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	f.WriteString(ExampleCfg)
}
