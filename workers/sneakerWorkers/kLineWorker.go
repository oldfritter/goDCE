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
		Period    int   `json:"period"`
		Timestamp int64 `json:"timestamp"`
	}
	json.Unmarshal([]byte(*payloadJson), &payload)
	periods := []int64{1, 5, 15, 30, 60, 120, 240, 360, 720, 1440, 4320, 10080}
	timestamp := payload.Timestamp
	for _, period := range periods {
		previous := timestamp/(period*60)*(period*60) - period*60
		begin := timestamp / (period * 60) * (period * 60)
		end := timestamp/(period*60)*(period*60) + period*60
		createPoint(payload.MarketId, period, time.Unix(previous, 0), time.Unix(begin, 0))
		createPoint(payload.MarketId, period, time.Unix(begin, 0), time.Unix(end, 0))
	}
	return
}

func createPoint(marketId int, period int64, begin, end time.Time) {
	db := utils.MainDbBegin()
	defer db.DbRollback()
	var k KLine
	if db.Where("market_id = ?", marketId).Where("period = ?", period).Where("timestamp = ?", begin).First(&k).RecordNotFound() {
		k.MarketId = marketId
		k.Period = int(period)
		k.Timestamp = begin.Unix()
	}

	db.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, end).Select("first(price) as open").Scan(&k)
	if k.Open.Equal(decimal.Zero) {
		db.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("created_at < ?", begin).Select("last(price) as open").Scan(&k)
		var latestK KLine
		db.Order("id DESC").Where("market_id = ?", marketId).Where("period = ?", period).First(&latestK)
		k.High = latestK.High
		k.Low = latestK.Low
		k.Close = latestK.Close
		k.Volume = decimal.Zero
	} else {
		db.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, end).Select("max(price) as high").Scan(&k)
		db.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, end).Select("min(price) as low").Scan(&k)
		db.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, end).Select("last(price) as close").Scan(&k)
		db.Model(Trade{}).Order("id ASC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", begin, end).Select("sum(volume) as volume").Scan(&k)
	}

	db.Save(&k)
	db.DbCommit()

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
