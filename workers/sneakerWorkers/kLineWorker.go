package sneakerWorkers

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
	"github.com/shopspring/decimal"
)

func (worker *Worker) KLineWorker(payloadJson *[]byte) (queueName string, message []byte) {
	var payload struct {
		MarketId  int   `json:"market_id"`
		Period    int64 `json:"period"`
		Timestamp int64 `json:"timestamp"`
	}
	json.Unmarshal([]byte(*payloadJson), &payload)
	timestamp := payload.Timestamp
	previous := timestamp/(payload.Period*60)*(payload.Period*60) - payload.Period*60
	begin := timestamp / (payload.Period * 60) * (payload.Period * 60)
	end := timestamp/(payload.Period*60)*(payload.Period*60) + payload.Period*60
	createPoint(payload.MarketId, payload.Period, time.Unix(previous, 0), time.Unix(begin, 0))
	createPoint(payload.MarketId, payload.Period, time.Unix(begin, 0), time.Unix(end, 0))
	return
}

func createPoint(marketId int, period int64, begin, end time.Time) {
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	var k KLine
	if mainDB.Where("market_id = ?", marketId).Where("period = ?", period).Where("timestamp = ?", begin).First(&k).RecordNotFound() {
		k.MarketId = marketId
		k.Period = int(period)
		k.Timestamp = begin.Unix()
	}
	mainDB.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, end).Select("first(price) as open").Scan(&k)
	if k.Open.Equal(decimal.Zero) {
		mainDB.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("created_at < ?", begin).Select("last(price) as open").Scan(&k)
		var latestK KLine
		mainDB.Order("id DESC").Where("market_id = ?", marketId).Where("period = ?", period).First(&latestK)
		k.High = latestK.High
		k.Low = latestK.Low
		k.Close = latestK.Close
		k.Volume = decimal.Zero
	} else {
		mainDB.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, end).Select("max(price) as high").Scan(&k)
		mainDB.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, end).Select("min(price) as low").Scan(&k)
		mainDB.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, end).Select("last(price) as close").Scan(&k)
		mainDB.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, end).Select("sum(volume) as volume").Scan(&k)
	}
	mainDB.Save(&k)
	mainDB.DbCommit()

	kRedis := utils.GetRedisConn("k")
	defer kRedis.Close()
	b, _ := json.Marshal(k.Data())
	kRedis.Send("ZREMRANGEBYSCORE", k.RedisKey(), k.Timestamp)
	kRedis.Send("ZADD", k.RedisKey(), k.Timestamp, string(b))
	if _, err := kRedis.Do(""); err != nil {
		fmt.Println(err)
		return
	}
}
