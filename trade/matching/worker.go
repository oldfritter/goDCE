package matching

import (
	"encoding/json"
	"log"

	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/matching"
)

func doMatching(payloadJson *[]byte, workerLog *log.Logger) {
	var payload MatchingPayload
	json.Unmarshal([]byte(*payloadJson), &payload)
	order, err := matching.InitializeOrder(payload.OrderAttrs())
	if err != nil {
		return
	}
	engine := Engines[order.MarketId]
	if payload.Action == "submit" {
		engine.Submit(order)
	} else if payload.Action == "cancel" {
		engine.Cancel(order)
	}
}
