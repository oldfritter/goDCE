package v1

import (
	"encoding/json"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func V1GetTickers(context echo.Context) error {
	tickerRedis := utils.GetRedisConn("ticker")
	defer tickerRedis.Close()
	values, _ := redis.Values(tickerRedis.Do("HGETALL", TickersRedisKey))
	var tickers []interface{}
	for i, value := range values {
		if i%2 == 1 {
			ticker := Ticker{}
			json.Unmarshal(value.([]byte), &ticker)
			tickers = append(tickers, ticker)
		}
	}
	response := utils.SuccessResponse
	response.Body = tickers
	return context.JSON(http.StatusOK, response)
}

func V1GetTickersMarket(context echo.Context) error {
	var market Market
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	if mainDB.Where("code = ?", context.Param("market")).First(&market).RecordNotFound() {
		return utils.BuildError("1021")
	}
	tickerRedis := utils.GetRedisConn("ticker")
	defer tickerRedis.Close()
	value, _ := tickerRedis.Do("HGET", TickersRedisKey, market.Id)
	ticker := Ticker{}
	json.Unmarshal(value.([]byte), &ticker)
	response := utils.SuccessResponse
	response.Body = ticker
	return context.JSON(http.StatusOK, response)
}
