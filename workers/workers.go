package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"

	envConfig "github.com/oldfritter/goDCE/config"
	"github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
	"github.com/oldfritter/goDCE/workers/sneakerWorkers"
	"github.com/streadway/amqp"
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

	err := ioutil.WriteFile("pids/workers.pid", []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func closeResource() {
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
		for i := 0; i < w.Threads; i++ {
			go func() {
				w.SubscribeMessageByQueue(amqp.Table{})
			}()
		}
	}
}
