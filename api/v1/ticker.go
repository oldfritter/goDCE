package v1

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func V1GetTickersMarket(context echo.Context) error {
	market, err := FindMarketByCode(context.Param("market"))
	if err != nil {
		return utils.BuildError("1021")
	}
	ticker := Ticker{MarketId: market.Id, TickerAspect: market.Ticker}
	response := utils.SuccessResponse
	response.Body = ticker
	return context.JSON(http.StatusOK, response)
}

func V1GetTickers(context echo.Context) error {
	var tickers []Ticker
	for _, market := range AllMarkets {
		tickers = append(tickers, Ticker{
			MarketId:     market.Id,
			At:           time.Now().Unix(),
			Name:         market.Name,
			TickerAspect: market.Ticker,
		})
	}

	response := utils.SuccessResponse
	response.Body = tickers
	return context.JSON(http.StatusOK, response)
}
