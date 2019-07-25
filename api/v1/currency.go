package v1

import (
	"net/http"

	"github.com/labstack/echo"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func V1GetCurrencies(context echo.Context) error {
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()

	var currencies []Currency
	mainDB.Find(&currencies)

	response := utils.SuccessResponse
	response.Body = currencies
	return context.JSON(http.StatusOK, response)
}
