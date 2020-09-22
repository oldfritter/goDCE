package initializers

import (
	"encoding/json"
	"fmt"

	"github.com/oldfritter/goDCE/config"
	. "github.com/oldfritter/goDCE/models"
)

func LoadLatestKLines() {
	go func() {
		channel, err := config.RabbitMqConnect.Channel()
		if err != nil {
			fmt.Errorf("Channel: %s", err)
		}
		channel.ExchangeDeclare(config.AmqpGlobalConfig.Exchange["fanout"]["k"], "fanout", true, false, false, false, nil)
		queue, err := channel.QueueDeclare("", true, true, false, false, nil)
		if err != nil {
			return
		}
		channel.QueueBind(queue.Name, queue.Name, config.AmqpGlobalConfig.Exchange["fanout"]["k"], false, nil)
		msgs, _ := channel.Consume(queue.Name, "", true, false, false, false, nil)
		for d := range msgs {
			var notifyKLine KLine
			json.Unmarshal(d.Body, &notifyKLine)
			for i, _ := range AllMarkets {
				if AllMarkets[i].Id == notifyKLine.MarketId {
					AllMarkets[i].LatestKLines[notifyKLine.Period] = notifyKLine
				}
			}
		}
		return
	}()
}
