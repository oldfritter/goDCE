package models

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/oldfritter/goDCE/utils"
	"github.com/shopspring/decimal"
)

var (
	UNKNOWN                = 0
	FIX                    = 1
	STRIKE_FEE             = 100
	STRIKE_ADD             = 110
	STRIKE_SUB             = 120
	STRIKE_UNLOCK          = 130
	ORDER_SUBMIT           = 600
	ORDER_CANCEL           = 610
	ORDER_FULLFILLED       = 620
	TRANSFER               = 700
	TRANSFER_BACK          = 710
	WALLET_TRANSFER        = 720
	WALLET_TRANSFER_BACK   = 730
	WALLET_TRANSFER_LOCK   = 735
	WALLET_TRANSFER_UNLOCK = 736
	WITHDRAW_LOCK          = 800
	WITHDRAW_UNLOCK        = 810
	DEPOSIT                = 1000
	WITHDRAW               = 2000
	AWARD                  = 3000
	OTC_ORDER_SUBMIT       = 900
	OTC_ORDER_FINISHED     = 910
	OTC_ORDER_FINISHED_2   = 911
	OTC_ORDER_CANCELLED    = 920
	OTC_ORDER_CANCELLED_2  = 921
	OPTION_DEPOSIT         = 1100
	OPTION_UNLOCK          = 1200
	OPTION_CANCEL          = 1500
	PROFIT_LOCK            = 1300
	PROFIT_UNLOCK          = 1400
	RECYCLE_TRANS_FEE      = 1600

	FUNS = map[string]int{
		"UnlockFunds":         1,
		"LockFunds":           2,
		"PlusFunds":           3,
		"SubFunds":            4,
		"UnlockedAndSubFunds": 5,
	}
)

type Account struct {
	CommonModel
	UserId     int             `json:"user_id"`
	CurrencyId int             `json:"currency_id"`
	Balance    decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"balance"`
	Locked     decimal.Decimal `gorm:"type:decimal(32,16);default:null;" json:"locked"`
}

func (account *Account) AfterSave(db *gorm.DB) {}

func (account *Account) Amount() (amount decimal.Decimal) {
	amount = account.Balance.Add(account.Locked)
	return
}

func (account *Account) PlusFunds(db *utils.GormDB, amount, fee decimal.Decimal, reason, modifiableId int, modifiableType string) (err error) {
	if amount.LessThan(decimal.Zero) || fee.GreaterThan(amount) {
		err = fmt.Errorf("cannot add funds (amount: %v)", amount)
		return
	}
	err = account.changeBalanceAndLocked(db, amount, decimal.Zero)
	if err != nil {
		return
	}
	opts := map[string]string{
		"fee":             fee.String(),
		"reason":          strconv.Itoa(reason),
		"modifiable_id":   strconv.Itoa(modifiableId),
		"modifiable_type": modifiableType,
	}
	err = account.after(db, FUNS["PlusFunds"], amount, &opts)
	return
}

func (account *Account) SubFunds(db *utils.GormDB, amount, fee decimal.Decimal, reason, modifiableId int, modifiableType string) (err error) {
	if amount.LessThan(decimal.Zero) || fee.GreaterThan(amount) {
		err = fmt.Errorf("cannot add funds (amount: %v)", amount)
		return
	}
	err = account.changeBalanceAndLocked(db, amount.Neg(), decimal.Zero)
	if err != nil {
		return
	}
	opts := map[string]string{
		"fee":             fee.String(),
		"reason":          strconv.Itoa(reason),
		"modifiable_id":   strconv.Itoa(modifiableId),
		"modifiable_type": modifiableType,
	}
	err = account.after(db, FUNS["SubFunds"], amount, &opts)
	return
}

func (account *Account) LockFunds(db *utils.GormDB, amount decimal.Decimal, reason, modifiableId int, modifiableType string) (err error) {
	if amount.LessThanOrEqual(decimal.Zero) || amount.GreaterThan(account.Balance) {
		err = fmt.Errorf("cannot lock funds (amount: %v)", amount)
		return
	}

	err = account.changeBalanceAndLocked(db, amount.Neg(), amount)
	if err != nil {
		return
	}
	opts := map[string]string{
		"reason":          strconv.Itoa(reason),
		"modifiable_id":   strconv.Itoa(modifiableId),
		"modifiable_type": modifiableType,
	}
	err = account.after(db, FUNS["LockFunds"], amount, &opts)
	return
}

func (account *Account) UnlockFunds(db *utils.GormDB, amount decimal.Decimal, reason, modifiableId int, modifiableType string) (err error) {
	if amount.LessThanOrEqual(decimal.Zero) || amount.GreaterThan(account.Locked) {
		err = fmt.Errorf("cannot unlock funds (amount: %v)", amount)
		return
	}

	err = account.changeBalanceAndLocked(db, amount, amount.Neg())
	if err != nil {
		return
	}
	opts := map[string]string{
		"reason":          strconv.Itoa(reason),
		"modifiable_id":   strconv.Itoa(modifiableId),
		"modifiable_type": modifiableType,
	}
	err = account.after(db, FUNS["UnlockFunds"], amount, &opts)
	return
}

func (account *Account) UnlockedAndSubFunds(db *utils.GormDB, amount, locked, fee decimal.Decimal, reason, modifiableId int, modifiableType string) (err error) {
	if amount.LessThanOrEqual(decimal.Zero) || amount.GreaterThan(locked) {
		err = fmt.Errorf("cannot unlock and subtract funds (amount: %v)", amount)
		return
	}
	if locked.LessThanOrEqual(decimal.Zero) {
		err = fmt.Errorf("invalid lock amount")
		return
	}
	if locked.GreaterThan(account.Locked) {
		err = fmt.Errorf("Account# %v invalid lock amount (amount: %v, locked: %v, self.locked: %v)", account.Id, amount, locked, account.Locked)
		return
	}
	err = account.changeBalanceAndLocked(db, locked.Sub(amount), locked.Neg())
	if err != nil {
		return
	}
	opts := map[string]string{
		"fee":             fee.String(),
		"locked":          locked.String(),
		"reason":          strconv.Itoa(reason),
		"modifiable_id":   strconv.Itoa(modifiableId),
		"modifiable_type": modifiableType,
	}
	err = account.after(db, FUNS["UnlockedAndSubFunds"], amount, &opts)
	return
}

func (account *Account) after(db *utils.GormDB, fun int, amount decimal.Decimal, opts *map[string]string) (err error) {
	var fee decimal.Decimal
	if (*opts)["fee"] != "" {
		fee, _ = decimal.NewFromString((*opts)["fee"])
	}
	var reason int
	if (*opts)["reason"] == "" {
		reason = UNKNOWN
	}
	attributes := map[string]string{
		"fun":             strconv.Itoa(fun),
		"fee":             fee.String(),
		"reason":          strconv.Itoa(reason),
		"amount":          account.Amount().String(),
		"currency_id":     strconv.Itoa(account.CurrencyId),
		"user_id":         strconv.Itoa(account.UserId),
		"account_id":      strconv.Itoa(account.Id),
		"modifiable_id":   (*opts)["modifiable_id"],
		"modifiable_type": (*opts)["modifiable_type"],
	}
	attributes["locked"], attributes["balance"], err = computeLockedAndBalance(fun, amount, opts)
	if err != nil {
		return
	}
	err = optimisticallyLockAccountAndCreate(db, account.Balance, account.Locked, &attributes)
	return
}

func (account *Account) changeBalanceAndLocked(db *utils.GormDB, deltaB, deltaL decimal.Decimal) (err error) {
	db.Set("gorm:query_option", "FOR UPDATE").First(&account, account.Id)
	account.Balance = account.Balance.Add(deltaB)
	account.Locked = account.Locked.Add(deltaL)
	updateSql := fmt.Sprintf("UPDATE accounts SET balance = balance + %v, locked = locked + %v WHERE accounts.id = %v ", deltaB, deltaL, account.Id)
	accountresult := db.Exec(updateSql)
	if accountresult.RowsAffected != 1 {
		err = fmt.Errorf("Insert row failed.")
	}
	return
}

func computeLockedAndBalance(fun int, amount decimal.Decimal, opts *map[string]string) (locked, balance string, err error) {
	switch fun {
	case 1:
		locked = amount.Neg().String()
		balance = amount.String()
	case 2:
		locked = amount.String()
		balance = amount.Neg().String()
	case 3:
		locked = "0"
		balance = amount.String()
	case 4:
		locked = "0"
		balance = amount.Neg().String()
	case 5:
		l, _ := decimal.NewFromString((*opts)["locked"])
		locked = l.Neg().String()
		balance = l.Sub(amount).String()
	default:
		err = fmt.Errorf("forbidden account operation")
	}
	return
}

func optimisticallyLockAccountAndCreate(db *utils.GormDB, balance, locked decimal.Decimal, attrs *map[string]string) (err error) {
	if (*attrs)["account_id"] == "" {
		err = fmt.Errorf("account must be specified")
	}
	(*attrs)["created_at"] = time.Now().Format("2006-01-02 15:04:05")
	(*attrs)["updated_at"] = (*attrs)["created_at"]

	sql := `INSERT INTO account_versions (user_id, account_id, reason, balance, locked, fee, amount, modifiable_id, modifiable_type, currency_id, fun, created_at, updated_at) SELECT ?,?,?,?,?,?,?,?,?,?,?,?,?  FROM accounts WHERE accounts.balance = ? AND accounts.locked = ? AND accounts.id = ?`
	result := db.Exec(sql, (*attrs)["user_id"], (*attrs)["account_id"], (*attrs)["reason"], (*attrs)["balance"], (*attrs)["locked"], (*attrs)["fee"], (*attrs)["amount"], (*attrs)["modifiable_id"], (*attrs)["modifiable_type"], (*attrs)["currency_id"], (*attrs)["fun"], (*attrs)["created_at"], (*attrs)["updated_at"], balance, locked, (*attrs)["account_id"])
	if result.RowsAffected != 1 {
		err = fmt.Errorf("Insert row failed.")
	}
	return
}
