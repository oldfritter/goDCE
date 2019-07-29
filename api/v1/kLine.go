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

type chart struct {
	K1     []interface{} `json:"k1"`
	K5     []interface{} `json:"k5"`
	K15    []interface{} `json:"k15"`
	K30    []interface{} `json:"k30"`
	K60    []interface{} `json:"k60"`
	K120   []interface{} `json:"k120"`
	K240   []interface{} `json:"k240"`
	K360   []interface{} `json:"k360"`
	K720   []interface{} `json:"k720"`
	K1440  []interface{} `json:"k1440"`
	K4320  []interface{} `json:"k4320"`
	K10080 []interface{} `json:"k10080"`
}

func V1GetK(context echo.Context) error {
	var market Market
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	if mainDB.Where("code = ?", context.QueryParam("market")).First(&market).RecordNotFound() {
		return utils.BuildError("1021")
	}
	limit := 100
	if context.QueryParam("limit") != "" {
		limit, _ = strconv.Atoi(context.QueryParam("limit"))
		if limit > 10000 {
			limit = 10000
		}
	}
	period := 1
	periodstr, _ := strconv.Atoi(context.QueryParam("period"))
	periods := []int{1, 5, 15, 30, 60, 120, 240, 360, 720, 1440, 4320, 10080}
	periodBool := false
	for _, per := range periods {
		if periodstr == per {
			periodBool = true
			period = per
		}
	}
	if !periodBool {
		return utils.BuildError("1053")
	}

	kRedis := utils.GetRedisConn("k")
	defer kRedis.Close()

	var timestamp int
	if context.QueryParam("timestamp") == "" {
		timestamp = int(time.Now().Unix())
	} else {
		timestamp, _ = strconv.Atoi(context.QueryParam("timestamp"))
	}
	minTimestamp := timestamp - period*60*limit

	values, _ := redis.Values(
		kRedis.Do(
			"ZREVRANGEBYSCORE",
			market.KLineRedisKey(period),
			timestamp,
			minTimestamp,
		),
	)

	var result []interface{}
	for _, value := range values {
		var item []string
		json.Unmarshal(value.([]byte), &item)
		result = append(result, item)
	}
	response := utils.SuccessResponse
	response.Body = result
	return context.JSON(http.StatusOK, response)
}

func V1GetChart(context echo.Context) error {
	var market Market
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	if mainDB.Where("code = ?", context.QueryParam("market")).First(&market).RecordNotFound() {
		return utils.BuildError("1021")
	}
	kRedis := utils.GetRedisConn("k")
	defer kRedis.Close()
	limit := 100
	timestamp := int(time.Now().Unix())

	var c chart
	periods := []int{1, 5, 15, 30, 60, 120, 240, 360, 720, 1440, 4320, 10080}
	for _, period := range periods {
		minTimestamp := timestamp - period*60*limit
		values, _ := redis.Values(
			kRedis.Do(
				"ZREVRANGEBYSCORE",
				market.KLineRedisKey(period),
				timestamp,
				minTimestamp,
			),
		)
		var items []interface{}
		for _, v := range values {
			var item [6]string
			json.Unmarshal(v.([]byte), &item)
			items = append(items, item)
		}
		switch period {
		case 1:
			c.K1 = items
		case 5:
			c.K5 = items
		case 15:
			c.K15 = items
		case 30:
			c.K30 = items
		case 60:
			c.K60 = items
		case 120:
			c.K120 = items
		case 240:
			c.K240 = items
		case 360:
			c.K360 = items
		case 720:
			c.K720 = items
		case 1440:
			c.K1440 = items
		case 4320:
			c.K4320 = items
		case 10080:
			c.K10080 = items
		}
	}
	response := utils.SuccessResponse
	response.Body = c
	return context.JSON(http.StatusOK, response)
}
