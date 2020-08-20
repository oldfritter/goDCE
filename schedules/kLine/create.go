package kLine

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/streadway/amqp"

	"github.com/oldfritter/goDCE/config"
	. "github.com/oldfritter/goDCE/models"
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
			err = config.RabbitMqConnect.PublishMessageWithRouteKey("goDCE.default", "goDCE.k", "text/plain", false, false, &b, amqp.Table{}, amqp.Persistent, "")
			if err != nil {
				fmt.Println("{ error:", err, "}")
				panic(err)
			}
		}
	}

}
