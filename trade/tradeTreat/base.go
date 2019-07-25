package tradeTreat

import (
	"fmt"
	"log"
	"os"

	envConfig "github.com/oldfritter/goDCE/config"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
	"github.com/streadway/amqp"
)

var (
	Assignments = make(map[int]Market)
)

func InitAssignments() {

	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	var markets []Market
	mainDB.Where("trade_treat_node = ?", envConfig.CurrentEnv.Node).Find(&markets)
	for _, market := range markets {
		market.Running = Assignments[market.Id].Running
		if market.MatchingAble && !market.Running {
			Assignments[market.Id] = market
		} else if !market.MatchingAble {
			delete(Assignments, market.Id)
		}
	}
	mainDB.DbRollback()
	for id, assignment := range Assignments {
		if assignment.MatchingAble && !assignment.Running {
			go func(id int) {
				a := Assignments[id]
				subscribeMessageByQueue(&a, amqp.Table{})
			}(id)
			assignment.Running = true
			Assignments[id] = assignment
		}
	}
}

func subscribeMessageByQueue(assignment *Market, arguments amqp.Table) error {
	channel, err := utils.RabbitMqConnect.Channel()
	if err != nil {
		fmt.Errorf("Channel: %s", err)
	}
	queueName := (*assignment).TradeTreatQueue()
	queue, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		arguments,
	)
	if err != nil {
		return fmt.Errorf("Queue Declare: %s", err)
	}

	channel.ExchangeDeclare((*assignment).TradeTreatExchange(), "topic", (*assignment).Durable, false, false, false, nil)
	channel.QueueBind((*assignment).TradeTreatQueue(), (*assignment).Code, (*assignment).TradeTreatExchange(), false, nil)

	msgs, err := channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	go func(id int) {
		for d := range msgs {
			a := Assignments[id]

			logFile, err := os.Create(a.TradeTreatLogFilePath())
			defer logFile.Close()
			if err != nil {
				log.Fatalln("open log file error !")
			}
			workerLog := log.New(logFile, "[Info]", log.LstdFlags)
			workerLog.SetPrefix("[Info]")

			Treat(&d.Body, workerLog)
			d.Ack(a.Ack)
		}
		return
	}(assignment.Id)

	return nil
}

func SubscribeReload() (err error) {
	channel, err := utils.RabbitMqConnect.Channel()
	if err != nil {
		fmt.Errorf("Channel: %s", err)
		return
	}
	channel.ExchangeDeclare(utils.AmqpGlobalConfig.Exchange.Default["key"], "topic", true, false, false, false, nil)
	channel.QueueBind(utils.AmqpGlobalConfig.Queue.Trade["reload"], utils.AmqpGlobalConfig.Queue.Trade["reload"], utils.AmqpGlobalConfig.Exchange.Default["key"], false, nil)
	return
}
