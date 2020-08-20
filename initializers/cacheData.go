package initializers

import (
	"fmt"

	"github.com/oldfritter/goDCE/config"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func InitCacheData() {
	db := utils.MainDbBegin()
	defer db.DbRollback()
	InitAllCurrencies(db)
	InitAllMarkets(db)
}

func LoadCacheData() {
	go func() {
		channel, err := config.RabbitMqConnect.Channel()
		if err != nil {
			fmt.Errorf("Channel: %s", err)
		}
		channel.ExchangeDeclare(config.AmqpGlobalConfig.Exchange["fanout"]["name"], "fanout", true, false, false, false, nil)
		queue, err := channel.QueueDeclare("", true, true, false, false, nil)
		if err != nil {
			return
		}
		channel.QueueBind(queue.Name, queue.Name, config.AmqpGlobalConfig.Exchange["fanout"]["name"], false, nil)
		msgs, _ := channel.Consume(queue.Name, "", true, true, false, false, nil)
		for _ = range msgs {
			InitCacheData()
		}
		return
	}()

}
