package tradeMatching

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"

	envConfig "github.com/oldfritter/goDCE/config"
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
	mainDB.Where("matching_node = ?", envConfig.CurrentEnv.Matching.Node).Find(&markets)
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
	channel, err := utils.RabbitMqConnect.Channel()
	if err != nil {
		fmt.Errorf("Channel: %s", err)
	}
	queueName := (*assignment).MatchingQueue()
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

	channel.ExchangeDeclare((*assignment).MatchingExchange(), "topic", (*assignment).Durable, false, false, false, nil)
	channel.QueueBind((*assignment).MatchingQueue(), (*assignment).Code, (*assignment).MatchingExchange(), false, nil)

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
			if !a.MatchingAble {
				d.Reject(true)
				d.Nack(false, false)
				runtime.Goexit()
			}
			logFile, err := os.Create(a.MatchingLogFilePath())
			defer logFile.Close()
			if err != nil {
				log.Fatalln("open log file error !")
			}
			workerLog := log.New(logFile, "[Info]", log.LstdFlags)
			workerLog.SetPrefix("[Info]")
			doMatching(&d.Body, workerLog)
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
			err = utils.PublishMessageWithRouteKey((*assignment).TradeTreatExchange(), (*assignment).Code, "text/plain", &b, amqp.Table{}, amqp.Persistent)
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
			err = utils.PublishMessageWithRouteKey((*assignment).OrderCancelExchange(), (*assignment).Code, "text/plain", &b, amqp.Table{}, amqp.Persistent)
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
	channel, err := utils.RabbitMqConnect.Channel()
	if err != nil {
		fmt.Errorf("Channel: %s", err)
		return
	}
	channel.ExchangeDeclare(utils.AmqpGlobalConfig.Exchange.Default["key"], "topic", false, false, false, false, nil)
	channel.QueueBind(utils.AmqpGlobalConfig.Queue.Matching["reload"], utils.AmqpGlobalConfig.Queue.Matching["reload"], utils.AmqpGlobalConfig.Exchange.Default["key"], false, nil)
	return
}
