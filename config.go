package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

const CONFIG_FILENAME = "config.yaml"

type Config struct {
	LambdaRoot string `yaml:"LAMBDA_TASK_ROOT"`
	Handler    string `yaml:"_HANDLER"`
}

func ParseConfig() map[string]Config {
	config := map[string]Config{}

	yamlFile, err := ioutil.ReadFile(CONFIG_FILENAME)
	if err != nil {
		fmt.Printf("Error reading YAML file: %s\n", err)
		return nil
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("%+v\n", config)

	return config
}
