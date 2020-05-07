package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"

	envConfig "github.com/oldfritter/goDCE/config"
	"github.com/oldfritter/goDCE/initializers"
	"github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/trade/matching"
	"github.com/oldfritter/goDCE/utils"
)

func main() {
	initialize()
	initAssignments()

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

	err := ioutil.WriteFile("pids/matching.pid", []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func closeResource() {
	initializers.CloseAmqpConnection()
	utils.CloseRedisPools()
	utils.CloseMainDB()
}

func initAssignments() {
	matching.InitAssignments()
	matching.SubscribeReload()

	go func() {
		channel, err := initializers.RabbitMqConnect.Channel()
		if err != nil {
			fmt.Errorf("Channel: %s", err)
		}
		queueName := initializers.AmqpGlobalConfig.Queue["matching"]["reload"]
		queue, err := channel.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			return
		}
		msgs, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
		for _ = range msgs {
			matching.InitAssignments()
		}
		return
	}()
}
