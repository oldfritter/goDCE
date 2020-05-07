package matching

import (
	"encoding/json"
	"fmt"
	"runtime"

	envConfig "github.com/oldfritter/goDCE/config"
	"github.com/oldfritter/goDCE/initializers"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
	"github.com/oldfritter/matching"
	"github.com/streadway/amqp"
)

var (
	Assignments = make(map[int]Market)
	Engines     = make(map[int]matching.Engine)
)

func InitAssignments() {

	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	var markets []Market
	mainDB.Where("matching_node = ?", envConfig.CurrentEnv.Node).Find(&markets)
	for _, market := range markets {
		market.Running = Assignments[market.Id].Running
		if market.MatchingAble && !market.Running {
			Assignments[market.Id] = market
			engine := matching.InitializeEngine(market.Id, matching.Options{})
			Engines[market.Id] = engine
			// 加载过往订单
			var orders []Order
			if !mainDB.Where("state = ?", 100).Where("market_id = ?", market.Id).Find(&orders).RecordNotFound() {
				for _, o := range orders {
					order, _ := matching.InitializeOrder(o.OrderAttrs())
					engine.Submit(order)
				}
			}
		} else if !market.MatchingAble {
			delete(Assignments, market.Id)
			delete(Engines, market.Id)
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
	channel, err := initializers.RabbitMqConnect.Channel()
	if err != nil {
		fmt.Errorf("Channel: %s", err)
	}
	channel.ExchangeDeclare((*assignment).MatchingExchange(), "topic", (*assignment).Durable, false, false, false, nil)
	channel.QueueBind((*assignment).MatchingQueue(), (*assignment).Code, (*assignment).MatchingExchange(), false, nil)

	go func(id int) {
		a := Assignments[id]
		channel, err := initializers.RabbitMqConnect.Channel()
		if err != nil {
			fmt.Errorf("Channel: %s", err)
		}
		msgs, _ := channel.Consume(
			a.MatchingQueue(), // queue
			"",                // consumer
			false,             // auto-ack
			false,             // exclusive
			false,             // no-local
			false,             // no-wait
			nil,               // args
		)

		for d := range msgs {
			if !a.MatchingAble {
				d.Reject(true)
				d.Nack(false, false)
				runtime.Goexit()
			}
			doMatching(&d.Body)
			d.Ack(a.Ack)
		}
		return
	}((*assignment).Id)

	engine := Engines[(*assignment).Id]
	buildDepth(assignment)

	// trade
	go func() {
		for offer := range engine.Traded {
			b, err := json.Marshal(offer)
			if err != nil {
				fmt.Println("error:", err)
			}
			err = initializers.PublishMessageWithRouteKey((*assignment).TradeTreatExchange(), (*assignment).Code, "text/plain", &b, amqp.Table{}, amqp.Persistent)
			if err != nil {
				fmt.Println("{ error:", err, "}")
			} else {
				buildDepth(assignment)
				buildLatestTrades(assignment)
			}
		}
	}()

	// cancel
	go func() {
		for offer := range engine.Canceled {
			b, err := json.Marshal(offer)
			if err != nil {
				fmt.Println("error:", err)
			}
			err = initializers.PublishMessageWithRouteKey((*assignment).OrderCancelExchange(), (*assignment).Code, "text/plain", &b, amqp.Table{}, amqp.Persistent)
			if err != nil {
				fmt.Println("{ error:", err, "}")
			} else {
				buildDepth(assignment)
			}
		}
	}()

	return nil
}

func SubscribeReload() (err error) {
	channel, err := initializers.RabbitMqConnect.Channel()
	if err != nil {
		fmt.Errorf("Channel: %s", err)
		return
	}
	channel.ExchangeDeclare(initializers.AmqpGlobalConfig.Exchange["default"]["key"], "topic", true, false, false, false, nil)
	channel.QueueBind(initializers.AmqpGlobalConfig.Queue["matching"]["reload"], initializers.AmqpGlobalConfig.Queue["matching"]["reload"], initializers.AmqpGlobalConfig.Exchange["default"]["key"], false, nil)
	return
}
