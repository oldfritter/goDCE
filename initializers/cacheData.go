package initializers

import (
	"fmt"

	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

var QueueName string

func InitCacheData() {
	db := utils.MainDbBegin()
	defer db.DbRollback()
	InitAllCurrencies(db)
	InitAllMarkets(db)
}

func LoadCacheData() {
	go func() {
		channel, err := utils.RabbitMqConnect.Channel()
		if err != nil {
			fmt.Errorf("Channel: %s", err)
		}
		channel.ExchangeDeclare(utils.AmqpGlobalConfig.Exchange.Fanout["name"], "fanout", true, false, false, false, nil)
		queue, err := channel.QueueDeclare("", true, false, false, false, nil)
		if err != nil {
			return
		}
		QueueName = queue.Name
		channel.QueueBind(queue.Name, QueueName, utils.AmqpGlobalConfig.Exchange.Fanout["name"], false, nil)
		msgs, _ := channel.Consume(queue.Name, "", true, false, false, false, nil)
		for _ = range msgs {
			InitCacheData()
		}
		return
	}()

}

func DeleteListeQueue() {
	channel, err := utils.RabbitMqConnect.Channel()
	if err != nil {
		fmt.Errorf("Channel: %s", err)
	}
	channel.QueueDelete(QueueName, false, false, false)

}
