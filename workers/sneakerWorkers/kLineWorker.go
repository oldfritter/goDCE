package sneakerWorkers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
	"github.com/shopspring/decimal"
)

func (worker *Worker) KLineWorker(payloadJson *[]byte) (queueName string, message []byte) {
	start := time.Now().UnixNano()
	var payload struct {
		MarketId   int    `json:"market_id"`
		Timestamp  int64  `json:"timestamp"`
		Period     int64  `json:"period"`
		DataSource string `json:"data_source"`
	}
	json.Unmarshal([]byte(*payloadJson), &payload)
	if payload.Period == 0 {
		fmt.Println("INFO--KLineWorker payload: ", payload)
		return
	}
	timestamp := payload.Timestamp
	begin := timestamp / (payload.Period * 60) * (payload.Period * 60)
	end := timestamp/(payload.Period*60)*(payload.Period*60) + payload.Period*60
	createPoint(payload.MarketId, payload.Period, begin, end, payload.DataSource)
	fmt.Println("INFO--KLineWorker payload: ", payload, ", time:", (time.Now().UnixNano()-start)/1000000, " ms")
	return
}

func createPoint(marketId int, period, begin, end int64, dataSource string) {
	if dataSource == "redis" {
		synFromRedis(marketId, period, begin, end)
	} else {
		if period == 1 {
			calculateInDB(marketId, period, begin, end)
		} else {
			calculateInBackupDB(marketId, period, begin, end)
		}
	}
}

func calculateInDB(marketId int, period, begin, end int64) (k KLine) {
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()

	backupDB := utils.BackupDbBegin()
	defer backupDB.DbRollback()
	if backupDB.Where("market_id = ?", marketId).Where("period = ?", period).Where("timestamp = ?", begin).First(&k).RecordNotFound() {
		k.MarketId = marketId
		k.Period = int(period)
		k.Timestamp = begin
	}

	mainDB.Model(Trade{}).Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", time.Unix(begin, 0), time.Unix(end, 0)).Limit(1).Select("price as open").Scan(&k)
	if k.Open.Equal(decimal.Zero) {
		mainDB.Model(Trade{}).Order("created_at DESC").Where("market_id = ?", marketId).Where("created_at < ?", time.Unix(begin, 0)).Limit(1).Select("price as open").Scan(&k)
		var latestK KLine
		backupDB.Order("timestamp DESC").Where("market_id = ?", marketId).Where("period = ?", period).First(&latestK)
		k.High = latestK.High
		k.Low = latestK.Low
		k.Close = latestK.Close
		k.Volume = decimal.Zero
	} else {
		mainDB.Model(Trade{}).Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", time.Unix(begin, 0), time.Unix(end, 0)).Select("max(price) as high").Scan(&k)
		mainDB.Model(Trade{}).Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", time.Unix(begin, 0), time.Unix(end, 0)).Select("min(price) as low").Scan(&k)
		mainDB.Model(Trade{}).Order("created_at DESC").Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", time.Unix(begin, 0), time.Unix(end, 0)).Limit(1).Select("price as close").Scan(&k)
		mainDB.Model(Trade{}).Where("market_id = ?", marketId).Where("? <= created_at AND created_at < ?", time.Unix(begin, 0), time.Unix(end, 0)).Select("sum(volume) as volume").Scan(&k)
	}
	backupDB.Save(&k)
	backupDB.DbCommit()
	synToRedis(&k)
	return
}

func calculateInBackupDB(marketId int, period, begin, end int64) (k KLine) {
	backupDB := utils.BackupDbBegin()
	defer backupDB.DbRollback()
	if backupDB.Where("market_id = ?", marketId).Where("period = ?", period).Where("timestamp = ?", begin).First(&k).RecordNotFound() {
		k.MarketId = marketId
		k.Period = int(period)
		k.Timestamp = begin
	}
	periods := []int64{1, 5, 15, 30, 60, 120, 240, 360, 720, 1440, 4320, 10080}
	lastPeriod := periods[0]
	for _, p := range periods {
		lastPeriod = p
		if p >= period || lastPeriod == 120 {
			break
		}
	}
	backupDB.Model(KLine{}).Where("market_id = ?", marketId).Where("period = ?", lastPeriod).Where("? <= timestamp AND timestamp < ?", begin, end).Limit(1).Select("open as open").Scan(&k)
	if k.Open.Equal(decimal.Zero) {
		backupDB.Model(KLine{}).Where("market_id = ?", marketId).Where("timestamp < ?", begin).Limit(1).Select("close as open").Scan(&k)
	}
	backupDB.Model(KLine{}).Where("market_id = ?", marketId).Where("period = ?", lastPeriod).Where("? <= timestamp AND timestamp < ?", begin, end).Select("max(high) as high").Scan(&k)
	backupDB.Model(KLine{}).Where("market_id = ?", marketId).Where("period = ?", lastPeriod).Where("? <= timestamp AND timestamp < ?", begin, end).Select("min(low) as low").Scan(&k)
	backupDB.Model(KLine{}).Where("market_id = ?", marketId).Where("period = ?", lastPeriod).Where("? <= timestamp AND timestamp < ?", begin, end).Select("sum(volume) as volume").Scan(&k)
	backupDB.Model(KLine{}).Where("market_id = ?", marketId).Where("period = ?", lastPeriod).Where("? <= timestamp AND timestamp < ?", begin, end).Order("timestamp DESC").Limit(1).Select("close as close").Scan(&k)
	backupDB.Save(&k)
	backupDB.DbCommit()
	synToRedis(&k)
	return
}

func synFromRedis(marketId int, period, begin, end int64) (k KLine) {
	market, _ := FindMarketById(marketId)
	backupDB := utils.BackupDbBegin()
	defer backupDB.DbRollback()
	if backupDB.Where("market_id = ?", marketId).Where("period = ?", period).Where("timestamp = ?", begin).First(&k).RecordNotFound() {
		k.MarketId = marketId
		k.Period = int(period)
		k.Timestamp = begin
	}
	kRedis := utils.GetRedisConn("k")
	defer kRedis.Close()
	key := fmt.Sprintf("peatio:%v:k:%v", market.Id, period)
	value, _ := redis.String(kRedis.Do("LINDEX", key, 0))
	var item [6]decimal.Decimal
	json.Unmarshal([]byte(value), &item)
	offset := (begin - item[0].IntPart()) / 60 / int64(period)
	if offset < 0 {
		offset = 0
	}
	values, err := redis.Values(kRedis.Do("LRANGE", key, offset, offset+2-1))
	if err != nil {
		fmt.Println("lrange err", err.Error())
		return
	}
	for _, v := range values {
		json.Unmarshal(v.([]byte), &item)
		if item[0].IntPart() == begin {
			k.Timestamp = begin
			k.Open = item[1]
			k.High = item[2]
			k.Low = item[3]
			k.Close = item[4]
			k.Volume = item[5]
		}
	}
	if k.Open.IsZero() && k.High.IsZero() && k.Low.IsZero() && k.Close.IsZero() && k.Volume.IsZero() {
		backupDB.DbRollback()
	} else {
		backupDB.Save(&k)
		backupDB.DbCommit()
	}
	return
}

func synToRedis(k *KLine) {
	kRedis := utils.GetRedisConn("k")
	defer kRedis.Close()

	b, _ := json.Marshal((*k).Data())
	kRedis.Send("ZREMRANGEBYSCORE", (*k).RedisKey(), (*k).Timestamp)
	kRedis.Do("ZADD", k.RedisKey(), (*k).Timestamp, string(b))

}
