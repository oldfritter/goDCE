package initializers

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/oldfritter/goDCE/config"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

type Payload struct {
	Update string `json:"update"`
	Symbol int    `json:"symbol"`
}

func InitCacheData() {
	db := utils.MainDbBegin()
	defer db.DbRollback()
	InitAllCurrencies(db)
	InitAllMarkets(db)
}

func LoadCacheData() {
	InitCacheData()
	go func() {
		channel, err := config.RabbitMqConnect.Channel()
		if err != nil {
			log.Println(fmt.Errorf("Channel: %s", err))
		}
		channel.ExchangeDeclare(config.AmqpGlobalConfig.Exchange["fanout"]["default"], "fanout", true, false, false, false, nil)
		queue, err := channel.QueueDeclare("", true, true, false, false, nil)
		if err != nil {
			return
		}
		channel.QueueBind(queue.Name, queue.Name, config.AmqpGlobalConfig.Exchange["fanout"]["default"], false, nil)
		msgs, _ := channel.Consume(queue.Name, "", true, false, false, false, nil)
		for d := range msgs {
			var payload Payload
			err := json.Unmarshal(d.Body, &payload)
			if err == nil {
				reflect.ValueOf(&payload).MethodByName(payload.Update).Call([]reflect.Value{reflect.ValueOf(payload.Symbol)})
			} else {
				log.Println(fmt.Sprintf("{error: %v}", err))
			}
		}
		return
	}()
}

func (payload *Payload) ReloadCurrencies() {
	db := utils.MainDbBegin()
	defer db.DbRollback()
	InitAllCurrencies(db)
}

func (payload *Payload) ReloadMarkets() {
	db := utils.MainDbBegin()
	defer db.DbRollback()
	InitAllMarkets(db)
}
