package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

type Depth struct {
	Timestamp       int64       `json:"timestamp"`
	Asks            interface{} `json:"asks"`
	Bids            interface{} `json:"bids"`
	PriceGroupFixed string      `json:"price_group_fixed"`
}

func V1Getdepth(context echo.Context) error {
	var market Market
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	tickerRedis := utils.GetRedisConn("ticker")
	defer tickerRedis.Close()
	if mainDB.Where("name = ?", context.QueryParam("market")).First(&market).RecordNotFound() {
		return utils.BuildError("1021")
	}
	limit := 300
	if context.QueryParam("limit") != "" {
		limit, _ = strconv.Atoi(context.QueryParam("limit"))
		if limit > 1000 {
			limit = 300
		}
	}
	vAsk, _ := redis.String(tickerRedis.Do("GET", market.AskRedisKey()))
	vBid, _ := redis.String(tickerRedis.Do("GET", market.BidRedisKey()))
	if vAsk == "" || vBid == "" {
		return utils.BuildError("1026")
	}
	var asksOrigin, bidsOrigin interface{}
	var asks, bids []interface{}
	json.Unmarshal([]byte(vAsk), &asksOrigin)
	json.Unmarshal([]byte(vBid), &bidsOrigin)
	for i, askOrigin := range asksOrigin.([]interface{}) {
		if i < limit {
			asks = append(asks, askOrigin)
		}
	}
	for i, bidOrigin := range bidsOrigin.([]interface{}) {
		if i < limit {
			bids = append(bids, bidOrigin)
		}
	}
	var depth Depth
	depth.PriceGroupFixed = strconv.Itoa(market.PriceGroupFixed)
	depth.Timestamp = time.Now().Unix()
	depth.Asks = asks
	depth.Bids = bids
	return context.JSON(http.StatusOK, depth)
}
