package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"

	sneaker "github.com/oldfritter/sneaker-go"
	"github.com/streadway/amqp"

	envConfig "github.com/oldfritter/goDCE/config"
	"github.com/oldfritter/goDCE/initializers"
	"github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
	"github.com/oldfritter/goDCE/workers/sneakerWorkers"
)

func main() {
	initialize()
	initWorkers()

	StartAllWorkers()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	closeResource()
}

func initialize() {
	envConfig.InitEnv()
	utils.InitMainDB()
	utils.InitBackupDB()
	models.AutoMigrations()
	utils.InitRedisPools()
	initializers.InitializeAmqpConfig()
	initializers.LoadCacheData()
	initializers.InitializeAmqpConfig()

	setLog()
	err := ioutil.WriteFile("pids/workers.pid", []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		log.Println(err)
	}
}

func closeResource() {
	initializers.CloseAmqpConnection()
	utils.CloseRedisPools()
	utils.CloseMainDB()
	utils.CloseBackupDB()
}

func initWorkers() {
	sneakerWorkers.InitWorkers()
}

func StartAllWorkers() {
	for _, w := range sneakerWorkers.AllWorkers {
		for i := 0; i < w.GetThreads(); i++ {
			go func(w sneakerWorkers.Worker) {
				sneaker.SubscribeMessageByQueue(initializers.RabbitMqConnect.Connection, w, amqp.Table{})
			}(w)
		}
	}
}

func setLog() {
	err := os.Mkdir("logs", 0755)
	if err != nil {
		if !os.IsExist(err) {
			log.Fatalf("create folder error: %v", err)
		}
	}

	file, err := os.OpenFile("logs/workers.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("open file error: %v", err)
	}
	log.SetOutput(file)

}
