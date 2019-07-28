package sneakerWorkers

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
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
	last, err := redis.String(kRedis.Do("LINDEX", k.RedisKey(), -1))
	if err != nil {
		kRedis.Do("RPUSH", k.RedisKey(), k.Data())
	} else {
		var item [6]decimal.Decimal
		json.Unmarshal([]byte(last), &item)
		if item[0].String() == strconv.FormatInt((begin.Unix()-period*60), 10) {
			kRedis.Do("RPUSH", k.RedisKey(), k.Data())
		} else if item[0].String() == strconv.FormatInt((begin.Unix()), 10) {
			kRedis.Do("RPOP", k.RedisKey())
			kRedis.Do("RPUSH", k.RedisKey(), k.Data())
		}
	}
}
