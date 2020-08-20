package initializers

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"github.com/oldfritter/sneaker-go/v3"
	"github.com/streadway/amqp"
	"gopkg.in/yaml.v2"

	"github.com/oldfritter/goDCE/config"
)

func InitializeAmqpConfig() {
	path_str, _ := filepath.Abs("config/amqp.yml")
	content, err := ioutil.ReadFile(path_str)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = yaml.Unmarshal(content, &config.AmqpGlobalConfig)
	if err != nil {
		log.Fatal(err)
		return
	}
	InitializeAmqpConnection()
}

func InitializeAmqpConnection() {
	var err error
	conn, err := amqp.Dial("amqp://" + config.AmqpGlobalConfig.Connect.Username + ":" + config.AmqpGlobalConfig.Connect.Password + "@" + config.AmqpGlobalConfig.Connect.Host + ":" + config.AmqpGlobalConfig.Connect.Port + "/" + config.AmqpGlobalConfig.Connect.Vhost)
	config.RabbitMqConnect = sneaker.RabbitMqConnect{conn}
	if err != nil {
		time.Sleep(5000)
		InitializeAmqpConnection()
		return
	}
	go func() {
		<-config.RabbitMqConnect.NotifyClose(make(chan *amqp.Error))
		InitializeAmqpConnection()
	}()
}

func CloseAmqpConnection() {
	config.RabbitMqConnect.Close()
}

func GetRabbitMqConnect() sneaker.RabbitMqConnect {
	return config.RabbitMqConnect
}
