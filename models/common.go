package models

import (
	"time"

	"github.com/oldfritter/goDCE/utils"
)

type CommonModel struct {
	Id        int       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func AutoMigrations() {
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	backupDB := utils.BackupDbBegin()
	defer backupDB.DbRollback()

	// account_version_check_point
	mainDB.AutoMigrate(&AccountVersionCheckPoint{})
	mainDB.Model(&AccountVersionCheckPoint{}).AddIndex("index_account_version_check_points_on_account_id", "account_id")

	// account_version
	mainDB.AutoMigrate(&AccountVersion{})
	mainDB.Model(&AccountVersion{}).AddIndex("index_account_versions_on_user_id_and_reason", "user_id", "reason")
	mainDB.Model(&AccountVersion{}).AddIndex("index_account_versions_on_account_id_and_reason", "account_id", "reason")
	mainDB.Model(&AccountVersion{}).AddIndex("index_account_versions_on_modifiable_id_and_modifiable_type", "modifiable_id", "modifiable_type")
	mainDB.Model(&AccountVersion{}).AddIndex("index_account_versions_on_currency_id_and_created_at", "currency_id", "created_at")

	backupDB.AutoMigrate(&AccountVersion{})
	backupDB.Model(&AccountVersion{}).AddIndex("index_account_versions_on_user_id_and_reason", "user_id", "reason")
	backupDB.Model(&AccountVersion{}).AddIndex("index_account_versions_on_account_id_and_reason", "account_id", "reason")
	backupDB.Model(&AccountVersion{}).AddIndex("index_account_versions_on_modifiable_id_and_modifiable_type", "modifiable_id", "modifiable_type")
	backupDB.Model(&AccountVersion{}).AddIndex("index_account_versions_on_currency_id_and_created_at", "currency_id", "created_at")

	// account
	mainDB.AutoMigrate(&Account{})

	// api_token
	mainDB.AutoMigrate(&ApiToken{})
	mainDB.Model(&ApiToken{}).AddIndex("api_tokens_idx0", "access_key", "deleted_at")

	// currency
	mainDB.AutoMigrate(&Currency{})

	// device
	mainDB.AutoMigrate(&Device{})

	// identity
	mainDB.AutoMigrate(&Identity{})
	mainDB.Model(&Identity{}).AddIndex("identity_idx0", "source", "symbol")

	// k
	backupDB.AutoMigrate(&KLine{})
	backupDB.Model(&KLine{}).AddUniqueIndex("k_line_idx0", "market_id", "period", "timestamp")

	// market
	mainDB.AutoMigrate(&Market{})
	mainDB.Model(&Market{}).AddUniqueIndex("markets_idx0", "code")

	// order
	mainDB.AutoMigrate(&Order{})
	backupDB.AutoMigrate(&Order{})

	// token
	mainDB.AutoMigrate(&Token{})
	backupDB.AutoMigrate(&Token{})

	// trade
	mainDB.AutoMigrate(&Trade{})
	backupDB.AutoMigrate(&Trade{})

	// user
	mainDB.AutoMigrate(&User{})

}
