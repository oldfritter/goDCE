package models

import (
	"github.com/jinzhu/gorm"
	"github.com/oldfritter/goDCE/utils"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	CommonModel
	Sn             string `gorm:"type:varchar(32);default:null" json:"sn"`
	PasswordDigest string `gorm:"type:varchar(64);default:null"`
	Nickname       string `gorm:"type:varchar(32);default:null" json:"nickname"`
	State          int    `gorm:"default:null" json:"state"`
	Activated      bool   `gorm:"default:null" json:"activated"`
	Disabled       bool   `json:"disabled"`
	ApiDisabled    bool   `json:"api_disabled"`

	Password string    `sql:"-"`
	Tokens   []Token   `sql:"-" json:"tokens"`
	Accounts []Account `sql:"-" json:"accounts"`
}

func (user *User) GenerateSn() {
	user.Sn = "PEA" + utils.RandStringRunes(8) + "TIO"
}

func (user *User) AfterSave(db *gorm.DB) {
}

func (user *User) CompareHashAndPassword() bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(user.Password))
	if err == nil {
		return true
	}
	return false
}

func (user *User) SetPasswordDigest() {
	b, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.PasswordDigest = string(b[:])
}
