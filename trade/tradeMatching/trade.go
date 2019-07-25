package tradeMatching

import (
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func buildLatestTrades(market *Market) {
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()

	var trades []Trade
	var results []Attrs
	mainDB.Where("market_id = ?", (*market).Id).Order("id DESC").Limit(80).Find(&trades)
	for _, trade := range trades {
		results = append(results, trade.SimpleAttrs())
	}

	dataRedis := utils.GetRedisConn("data")
	defer dataRedis.Close()

	dataRedis.Do("SET", (*market).LatestTradesRedisKey(), results)
}
