package sneakerWorkers

import (
	"encoding/json"
	"fmt"

	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func (worker Worker) RebuildKLineToRedisWorker(payloadJson *[]byte) (queueName string, message []byte) {
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
	mainDB.DbRollback()
	kRedis := utils.GetRedisConn("k")
	defer kRedis.Close()
	for i, k := range ks {
		b, _ := json.Marshal(k.Data())
		kRedis.Send("ZREMRANGEBYSCORE", k.RedisKey(), k.Timestamp)
		kRedis.Send("ZADD", k.RedisKey(), k.Timestamp, string(b))
		if i%10 == 9 {
			if _, err := kRedis.Do(""); err != nil {
				fmt.Println(err)
			}
		}
	}
	if _, err := kRedis.Do(""); err != nil {
		fmt.Println(err)
	}
	return
}
