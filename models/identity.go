package models

type Identity struct {
	CommonModel
	UserId int    `json:"user_id"`
	Source string `json:"source" gorm:"type:varchar(32)"` // Email or Phone,
	Symbol string `json:"symbol" gorm:"type:varchar(64)"` // Email address or Phone number
}
