package kLine

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
	"github.com/streadway/amqp"
)

func CreateLatestKLine() {

	markets := FindAllMarket()
	for _, market := range markets {
		periods := []int64{1, 5, 15, 30, 60, 120, 240, 360, 720, 1440, 4320, 10080}
		for _, period := range periods {
			payload := struct {
				MarketId   int    `json:"market_id"`
				Timestamp  int64  `json:"timestamp"`
				Period     int64  `json:"period"`
				DataSource string `json:"data_source"`
			}{
				MarketId:   market.Id,
				Timestamp:  time.Now().Unix(),
				Period:     period,
				DataSource: "db",
			}
			b, err := json.Marshal(payload)
			if err != nil {
				fmt.Println("error:", err)
			}
			err = initializers.PublishMessageWithRouteKey("goDCE.default", "goDCE.k", "text/plain", &b, amqp.Table{}, amqp.Persistent)
			if err != nil {
				fmt.Println("{ error:", err, "}")
				panic(err)
			}
		}
	}

}
