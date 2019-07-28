package models

import (
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type Fee struct {
	CurrencyId   int             `json:"currency_id"`
	CurrencyCode string          `json:"currency_code" gorm:"type:varchar(16)"`
	Amount       decimal.Decimal `json:"amount"`
}

type Trade struct {
	CommonModel
	Trend    int             `json:"trend"`
	MarketId int             `json:"market_id"`
	AskId    int             `json:"ask_id"`
	BidId    int             `json:"bid_id"`
	Price    decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"price"`
	Volume   decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"volume"`
	Funds    decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"funds"`

	Side      string `json:"side" sql:"-" gorm:"type:varchar(4)"`
	Market    Market `json:"market" sql:"-"`
	AskUserId int    `json:"ask_user_id"`
	BidUserId int    `json:"bid_user_id"`
	AskFee    Fee    `json:"ask_fee" sql:"-"`
	BidFee    Fee    `json:"bid_fee" sql:"-"`
}
type Attrs struct {
	Tid    int             `json:"tid"`
	Amount decimal.Decimal `json:"amount"`
	Price  decimal.Decimal `json:"price"`
	Date   int64           `json:"date"`
}

func (trade *Trade) AfterFind(db *gorm.DB) {
}

func (trade *Trade) SimpleAttrs() (attrs Attrs) {
	attrs.Tid = trade.Id
	attrs.Amount = trade.Volume
	attrs.Price = trade.Price
	attrs.Date = trade.CreatedAt.Unix()
	return
}
