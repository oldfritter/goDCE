package sneakerWorkers

import (
	"encoding/json"

	"github.com/oldfritter/goDCE/utils"
	sneaker "github.com/oldfritter/sneaker-go/v3"
	"github.com/shopspring/decimal"

	"github.com/oldfritter/goDCE/config"
	. "github.com/oldfritter/goDCE/models"
)

func InitializeAccountVersionCheckPointWorker() {
	for _, w := range config.AllWorkers {
		if w.Name == "AccountVersionCheckPointWorker" {
			config.AllWorkerIs = append(config.AllWorkerIs, &AccountVersionCheckPointWorker{w})
			return
		}
	}
}

type AccountVersionCheckPointWorker struct {
	sneaker.Worker
}

func (worker *AccountVersionCheckPointWorker) Work(payloadJson *[]byte) (err error) {
	var payload struct {
		AccountId string `json:"account_id"`
	}
	json.Unmarshal([]byte(*payloadJson), &payload)

	db := utils.MainDb
	var account Account
	if db.Where("id = ?", payload.AccountId).First(&account).RecordNotFound() {
		return
	}
	fixAccountVersions(account.Id, 200)
	return
}

func fixAccountVersions(accountId, limit int) {
	db := utils.MainDbBegin()
	defer db.DbRollback()
	var point AccountVersionCheckPoint
	version := 0
	sum, _ := decimal.NewFromString("0")
	if !db.Where("account_id = ?", accountId).First(&point).RecordNotFound() {
		version = point.AccountVersionId
		sum = point.Balance
	}
	var accountVersions []AccountVersion
	if db.Order("id ASC").Where("id > ?", version).Where("account_id = ?", accountId).Limit(limit).Find(&accountVersions).RecordNotFound() {
		return
	}
	for _, av := range accountVersions {
		point.AccountVersionId = av.Id
		sum = sum.Add(av.Balance).Add(av.Locked)
		if sum != av.Amount {
			point.Fixed = "unfixed"
			point.FixedNum = point.FixedNum.Add(av.Amount.Sub(sum))
			db.Save(&point)
			db.DbCommit()
			return
		}
	}
	if point.Fixed == "" {
		point.Fixed = "nomal"
	}
	db.Save(&point)
	db.DbCommit()
	if len(accountVersions) == limit {
		fixAccountVersions(accountId, limit)
	}
	return
}
