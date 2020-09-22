package initializers

import (
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"

	"github.com/oldfritter/goDCE/config"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func LoadLatestTickers() {
	go func() {
		channel, err := config.RabbitMqConnect.Channel()
		if err != nil {
			fmt.Errorf("Channel: %s", err)
		}
		channel.ExchangeDeclare(config.AmqpGlobalConfig.Exchange["fanout"]["ticker"], "fanout", true, false, false, false, nil)
		queue, err := channel.QueueDeclare("", true, true, false, false, nil)
		if err != nil {
			return
		}
		channel.QueueBind(queue.Name, queue.Name, config.AmqpGlobalConfig.Exchange["fanout"]["ticker"], false, nil)
		msgs, _ := channel.Consume(queue.Name, "", true, false, false, false, nil)
		for d := range msgs {
			var ticker Ticker
			json.Unmarshal(d.Body, &ticker)
			for i, _ := range AllMarkets {
				if AllMarkets[i].Id == ticker.MarketId {
					AllMarkets[i].Ticker = ticker.TickerAspect
				}
			}
		}
		return
	}()

	go func() {
		tickerRedis := utils.GetRedisConn("ticker")
		defer tickerRedis.Close()
		for i, market := range AllMarkets {
			jsonStr, err := redis.Bytes(tickerRedis.Do("GET", market.TickerRedisKey()))
			if err != nil {
				continue
			}
			var ticker Ticker
			json.Unmarshal(jsonStr, &ticker)
			if AllMarkets[i].Ticker == nil {
				AllMarkets[i].Ticker = ticker.TickerAspect
			}
		}
	}()
}
