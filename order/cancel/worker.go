package cancel

import (
	"encoding/json"
	"log"

	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func Cancel(payloadJson *[]byte, workerLog *log.Logger) {
	var payload struct {
		Id int `json:"id"`
	}
	json.Unmarshal([]byte(*payloadJson), &payload)

	db := utils.MainDbBegin()
	defer db.DbRollback()
	var order Order
	if db.Where("id = ?", payload.Id).First(&order).RecordNotFound() {
		return
	}
	order.State = 0
	db.Save(&order)
	var account Account
	db.Where("market_id = ?", order.MarketId).Where("user_id = ?", order.UserId).First(&account)
	account.UnlockFunds(db, order.Locked, ORDER_CANCEL, order.Id, "Order")
	db.DbCommit()
	return
}
