package sneakerWorkers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/oldfritter/goDCE/config"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
	sneaker "github.com/oldfritter/sneaker-go/v3"
	"github.com/streadway/amqp"
)

func InitializeTickerWorker() {
	for _, w := range config.AllWorkers {
		if w.Name == "TickerWorker" {
			config.AllWorkerIs = append(config.AllWorkerIs, &TickerWorker{w})
			return
		}
	}
}

type TickerWorker struct {
	sneaker.Worker
}

func (worker *TickerWorker) Work(payloadJson *[]byte) (err error) {
	var payload struct {
		MarketId int `json:"market_id"`
	}
	json.Unmarshal([]byte(*payloadJson), &payload)
	buildTicker(payload.MarketId)
	return
}

func buildTicker(marketId int) {
	market, err := FindMarketById(marketId)
	if err != nil {
		fmt.Println("error:", err)
	}
	dataRedis := utils.GetRedisConn("data")
	defer dataRedis.Close()
	ticker := refreshTicker(dataRedis, &market)
	t, err := json.Marshal(ticker)
	if err != nil {
		fmt.Println("error:", err)
	}
	err = config.RabbitMqConnect.PublishMessageWithRouteKey(config.AmqpGlobalConfig.Exchange["fanout"]["ticker"], "#", "text/plain", false, false, &t, amqp.Table{}, amqp.Persistent, "")
	if err != nil {
		fmt.Println("{ error:", err, "}")
	}
	dataRedis.Do("SET", market.TickerRedisKey(), string(t))

}

func refreshTicker(dataRedis redis.Conn, market *Market) (ticker Ticker) {
	now := time.Now()
	ticker.MarketId = (*market).Id
	ticker.At = now.Unix()
	ticker.Name = (*market).Name
	kJsons, _ := redis.Values(dataRedis.Do("ZRANGEBYSCORE", (*market).KLineRedisKey(1), now.Add(-time.Hour*24).Unix(), now.Unix()))
	var k KLine
	for i, kJson := range kJsons {
		json.Unmarshal(kJson.([]byte), &k)
		if i == 0 {
			ticker.TickerAspect.Open = k.Open
		}
		ticker.TickerAspect.Last = k.Close
		if ticker.TickerAspect.Low.IsZero() || ticker.TickerAspect.Low.GreaterThan(k.Low) {
			ticker.TickerAspect.Low = k.Low
		}
		if ticker.TickerAspect.High.LessThan(k.High) {
			ticker.TickerAspect.High = k.High
		}
		ticker.TickerAspect.Volume = ticker.TickerAspect.Volume.Add(k.Volume)
	}
	return
}
