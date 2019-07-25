package models

import (
	"fmt"

	"github.com/oldfritter/goDCE/utils"
	"github.com/shopspring/decimal"
)

type OrderCurrency struct {
	Fee        decimal.Decimal `json:"fee"`
	Currency   string          `json:"currency"`
	CurrencyId int             `json:"currency_id"`
	Fixed      int             `json:"fixed"`
}

type Market struct {
	Id              int
	Name            string `gorm:"type:varchar(16)"`
	Code            string `gorm:"type:varchar(16)"`
	PriceGroupFixed int
	SortOrder       int
	AskCurrencyId   int             `json:"ask_currency_id"`
	BidCurrencyId   int             `json:"bid_currency_id"`
	AskFee          decimal.Decimal `json:"ask_fee" gorm:"type:decimal(32,16);default:null;"`
	BidFee          decimal.Decimal `json:"bid_fee" gorm:"type:decimal(32,16);default:null;"`
	AskFixed        int             `json:"ask_fixed"`
	BidFixed        int             `json:"bid_fixed"`
	Visible         bool
	Tradable        bool
	// Logo             string
	// CoinmarketcapUrl string
	// BaseUnit         string
	// QuoteUnit        string

	// 撮合相关属性
	Ack             bool
	Durable         bool
	MatchingAble    bool
	MatchingNode    string `gorm:"default:'a'; type:varchar(11)"`
	TradeTreatNode  string `gorm:"default:'a'; type:varchar(11)"`
	OrderCancelNode string `gorm:"default:'a'; type:varchar(11)"`
	Running         bool   `sql:"-"`
}

// Exchange
func (assignment *Market) MatchingExchange() string {
	return utils.AmqpGlobalConfig.Exchange.Matching["key"]
}
func (assignment *Market) TradeTreatExchange() string {
	return utils.AmqpGlobalConfig.Exchange.Trade["key"]
}
func (assignment *Market) OrderCancelExchange() string {
	return utils.AmqpGlobalConfig.Exchange.Cancel["key"]
}

// Queue
func (assignment *Market) MatchingQueue() string {
	return assignment.MatchingExchange() + "." + assignment.Code
}
func (assignment *Market) TradeTreatQueue() string {
	return assignment.TradeTreatExchange() + "." + assignment.Code
}
func (assignment *Market) OrderCancelQueue() string {
	return assignment.OrderCancelExchange() + "." + assignment.Code
}

// reload Queue
func (assignment *Market) OrderCancelReloadQueue() string {
	return utils.AmqpGlobalConfig.Queue.Cancel["reload"]
}

// LogFilePath
func (assignment *Market) MatchingLogFilePath() string {
	return "logs/matching-" + assignment.Code + ".log"
}
func (assignment *Market) TradeTreatLogFilePath() string {
	return "logs/trade-" + assignment.Code + ".log"
}
func (assignment *Market) OrderCancelLogFilePath() string {
	return "logs/order-cancel-" + assignment.Code + ".log"
}

func (market *Market) LatestTradesRedisKey() string {
	return fmt.Sprintf("goHex:latestTrades:%v", market.Code)
}
func (market *Market) TickerRedisKey() string {
	return "goHex:ticker:" + market.Code
}
func (market *Market) KLineRedisKey(period string) string {
	return fmt.Sprintf("goHex:k:%v:%v", market.Id, period)
}

func (market *Market) AskRedisKey() string {
	return fmt.Sprintf("goHex:depth:%v:ask", market.Id)
}
func (market *Market) BidRedisKey() string {
	return fmt.Sprintf("goHex:depth:%v:bid", market.Id)
}
