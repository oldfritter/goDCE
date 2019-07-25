package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"

	envConfig "github.com/oldfritter/goDCE/config"
	"github.com/oldfritter/goDCE/trade/treat"
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
	utils.InitRedisPools()
	utils.InitializeAmqpConfig()

	err := ioutil.WriteFile("pids/treat.pid", []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func closeResource() {
	utils.CloseAmqpConnection()
	utils.CloseRedisPools()
	utils.CloseMainDB()
}

func initAssignments() {
	treat.InitAssignments()
	treat.SubscribeReload()

	go func() {
		channel, err := utils.RabbitMqConnect.Channel()
		if err != nil {
			fmt.Errorf("Channel: %s", err)
		}
		queueName := utils.AmqpGlobalConfig.Queue.Trade["reload"]
		queue, err := channel.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			return
		}
		msgs, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
		for _ = range msgs {
			treat.InitAssignments()
		}
		return
	}()
}
