package models

type Identity struct {
	CommonModel
	UserId int    `json:"user_id"`
	Source string `json:"source"` // Email or Phone,
	Symbol string `json:"symbol"` // Email address or Phone number
}
