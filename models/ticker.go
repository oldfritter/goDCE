package models

import (
	"github.com/shopspring/decimal"
)

var (
	TickersRedisKey = "goDCE:tickers"
)

type TickerAspect struct {
	Buy    decimal.Decimal `json:"buy"`
	Sell   decimal.Decimal `json:"sell"`
	Low    decimal.Decimal `json:"low"`
	High   decimal.Decimal `json:"high"`
	Last   decimal.Decimal `json:"last"`
	Open   decimal.Decimal `json:"open"`
	Volume decimal.Decimal `json:"volume"`
}

type Ticker struct {
	MarketId     int          `json:"market_id"`
	Name         string       `json:"name"`
	Code         string       `json:"code"`
	TickerAspect TickerAspect `json:"ticker"`
}
