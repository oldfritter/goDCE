package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func V1GetTickers(context echo.Context) error {
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	tickerRedis := utils.GetRedisConn("ticker")
	defer tickerRedis.Close()
	values, _ := redis.Values(tickerRedis.Do("HGETALL", TickersRedisKey))
	tickers := make([]interface{}, 0)
	var market Market
	for i, value := range values {
		if i%2 == 1 {
			ticker := Ticker{
				MarketId: market.Id,
				At:       time.Now().Unix(),
				Name:     market.Name,
			}
			json.Unmarshal(value.([]byte), &ticker.TickerAspect)
			tickers = append(tickers, ticker)
		} else {
			marketId, _ := redis.String(value, nil)
			if mainDB.Where("id = ?", marketId).First(&market).RecordNotFound() {
				return utils.BuildError("1021")
			}
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
	ticker := Ticker{
		MarketId: market.Id,
		At:       time.Now().Unix(),
		Name:     market.Name,
	}
	json.Unmarshal(value.([]byte), &ticker.TickerAspect)
	response := utils.SuccessResponse
	response.Body = ticker
	return context.JSON(http.StatusOK, response)
}
