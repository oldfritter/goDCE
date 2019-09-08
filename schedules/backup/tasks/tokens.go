package tasks

import (
	"time"

	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func CleanTokens() {
	utils.InitMainDB()
	db := utils.MainDbBegin()
	defer db.DbRollback()

	db.Where("expire_at < ?", time.Now().Add(-time.Hour*8)).Delete(Token{})
	db.DbCommit()
	utils.CloseMainDB()
}
