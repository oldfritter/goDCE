package treat

import (
	"encoding/json"
	"log"
	"time"

	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
	"github.com/oldfritter/matching"
)

func Treat(payloadJson *[]byte, workerLog *log.Logger) {
	var offer matching.Offer
	json.Unmarshal([]byte(*payloadJson), &offer)
	tryCreateTrade(&offer, 9)
	return
}

func tryCreateTrade(offer *matching.Offer, times int) {
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	var ask, bid Order
	if mainDB.Where("id = ?", offer.AskId).First(&ask).RecordNotFound() {
		return
	}
	if mainDB.Where("id = ?", offer.BidId).First(&bid).RecordNotFound() {
		return
	}

	trade := Trade{
		MarketId:  (*offer).MarketId,
		AskId:     (*offer).AskId,
		BidId:     (*offer).BidId,
		Price:     (*offer).StrikePrice,
		Volume:    (*offer).Volume,
		Funds:     (*offer).Funds,
		AskUserId: ask.UserId,
		BidUserId: bid.UserId,
	}
	result := mainDB.Create(&trade)
	if result.Error != nil {
		mainDB.DbRollback()
		if times > 0 {
			trade.Id = 0
			tryCreateTrade(offer, times-1)
			return
		}
	}

	errAsk := ask.Strike(mainDB, trade)
	errBid := bid.Strike(mainDB, trade)
	if errAsk == nil && errBid == nil {
		mainDB.DbCommit()
		return
	}
	mainDB.DbRollback()
	if times > 0 {
		trade.Id = 0
		time.Sleep(500 * time.Millisecond)
		tryCreateTrade(offer, times-1)
	}
	return
}
