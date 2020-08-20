package config

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/oldfritter/sneaker-go/v3"
	"gopkg.in/yaml.v2"
)

var (
	RabbitMqConnect sneaker.RabbitMqConnect
)

var AmqpGlobalConfig struct {
	Connect struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Vhost    string `yaml:"vhost"`
	} `yaml:"connect"`

	Exchange map[string]map[string]string `yaml:"exchange"`
	Queue    map[string]map[string]string `yaml:"queue"`
}

func InitAmqpConfig() {
	pathStr, _ := filepath.Abs("config/amqp.yml")
	content, err := ioutil.ReadFile(pathStr)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = yaml.Unmarshal(content, &AmqpGlobalConfig)
	if err != nil {
		log.Fatal(err)
		return
	}
}
