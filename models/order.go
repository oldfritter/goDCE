package models

import (
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/oldfritter/goDCE/utils"
	"github.com/shopspring/decimal"
)

var (
	CANCEL = 0
	WAIT   = 100
	DONE   = 200
)

type Order struct {
	CommonModel
	UserId        int             `json:"user_id"`
	AskAccountId  int             `json:"ask_account_id" gorm:"default:null"`
	BidAccountId  int             `json:"bid_account_id" gorm:"default:null"`
	MarketId      int             `json:"market_id"`
	State         int             `json:"-"`
	Type          string          `gorm:"type:varchar(16)" json:"type"`
	Sn            string          `gorm:"type:varchar(16)" json:"sn" gorm:"default:null"`
	Source        string          `gorm:"type:varchar(16)" json:"source"`
	OrderType     string          `gorm:"type:varchar(16)" json:"order_type"`
	Price         decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"price"`
	Volume        decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"volume"`
	OriginVolume  decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"origin_volume"`
	Locked        decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"locked"`
	OriginLocked  decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"origin_locked"`
	FundsReceived decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"funds_received"`
	TradesCount   int             `json:"trades_count"`

	holdAccount   Account         `sql:"-" json:"-"`
	expectAccount Account         `sql:"-" json:"-"`
	Market        Market          `sql:"-" json:"-"`
	StateStr      string          `sql:"-" json:"state"`
	AvgPrice      decimal.Decimal `sql:"-" json:"avg_price"`
}

type OrderJson struct {
	Id        int             `json:"id"`
	MarketId  int             `json:"market_id"`
	Type      string          `json:"type"`
	OrderType string          `json:"order_type"`
	Volume    decimal.Decimal `json:"volume"`
	Price     decimal.Decimal `json:"price"`
	Locked    decimal.Decimal `json:"locked"`
	Timestamp int64           `json:"timestamp"`
}

type MatchingPayload struct {
	Action string    `json:"action"`
	Order  OrderJson `json:"order"`
}

func (mp *MatchingPayload) OrderAttrs() (attrs map[string]string) {
	attrs["id"] = strconv.Itoa(mp.Order.Id)
	attrs["market_id"] = strconv.Itoa(mp.Order.MarketId)
	attrs["type"] = mp.Order.Type
	attrs["order_type"] = mp.Order.OrderType
	attrs["volume"] = mp.Order.Volume.String()
	attrs["price"] = mp.Order.Price.String()
	attrs["locked"] = mp.Order.Locked.String()
	attrs["timestamp"] = strconv.FormatInt(mp.Order.Timestamp, 10)
	return
}

func (order *Order) AfterFind(db *gorm.DB) {
	db.Where("id = ?", order.MarketId).First(&order.Market)
	if order.Type == "OrderAsk" {
		db.Where("id = ?", order.AskAccountId).First(&order.holdAccount)
		db.Where("id = ?", order.BidAccountId).First(&order.expectAccount)
	} else {
		db.Where("id = ?", order.AskAccountId).First(&order.expectAccount)
		db.Where("id = ?", order.BidAccountId).First(&order.holdAccount)
	}
}

func (order *Order) OrderAttrs() (attrs map[string]string) {
	attrs["id"] = strconv.Itoa(order.Id)
	attrs["market_id"] = strconv.Itoa(order.MarketId)
	attrs["type"] = order.Type
	attrs["order_type"] = order.OrderType
	attrs["volume"] = order.Volume.String()
	attrs["price"] = order.Price.String()
	attrs["locked"] = order.Locked.String()
	attrs["timestamp"] = strconv.FormatInt(order.CreatedAt.Unix(), 10)
	return
}

func (order *Order) OType() string {
	if order.Type == "OrderBid" {
		return "bid"
	} else if order.Type == "OrderAsk" {
		return "ask"
	}
	return ""
}

func (order *Order) InitStateStr() {
	switch order.State {
	case 200:
		order.StateStr = "done"
	case 100:
		order.StateStr = "wait"
	case 0:
		order.StateStr = "cancel"
	}
}

func (order *Order) CalculationAvgPrice() {
	funds_used := order.OriginLocked.Sub(order.Locked)
	zero, _ := decimal.NewFromString("0")
	if order.Type == "OrderBid" {
		if order.FundsReceived.Equal(zero) {
			order.AvgPrice = zero
		} else {
			order.AvgPrice = funds_used.Div(order.FundsReceived)
		}
	} else if order.Type == "OrderAsk" {
		if funds_used.Equal(zero) {
			order.AvgPrice = zero
		} else {
			order.AvgPrice = order.FundsReceived.Div(funds_used)
		}
	}
}

func (order *Order) Fee() (fee decimal.Decimal) {
	if order.Type == "OrderAsk" {
		fee = order.Market.BidFee
	} else {
		fee = order.Market.AskFee
	}
	return
}

func (order *Order) getAccountChanges(trade Trade) (a, b decimal.Decimal) {
	if order.Type == "OrderAsk" {
		a = trade.Volume
		b = trade.Funds
	} else {
		b = trade.Volume
		a = trade.Funds
	}
	return
}

func (order *Order) Strike(db *utils.GormDB, trade Trade) (err error) {
	realSub, add := order.getAccountChanges(trade)
	realFee := add.Mul(order.Fee())
	realAdd := add.Sub(realFee)
	err = order.holdAccount.UnlockedAndSubFunds(db, realSub, realSub, decimal.Zero, STRIKE_SUB, trade.Id, "Trade")
	if err != nil {
		return
	}
	err = order.expectAccount.PlusFunds(db, realAdd, realFee, STRIKE_ADD, trade.Id, "Trade")
	if err != nil {
		return
	}

	order.Volume = order.Volume.Sub(trade.Volume)
	order.Locked = order.Locked.Sub(realSub)
	order.FundsReceived = order.FundsReceived.Add(add)
	order.TradesCount += 1

	if order.Volume.Equal(decimal.Zero) {
		order.State = DONE
		//  unlock not used funds
		if order.Locked.GreaterThan(decimal.Zero) {
			err = order.holdAccount.UnlockFunds(db, order.Locked, ORDER_FULLFILLED, trade.Id, "Trade")
			if err != nil {
				return
			}
		}
	} else if order.OrderType == "market" && order.Locked.Equal(decimal.Zero) {
		order.State = CANCEL
	}
	return
}
