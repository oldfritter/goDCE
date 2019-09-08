package config

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Env struct {
	Model string `yaml:"model"`
	Node  string `yaml:"node"`

	Newrelic struct {
		AppName    string `yaml:"app_name"`
		LicenseKey string `yaml:"license_key"`
	} `yaml:"newrelic"`

	Schedules []string `yaml:"schedules"`
}

var CurrentEnv Env

func InitEnv() {
	path_str, _ := filepath.Abs("config/env.yml")
	content, err := ioutil.ReadFile(path_str)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = yaml.Unmarshal(content, &CurrentEnv)
	if err != nil {
		log.Fatal(err)
	}
}
