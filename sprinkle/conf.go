package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type CircuitConfig struct {
	Name        string
	BcmPort     int `yaml:"bcmPort"`
	MaxDuration int `yaml:"maxDuration"`
	Serial      string
}

type BridgeConfig struct {
	Name   string
	Serial string
}

type Config struct {
	Circuits     []CircuitConfig
	Bridge       BridgeConfig
	Manufacturer string
}

func ParseConfig(filename string) *Config {
	if filename == "" {
		filename = "sprinkle.conf.yml"
	}

	file, err := os.Open(filename) // For read access.
	if err != nil {
		log.Panic(err)
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Panic(err)
	}

	var conf Config

	err = yaml.Unmarshal(bytes, &conf)
	if err != nil {
		log.Panic(err)
	}

	return &conf
}

func (conf *Config) GetUsedPorts() (ports []int) {
	for _, c := range conf.Circuits {
		ports = append(ports, c.BcmPort)
	}
	return ports
}

var bcmPorts = []int{2, 3, 4, 14, 15, 17, 18, 27, 22, 23, 24, 10, 9, 25, 11, 8, 7}

func GenerateConfig(outfile string, numCircuits int) {
	var circuits []CircuitConfig

	if numCircuits > len(bcmPorts) {
		panic(errors.New("Not enough BCM ports to generate this many circuits."))
	}

	for i := 0; i < numCircuits; i++ {
		circuits = append(circuits, CircuitConfig{
			Name:    fmt.Sprintf("Circuit %d", i+1),
			BcmPort: bcmPorts[i],
			Serial:  fmt.Sprintf("circuit-%05d", i),
		})
	}

	conf := Config{
		Manufacturer: "My Name",
		Bridge: BridgeConfig{
			Name:   "MyBridge",
			Serial: "bridge-00001",
		},
		Circuits: circuits,
	}

	bytes, err := yaml.Marshal(conf)
	if err != nil {
		log.Panic(err)
	}

	file, err := os.Create(outfile)
	if err != nil {
		log.Panic(err)
	}

	defer file.Close()
	file.Write(bytes)

}
