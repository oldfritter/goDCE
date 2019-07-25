package models

import (
	"github.com/shopspring/decimal"
)

type AccountVersionCheckPoint struct {
	CommonModel
	AccountVersionId int             `json:"account_version_id"`
	UserId           int             `json:"user_id"`
	AccountId        int             `json:"account_id"`
	Fixed            string          `json:"fixed"`
	FixedNum         decimal.Decimal `json:"fixed_num" gorm:"type:decimal(32,16)"`
	Balance          decimal.Decimal `json:"balance" gorm:"type:decimal(32,16)"`
}
