package order

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func WaitingOrderCheck() {
	db := utils.MainDbBegin()
	defer db.DbRollback()

	for _, market := range Markets {
		ordersPerMarket(db, &market)
	}
	db.DbCommit()
}

func ordersPerMarket(db *utils.GormDB, market *Market) {
	dataRedis := utils.GetRedisConn("data")
	defer dataRedis.Close()

	var orderBook struct {
		AskIds []int `json:"ask_ids"`
		BidIds []int `json:"bid_ids"`
	}
	key := "goDCE:order_book:" + (*market).Code
	values, _ := redis.String(dataRedis.Do("GET", key))
	json.Unmarshal([]byte(values), &orderBook)

	var orders []Order
	if !db.Where("id not in (?)", orderBook.AskIds).
		Where("type = ?", "OrderAsk").
		Where("market_id = ?", (*market).Code).
		Where("state = ?", 100).
		Where("created_at < ?", time.Now().Add(-time.Second*10)).
		Find(&orders).RecordNotFound() {
		var ids []int
		for _, order := range orders {
			ids = append(ids, order.Id)
		}
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "--WaitingOrderCheck orders: ", ids)
	}
	if !db.Where("id not in (?)", orderBook.BidIds).
		Where("type = ?", "OrderBid").
		Where("market_id = ?", (*market).Code).
		Where("state = ?", 100).
		Where("created_at < ?", time.Now().Add(-time.Second*10)).
		Find(&orders).RecordNotFound() {
		var ids []int
		for _, order := range orders {
			ids = append(ids, order.Id)
		}
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "--WaitingOrderCheck orders: ", ids)
	}
}
