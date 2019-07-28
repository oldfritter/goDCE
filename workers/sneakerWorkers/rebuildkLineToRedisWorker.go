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

func (worker *Worker) RebuildKLineToRedisWorker(payloadJson *[]byte) (queueName string, message []byte) {
	var payload struct {
		MarketId int `json:"market_id"`
		Period   int `json:"period"`
	}
	json.Unmarshal([]byte(*payloadJson), &payload)
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	var ks []KLine
	if mainDB.Where("market_id = ?", payload.MarketId).Where("period = ?", payload.Period).Find(&ks).RecordNotFound() {
		return
	}

	kRedis := utils.GetRedisConn("k")
	defer kRedis.Close()
	for _, kLine := range ks {
		kRedis.Do("RPUSH", k.RedisKey()+":rebuild", kLine.Data())
	}
	kRedis.Do("RENAME", k.RedisKey()+":rebuild", k.RedisKey())

	return
}
