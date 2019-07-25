package models

import (
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
)

type KLine struct {
	CommonModel
	MarketId int `json:"market_id"`
	Period   int `json:"period"`

	Timestamp int64           `json:"timestamp"`                                       // 时间戳
	Open      decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"open"`   // 开盘价
	High      decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"high"`   // 最高价
	Low       decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"low"`    // 最低价
	Close     decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"close"`  // 收盘价
	Volume    decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"volume"` // 交易量
}

func (k *KLine) Data() (data [5]string) {
	data[0] = strconv.FormatInt(k.Timestamp, 10)
	data[1] = k.Open.String()
	data[2] = k.High.String()
	data[3] = k.Low.String()
	data[4] = k.Close.String()
	return
}

func (k *KLine) RedisKey() string {
	return fmt.Sprintln("goDCE:k:%v:%v", k.MarketId, k.Period)
}
