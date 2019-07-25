package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

type TradeCache struct {
	Amount    string `json:"amount"`
	Date      int    `json:"date"`
	OwnerSn   string `json:"owner_sn"`
	Price     string `json:"price"`
	Tid       int    `json:"tid"`
	TradeType string `json:"type"`
}

func V1GetTrades(context echo.Context) error {
	if context.QueryParam("market") == "" {
		return utils.BuildError("1021")
	}
	limit := 30
	if context.QueryParam("limit") != "" {
		limit, _ = strconv.Atoi(context.QueryParam("limit"))
	}
	tickerRedis := utils.GetRedisConn("ticker")
	defer tickerRedis.Close()
	var all, result []TradeCache
	key := "hex:redismodels::latesttrades:" + context.QueryParam("market")
	re, _ := redis.String(tickerRedis.Do("GET", key))
	if re == "" {
		return utils.BuildError("1026")
	}
	json.Unmarshal([]byte(re), &all)
	for i := 0; i < limit && i < len(all); i++ {
		result = append(result, all[i])
	}
	return context.JSON(http.StatusOK, result)
}

func V1GetTradesMy(context echo.Context) error {
	user := context.Get("current_user").(User)
	db := utils.MainDbBegin()
	defer db.DbRollback()
	limit := 30
	if context.QueryParam("limit") != "" {
		limit, _ = strconv.Atoi(context.QueryParam("limit"))
	}
	if context.QueryParam("market") == "" {
		return utils.BuildError("1021")
	}
	var market Market
	if db.Where("name = ?", context.QueryParam("market")).First(&market).RecordNotFound() {
		return utils.BuildError("1021")
	}
	page := 1
	if context.QueryParam("page") != "" {
		page, _ = strconv.Atoi(context.QueryParam("page"))
		if page < 1 {
			page = 1
		}
	}
	orderParam := "id DESC"
	if context.QueryParam("order") == "asc" {
		orderParam = "id ASC"
	}
	var trades []Trade
	db.Order(orderParam).Where("currency = ? AND (ask_user_id = ? OR bid_user_id = ?)", market.Code, user.Id, user.Id).
		Offset(limit * (page - 1)).Limit(limit).Find(&trades)

	for i, trade := range trades {
		if trade.BidUserId == user.Id && trade.AskUserId == user.Id {
			trades[i].Side = "self"
		} else if trade.BidUserId == user.Id {
			trades[i].Side = "buy"
		} else if trade.AskUserId == user.Id {
			trades[i].Side = "sell"
		}
	}
	response := utils.SuccessResponse
	response.Body = trades
	return context.JSON(http.StatusOK, response)
}
