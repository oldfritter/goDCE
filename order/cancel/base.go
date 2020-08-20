package cancel

import (
	"fmt"

	"github.com/streadway/amqp"

	envConfig "github.com/oldfritter/goDCE/config"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

var (
	Assignments = make(map[int]Market)
)

func InitAssignments() {

	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	var markets []Market
	mainDB.Where("order_cancel_node = ?", envConfig.CurrentEnv.Node).Find(&markets)
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
	channel, err := envConfig.RabbitMqConnect.Channel()
	if err != nil {
		fmt.Errorf("Channel: %s", err)
	}

	channel.ExchangeDeclare((*assignment).OrderCancelExchange(), "topic", (*assignment).Durable, false, false, false, nil)
	channel.QueueBind((*assignment).OrderCancelQueue(), (*assignment).Code, (*assignment).OrderCancelExchange(), false, nil)

	go func(id int) {
		channel, err := envConfig.RabbitMqConnect.Channel()
		if err != nil {
			fmt.Errorf("Channel: %s", err)
		}
		a := Assignments[id]
		msgs, err := channel.Consume(
			a.OrderCancelQueue(), // queue
			"",                   // consumer
			false,                // auto-ack
			false,                // exclusive
			false,                // no-local
			false,                // no-wait
			nil,                  // args
		)
		for d := range msgs {
			Cancel(&d.Body)
			d.Ack(a.Ack)
		}
		return
	}(assignment.Id)

	return nil
}

func SubscribeReload() (err error) {
	channel, err := envConfig.RabbitMqConnect.Channel()
	if err != nil {
		fmt.Errorf("Channel: %s", err)
		return
	}
	channel.ExchangeDeclare(envConfig.AmqpGlobalConfig.Exchange["default"]["key"], "topic", true, false, false, false, nil)
	channel.QueueBind(envConfig.AmqpGlobalConfig.Queue["cancel"]["reload"], envConfig.AmqpGlobalConfig.Queue["cancel"]["reload"], envConfig.AmqpGlobalConfig.Exchange["default"]["key"], false, nil)
	return
}
