package main

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type blindConfig struct {
	Name        string
	Transmitter string
	Remote      int
	Channel     int
}

type transmitterConfig struct {
	Serial string
	ID     string
}

type bridgeConfig struct {
	Name         string
	Manufacturer string
	Model        string

	Blinds       []*blindConfig
	Transmitters []*transmitterConfig
}

func parseConfig(filename string) *bridgeConfig {
	file, err := os.Open(filename) // For read access.
	if err != nil {
		log.Panic(err)
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Panic(err)
	}

	var conf bridgeConfig
	err = yaml.Unmarshal(bytes, &conf)
	if err != nil {
		log.Panic(err)
	}

	return &conf
}
