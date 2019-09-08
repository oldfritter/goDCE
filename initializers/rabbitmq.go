package initializers

import (
	"fmt"

	"github.com/oldfritter/goDCE/utils"
)

func IsRabbitMqConnected() bool {
	c := utils.RabbitMqConnect
	ok := true
	if c.IsClosed() {
		fmt.Println("Connection state: closed")
		ok = false
	}
	return ok
}
