package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
	"github.com/shopspring/decimal"
)

func V1GetK(context echo.Context) error {
	var market Market
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	if mainDB.Where("name = ?", context.QueryParam("market")).First(&market).RecordNotFound() {
		return utils.BuildError("1021")
	}
	limit := 30
	if context.QueryParam("limit") != "" {
		limit, _ = strconv.Atoi(context.QueryParam("limit"))
		if limit > 10000 {
			limit = 10000
		}
	}
	period := 1
	periodstr := context.QueryParam("period")
	periods := []string{"1", "5", "15", "30", "60", "120", "240", "360", "720", "1440", "4320", "10080"}
	periodBool := false
	for _, per := range periods {
		if periodstr == per {
			periodBool = true
		}
	}
	if periodBool {
		period, _ = strconv.Atoi(periodstr)
	} else {
		return utils.BuildError("1053")
	}

	kRedis := utils.GetRedisConn("k")
	defer kRedis.Close()

	key := market.KLineRedisKey(periodstr)
	var items [][6]decimal.Decimal
	if context.QueryParam("timestamp") != "" {
		stamp, _ := strconv.Atoi(context.QueryParam("timestamp"))
		timestamp := int64(stamp)
		value, _ := redis.String(kRedis.Do("LINDEX", key, 0))
		var item [6]decimal.Decimal
		json.Unmarshal([]byte(value), &item)

		offset := (timestamp - item[0].IntPart()) / 60 / int64(period)
		if offset < 0 {
			offset = 0
		}

		values, err := redis.Values(kRedis.Do("LRANGE", key, offset, offset+int64(limit)-1))
		if err != nil {
			fmt.Println("lrange err", err.Error())
			return utils.BuildError("1026")
		}
		for _, v := range values {
			json.Unmarshal(v.([]byte), &item)
			items = append(items, item)
		}
	} else {
		length, _ := kRedis.Do("LLEN", key)
		offset := length.(int64) - int64(limit)
		if offset < 0 {
			offset = 0
		}
		values, err := redis.Values(kRedis.Do("LRANGE", key, offset, -1))
		if err != nil {
			fmt.Println("lrange err", err.Error())
			return utils.BuildError("1026")
		}
		for _, v := range values {
			var item [6]decimal.Decimal
			json.Unmarshal(v.([]byte), &item)
			items = append(items, item)
		}
	}
	response := utils.SuccessResponse
	response.Body = items
	return context.JSON(http.StatusOK, response)
}

type chart struct {
	K1     [][6]decimal.Decimal `json:"k1"`
	K5     [][6]decimal.Decimal `json:"k5"`
	K15    [][6]decimal.Decimal `json:"k15"`
	K30    [][6]decimal.Decimal `json:"k30"`
	K60    [][6]decimal.Decimal `json:"k60"`
	K120   [][6]decimal.Decimal `json:"k120"`
	K240   [][6]decimal.Decimal `json:"k240"`
	K360   [][6]decimal.Decimal `json:"k360"`
	K720   [][6]decimal.Decimal `json:"k720"`
	K1440  [][6]decimal.Decimal `json:"k1440"`
	K4320  [][6]decimal.Decimal `json:"k4320"`
	K10080 [][6]decimal.Decimal `json:"k10080"`
}

func V1GetChart(context echo.Context) error {
	var market Market
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	if mainDB.Where("name = ?", context.QueryParam("market")).First(&market).RecordNotFound() {
		return utils.BuildError("1021")
	}
	kRedis := utils.GetRedisConn("k")
	defer kRedis.Close()

	var c chart
	periods := []string{"1", "5", "15", "30", "60", "120", "240", "360", "720", "1440", "4320", "10080"}
	for _, period := range periods {
		key := market.KLineRedisKey(period)
		length, _ := kRedis.Do("LLEN", key)
		offset := length.(int64) - 240
		if offset < 0 {
			offset = 0
		}
		values, err := redis.Values(kRedis.Do("LRANGE", key, offset, -1))
		if err != nil {
			fmt.Println("lrange err", err.Error())
			return utils.BuildError("1026")
		}
		var items [][6]decimal.Decimal
		for _, v := range values {
			var item [6]decimal.Decimal
			json.Unmarshal(v.([]byte), &item)
			items = append(items, item)
		}
		switch period {
		case "1":
			c.K1 = items
		case "5":
			c.K5 = items
		case "15":
			c.K15 = items
		case "30":
			c.K30 = items
		case "60":
			c.K60 = items
		case "120":
			c.K120 = items
		case "240":
			c.K240 = items
		case "360":
			c.K360 = items
		case "720":
			c.K720 = items
		case "1440":
			c.K1440 = items
		case "4320":
			c.K4320 = items
		case "10080":
			c.K10080 = items
		}
	}
	response := utils.SuccessResponse
	response.Body = c
	return context.JSON(http.StatusOK, response)
}
