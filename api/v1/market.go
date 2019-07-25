package v1

import (
	"net/http"

	"github.com/labstack/echo"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func V1GetMarkets(context echo.Context) error {
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()

	var markets []Market
	mainDB.Find(&markets)

	response := utils.SuccessResponse
	response.Body = markets
	return context.JSON(http.StatusOK, response)
}
