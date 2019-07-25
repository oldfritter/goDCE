package v1

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func V1GetUsersMe(context echo.Context) error {
	response := utils.SuccessResponse
	response.Body = context.Get("current_user")
	return context.JSON(http.StatusOK, response)
}

func V1GetUsersAccounts(context echo.Context) error {
	user := context.Get("current_user").(User)
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	var accounts []Account
	mainDB.Joins("INNER JOIN (currencies) ON (accounts.currency_id = currencies.id)").
		Where("user_id = ?", user.Id).Where("currencies.visible is true").Find(&accounts)

	response := utils.SuccessResponse
	response.Body = accounts
	return context.JSON(http.StatusOK, response)
}

func V1GetUsersAccountsCurrency(context echo.Context) error {
	user := context.Get("current_user").(User)
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	var currency Currency
	if mainDB.Where("code = ?", strings.ToLower(context.Param("currency"))).First(&currency).RecordNotFound() {
		return utils.BuildError("3027")
	}
	var account Account
	mainDB.Where("user_id = ? AND currency= ?", user.Id, currency.Id).First(&account)

	response := utils.SuccessResponse
	response.Body = account
	return context.JSON(http.StatusOK, response)
}

func V1PostUsersAccountsCurrency(context echo.Context) error {
	user := context.Get("current_user").(User)
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	var currency Currency
	if mainDB.Where("key = ?", strings.ToLower(context.Param("currency"))).First(&currency).RecordNotFound() {
		return utils.BuildError("1027")
	}
	var account Account
	if mainDB.Where("user_id = ? AND currency_id = ?", user.Id, currency.Id).First(&account).RecordNotFound() {
		account.UserId = user.Id
		account.CurrencyId = currency.Id
		now := time.Now()
		account.CreatedAt = now
		account.UpdatedAt = now
		mainDB.Save(&account)
		mainDB.DbCommit()
	}

	response := utils.SuccessResponse
	response.Body = account
	return context.JSON(http.StatusOK, response)
}

func V1PostUsersLogin(context echo.Context) error {
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	var user User
	if mainDB.Joins("INNER JOIN (identities) ON (identities.user_id users.id)").
		Where("identities.source = ?", context.FormValue("source")).
		Where("identities.symbol = ?", context.FormValue("symbol")).First(&user).RecordNotFound() {
		return utils.BuildError("2026")
	}
	user.Password = context.FormValue("password")
	if user.CompareHashAndPassword() {
		context.Set("current_user", user)
	} else {
		return utils.BuildError("2026")
	}

	var token, inToken Token
	token.UserId = user.Id
	token.InitializeLoginToken()
	if !mainDB.Where("token = ?", token.Token).First(&inToken).RecordNotFound() {
		token.InitializeLoginToken()
	}
	mainDB.Create(&token)
	mainDB.DbCommit()

	response := utils.SuccessResponse
	response.Body = user
	return context.JSON(http.StatusOK, response)
}
