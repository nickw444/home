package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type accessoryConfig struct {
	Model  string
	Serial string
	Name   string
	Conf   map[string]interface{}
}

type serviceConfig struct {
	Type string
	ID   string
	Conf map[string]interface{}
}

type bridgeConfig struct {
	Name         string
	Manufacturer string
	Model        string
	Accessories  []*accessoryConfig
	Services     []*serviceConfig
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
