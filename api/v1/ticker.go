package v1

import (
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func V1GetTickers(context echo.Context) error {
	var markets []Market
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	mainDB.Where("visible is true").Find(&markets)

	tickerRedis := utils.GetRedisConn("ticker")
	defer tickerRedis.Close()

	var tickers []Ticker
	for _, market := range markets {
		if !market.Visible {
			continue
		}
		var tickerStruct Ticker
		buildTickerWithMarket(&tickerStruct, &market, tickerRedis)
		tickers = append(tickers, tickerStruct)
	}
	response := utils.SuccessResponse
	response.Body = tickers
	return context.JSON(http.StatusOK, response)
}

func V1GetTickersMarket(context echo.Context) error {
	tickerRedis := utils.GetRedisConn("ticker")
	defer tickerRedis.Close()

	var market Market
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()

	if mainDB.Where("name = ?", context.QueryParam("market")).First(&market).RecordNotFound() {
		return utils.BuildError("1021")
	}
	var tickerStruct Ticker
	buildTickerWithMarket(&tickerStruct, &market, tickerRedis)
	response := utils.SuccessResponse
	response.Body = tickerStruct
	return context.JSON(http.StatusOK, response)
}

func buildTickerWithMarket(tickerStruct *Ticker, market *Market, tickerRedis redis.Conn) {
	var ticker TickerAspect
	result, _ := redis.Values(tickerRedis.Do("HGETALL", (*market).TickerRedisKey()))
	ticker.Add(result)
	tickerStruct.TickerAspect = ticker
}
