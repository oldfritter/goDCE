package models

import (
	"time"
)

type ApiToken struct {
	CommonModel
	UserId    int       `json:"user_id"`
	AccessKey string    `json:"access_key"`
	SecretKey string    `json:"secret_key"`
	Label     string    `json:"label"`
	Scopes    string    `json:"scopes"`
	ExpireAt  time.Time `json:"expire_at" gorm:"default:null"`
	DeletedAt time.Time `json:"-" gorm:"default:null"`
}
