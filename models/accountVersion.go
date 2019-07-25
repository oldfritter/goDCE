package models

import (
	"github.com/shopspring/decimal"
)

type AccountVersion struct {
	CommonModel
	UserId         int             `json:"user_id"`
	AccountId      int             `json:"account_id"`
	Reason         int             `json:"reason"`
	Balance        decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"balance"`
	Locked         decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"locked"`
	Fee            decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"fee"`
	Amount         decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"amount"`
	ModifiableId   int             `json:"modifiable_id"`
	ModifiableType string          `json:"modifiable_type"`
	CurrencyId     int             `json:"currency_id"`
	Fun            int             `json:"fun"`
}
