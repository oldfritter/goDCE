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

	var Tickers []Ticker
	for _, market := range markets {
		if !market.Visible {
			continue
		}
		var tickerStruct Ticker
		buildTickerWithMarket(&tickerStruct, &market, tickerRedis)
		Tickers = append(Tickers, tickerStruct)
	}
	return context.JSON(http.StatusOK, Tickers)
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
	return context.JSON(http.StatusOK, tickerStruct)
}

func buildTickerWithMarket(tickerStruct *Ticker, market *Market, tickerRedis redis.Conn) {
	var ticker TickerAspect
	result, _ := redis.Values(tickerRedis.Do("HGETALL", (*market).TickerRedisKey()))
	ticker.Add(result)
	tickerStruct.TickerAspect = ticker
}
