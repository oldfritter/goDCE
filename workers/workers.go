package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"

	sneaker "github.com/oldfritter/sneaker-go"
	sneakerUtils "github.com/oldfritter/sneaker-go/utils"
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
	utils.InitializeAmqpConfig()
	initializers.LoadCacheData()
	sneakerUtils.InitializeAmqpConfig()

	err := ioutil.WriteFile("pids/workers.pid", []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func closeResource() {
	initializers.DeleteListeQueue()
	utils.CloseAmqpConnection()
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
				sneaker.SubscribeMessageByQueue(w, amqp.Table{})
			}(w)
		}
	}
}
