package treat

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/streadway/amqp"

	"github.com/oldfritter/goDCE/config"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
	"github.com/oldfritter/matching"
)

func Treat(payloadJson *[]byte) {
	var offer matching.Offer
	json.Unmarshal([]byte(*payloadJson), &offer)
	tryCreateTrade(&offer, 2)
	return
}

func tryCreateTrade(offer *matching.Offer, times int) {
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	var ask, bid Order
	if mainDB.Where("id = ?", offer.AskId).First(&ask).RecordNotFound() {
		return
	}
	if mainDB.Where("id = ?", offer.BidId).First(&bid).RecordNotFound() {
		return
	}

	trade := Trade{
		MarketId:  (*offer).MarketId,
		AskId:     (*offer).AskId,
		BidId:     (*offer).BidId,
		Price:     (*offer).StrikePrice,
		Volume:    (*offer).Volume,
		Funds:     (*offer).Funds,
		AskUserId: ask.UserId,
		BidUserId: bid.UserId,
	}
	result := mainDB.Create(&trade)
	if result.Error != nil {
		mainDB.DbRollback()
		if times > 0 {
			trade.Id = 0
			tryCreateTrade(offer, times-1)
			return
		}
	}

	errAsk := ask.Strike(mainDB, trade)
	errBid := bid.Strike(mainDB, trade)
	if errAsk == nil && errBid == nil {
		mainDB.DbCommit()
		pushMessageToRefreshTicker((*offer).MarketId)
		pushMessageToRefreshKLine((*offer).MarketId)
		return
	}
	mainDB.DbRollback()
	if times > 0 {
		trade.Id = 0
		tryCreateTrade(offer, times-1)
	}
	return
}

func pushMessageToRefreshTicker(marketId int) {
	var payload struct {
		MarketId int `json:"market_id"`
	}
	payload.MarketId = marketId
	b, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("error:", err)
	}

	var exchange, routingKey string
	for _, w := range config.AllWorkers {
		if w.Name == "TickerWorker" {
			exchange = w.Exchange
			routingKey = w.RoutingKey
		}
	}
	err = config.RabbitMqConnect.PublishMessageWithRouteKey(exchange, routingKey, "text/plain", false, false, &b, amqp.Table{}, amqp.Persistent, "")
	if err != nil {
		fmt.Println("{ error:", err, "}")
		panic(err)
	}
	return
}

func pushMessageToRefreshKLine(marketId int) {
	var payload struct {
		MarketId  int   `json:"market_id"`
		Period    int64 `json:"period"`
		Timestamp int64 `json:"timestamp"`
	}
	payload.MarketId = marketId
	payload.Timestamp = time.Now().Unix()
	var exchange, routingKey string
	for _, w := range config.AllWorkers {
		if w.Name == "KLineWorker" {
			exchange = w.Exchange
			routingKey = w.RoutingKey
		}
	}
	for _, period := range []int64{1, 5, 15, 30, 60, 120, 240, 360, 720, 1440, 4320, 10080} {
		payload.Period = period
		b, err := json.Marshal(payload)
		if err != nil {
			fmt.Println("error:", err)
		}
		err = config.RabbitMqConnect.PublishMessageWithRouteKey(exchange, routingKey, "text/plain", false, false, &b, amqp.Table{}, amqp.Persistent, "")
		if err != nil {
			fmt.Println("{ error:", err, "}")
			return
		}
	}
	return
}
