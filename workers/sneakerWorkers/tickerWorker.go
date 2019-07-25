package sneakerWorkers

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func (worker *Worker) TickerWorker(payloadJson *[]byte) (queueName string, message []byte) {
	var payload struct {
		MarketId int `json:"market_id"`
	}
	json.Unmarshal([]byte(*payloadJson), &payload)
	buildTicker(payload.MarketId)
	return
}

func buildTicker(marketId int) {
	db := utils.MainDbBegin()
	defer db.DbRollback()
	now := time.Now()
	begin := now.Add(-time.Hour * 24)
	ticker := Ticker{MarketId: marketId}
	db.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, now).Select("min(price) as low").Scan(&ticker.TickerAspect)
	db.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, now).Select("max(price) as high").Scan(&ticker.TickerAspect)
	db.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, now).Select("last(price) as last").Scan(&ticker.TickerAspect)
	db.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, now).Select("sum(volume) as volume").Scan(&ticker.TickerAspect)
	db.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, now).Select("first(price) as open").Scan(&ticker.TickerAspect)

	db.Model(Order{}).Where("state = ?", 100).Where("type = ?", "OrderBid").Where("market_id = ?", marketId).Select("max(price) as buy").Scan(&ticker.TickerAspect)
	db.Model(Order{}).Where("state = ?", 100).Where("type = ?", "OrderAsk").Where("market_id = ?", marketId).Select("min(price) as sell").Scan(&ticker.TickerAspect)

	tickerRedis := utils.GetRedisConn("ticker")
	defer tickerRedis.Close()
	key := ticker.RedisKey()
	tickerRedis.Send("HSET", key, "low", ticker.TickerAspect.Low)
	tickerRedis.Send("HSET", key, "high", ticker.TickerAspect.High)
	tickerRedis.Send("HSET", key, "last", ticker.TickerAspect.Last)
	tickerRedis.Send("HSET", key, "volume", ticker.TickerAspect.Volume)
	tickerRedis.Send("HSET", key, "open", ticker.TickerAspect.Open)
	tickerRedis.Send("HSET", key, "buy", ticker.TickerAspect.Buy)
	tickerRedis.Send("HSET", key, "sell", ticker.TickerAspect.Sell)

	if _, err := tickerRedis.Do(""); err != nil {
		fmt.Println(err)
		return
	}

}
