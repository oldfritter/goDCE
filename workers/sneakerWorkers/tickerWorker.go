package sneakerWorkers

import (
	"encoding/json"
	"fmt"
	"time"

	sneaker "github.com/oldfritter/sneaker-go/v3"

	"github.com/oldfritter/goDCE/config"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
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
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	var market Market
	if mainDB.Where("id = ?", marketId).First(&market).RecordNotFound() {
		return
	}

	now := time.Now()
	begin := now.Add(-time.Hour * 24)
	ticker := Ticker{MarketId: marketId, Name: market.Name}
	mainDB.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, now).Select("min(price) as low").Scan(&ticker.TickerAspect)
	mainDB.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, now).Select("max(price) as high").Scan(&ticker.TickerAspect)
	mainDB.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, now).Select("last(price) as last").Scan(&ticker.TickerAspect)
	mainDB.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, now).Select("sum(volume) as volume").Scan(&ticker.TickerAspect)
	mainDB.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, now).Select("first(price) as open").Scan(&ticker.TickerAspect)
	mainDB.Model(Order{}).Where("state = ?", 100).Where("type = ?", "OrderBid").Where("market_id = ?", marketId).Select("max(price) as buy").Scan(&ticker.TickerAspect)
	mainDB.Model(Order{}).Where("state = ?", 100).Where("type = ?", "OrderAsk").Where("market_id = ?", marketId).Select("min(price) as sell").Scan(&ticker.TickerAspect)

	tickerRedis := utils.GetRedisConn("ticker")
	defer tickerRedis.Close()
	b, _ := json.Marshal(ticker)
	if _, err := tickerRedis.Do("HSET", TickersRedisKey, marketId, string(b)); err != nil {
		fmt.Println(err)
		return
	}
}
