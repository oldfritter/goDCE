package models

import (
	"github.com/oldfritter/goDCE/utils"
)

type Device struct {
	CommonModel
	UserId    int    `json:"user_id"`
	IsUsed    bool   `json:"is_used"`
	Token     string `json:"token"`
	PublicKey string `json:"-"`
}

func (device *Device) InitializeToken() {
	device.Token = utils.RandStringRunes(64)
}
