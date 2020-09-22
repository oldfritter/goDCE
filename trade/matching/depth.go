package matching

import (
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
	"github.com/shopspring/decimal"
)

type Depth struct {
	MarketId  int                        `json:"market_id"`
	AskOrders map[string]decimal.Decimal `json:"asks"`
	BidOrders map[string]decimal.Decimal `json:"bids"`
}

func InitializeDepth(marketId int) (depth Depth) {
	depth.MarketId = marketId
	return
}

func buildDepth(market *Market) {
	engine := Engines[(*market).Id]
	depth := InitializeDepth((*market).Id)

	askOrderBook := engine.AskOrderBook()
	askDepth := askOrderBook.LimitOrdersMap()
	for price, orders := range askDepth {
		var volume decimal.Decimal
		for _, order := range orders {
			volume = volume.Add(order.Volume)
		}
		depth.AskOrders[price] = volume
	}
	bidOrderBook := engine.BidOrderBook()
	bidDepth := bidOrderBook.LimitOrdersMap()
	for price, orders := range bidDepth {
		var volume decimal.Decimal
		for _, order := range orders {
			volume = volume.Add(order.Volume)
		}
		depth.BidOrders[price] = volume
	}

	dataRedis := utils.GetRedisConn("data")
	defer dataRedis.Close()

	dataRedis.Do("SET", (*market).AskRedisKey(), depth.AskOrders)
	dataRedis.Do("SET", (*market).BidRedisKey(), depth.BidOrders)
}
