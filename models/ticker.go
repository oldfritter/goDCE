package models

import (
	"fmt"
	"github.com/shopspring/decimal"
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
	MarketId        int          `json:"market_id"`
	Name            string       `json:"name"`
	Code            string       `json:"code"`
	Quote           string       `json:"quote"`
	PriceGroupFixed string       `json:"price_group_fixed"`
	TickerAspect    TickerAspect `json:"ticker"`
}

func (ticker *TickerAspect) Add(attrs []interface{}) {
	colume := ""
	for i, v := range attrs {
		if (i % 2) == 0 {
			colume = fmt.Sprintf("%s", v)
		} else {
			if colume == "buy" {
				ticker.Buy, _ = decimal.NewFromString(fmt.Sprintf("%s", v))
			} else if colume == "sell" {
				ticker.Sell, _ = decimal.NewFromString(fmt.Sprintf("%s", v))
			} else if colume == "low" {
				ticker.Low, _ = decimal.NewFromString(fmt.Sprintf("%s", v))
			} else if colume == "high" {
				ticker.High, _ = decimal.NewFromString(fmt.Sprintf("%s", v))
			} else if colume == "last" {
				ticker.Last, _ = decimal.NewFromString(fmt.Sprintf("%s", v))
			} else if colume == "open" {
				ticker.Open, _ = decimal.NewFromString(fmt.Sprintf("%s", v))
			} else if colume == "volume" {
				ticker.Volume, _ = decimal.NewFromString(fmt.Sprintf("%s", v))
			}
		}
	}

}

func (ticker *Ticker) RedisKey() string {
	return fmt.Sprintf("goHex:ticker:%v", ticker.Code)
}
