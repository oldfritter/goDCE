package sneakerWorkers

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/oldfritter/goDCE/utils"
	"github.com/streadway/amqp"
	"gopkg.in/yaml.v2"
)

var (
	AllWorkers []Worker
)

type Worker struct {
	Name       string            `yaml:"name"`
	Exchange   string            `yaml:"exchange"`
	RoutingKey string            `yaml:"routing_key"`
	Queue      string            `yaml:"queue"`
	Durable    bool              `yaml:"durable"`
	Ack        bool              `yaml:"ack"`
	Options    map[string]string `yaml:"options"`
	Arguments  map[string]string `yaml:"arguments"`
	Steps      []int             `yaml:"steps"`
	Threads    int               `yaml:"threads"`
	Log        string            `yaml:"log"`
	Logger     *log.Logger
}

func InitWorkers() {
	path_str, _ := filepath.Abs("config/workers.yml")
	content, err := ioutil.ReadFile(path_str)
	if err != nil {
		log.Fatal(err)
	}
	yaml.Unmarshal(content, &AllWorkers)
}

func (worker *Worker) SubscribeMessageByQueue(arguments amqp.Table) error {
	channel, err := utils.RabbitMqConnect.Channel()
	if err != nil {
		fmt.Errorf("Channel: %s", err)
	}

	if (*worker).Exchange != "" && (*worker).RoutingKey != "" {
		channel.ExchangeDeclare((*worker).Exchange, "topic", (*worker).Durable, false, false, false, nil)
		channel.QueueBind((*worker).Queue, (*worker).RoutingKey, (*worker).Exchange, false, nil)
		channel.ExchangeDeclare((*worker).Arguments["x-dead-letter-exchange"], "topic", (*worker).Durable, false, false, false, nil)
		channel.QueueBind((*worker).Queue, "#", (*worker).Arguments["x-dead-letter-exchange"], false, nil)
	}
	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)

	for i, step := range (*worker).Steps {
		_, err = channel.QueueDeclare(
			(*worker).Arguments["x-dead-letter-exchange"]+"-"+strconv.Itoa(i+1), // queue name
			(*worker).Durable, // durable
			false,             // delete when usused
			false,             // exclusive
			false,             // no-wait
			amqp.Table{"x-dead-letter-exchange": (*worker).Arguments["x-dead-letter-exchange"], "x-message-ttl": int32(step)}, // arguments
		)
		if err != nil {
			return fmt.Errorf("Queue Declare: %s", err)
		}
	}

	go func(queue string) {
		channel, err := utils.RabbitMqConnect.Channel()
		if err != nil {
			fmt.Errorf("Channel: %s", err)
		}
		msgs, _ := channel.Consume(
			queue, // queue
			"",    // consumer
			false, // auto-ack
			false, // exclusive
			false, // no-local
			false, // no-wait
			nil,   // args
		)
		for _, w := range AllWorkers {
			if w.Queue == queue {
				for d := range msgs {
					response := reflect.ValueOf(&w).MethodByName(w.Name).Call([]reflect.Value{reflect.ValueOf(&d.Body)})
					if !(response[0].String() == "") && !response[1].IsNil() {
						retry(response[0].String(), response[1].Bytes())
					}
					d.Ack(w.Ack)
				}
			}
		}
	}(worker.Queue)

	return nil
}

func retry(queueName string, message []byte) error {
	channel, err := utils.RabbitMqConnect.Channel()
	defer channel.Close()
	err = (*channel).Publish(
		"",        // publish to an exchange
		queueName, // routing to 0 or more queues
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            message,
			DeliveryMode:    amqp.Persistent, // amqp.Persistent, amqp.Transient // 1=non-persistent, 2=persistent
			Priority:        0,               // 0-9
			// a bunch of application/implementation-specific fields
		},
	)
	if err != nil {
		return err
	}
	return nil
}
