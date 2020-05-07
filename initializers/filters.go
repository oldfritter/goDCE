package initializers

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"github.com/oldfritter/goDCE/initializers/locale"
	"github.com/oldfritter/goDCE/utils"
)

func checkSign(context echo.Context, secretKey string, params *map[string]string) bool {
	sign := (*params)["signature"]
	targetStr := context.Request().Method + "|" + context.Path() + "|"

	keys := make([]string, len(*params)-1)
	i := 0
	for k, _ := range *params {
		if k == "signature" {
			continue
		}
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for i, key := range keys {
		if i > 0 {
			targetStr += "&"
		}
		targetStr += key + "=" + (*params)[key]
	}
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(targetStr))
	return sign == fmt.Sprintf("%x", mac.Sum(nil))
}

func LimitTrafficWithIp(context echo.Context) bool {
	dataRedis := utils.GetRedisConn("data")
	defer dataRedis.Close()
	key := "limit-traffic-with-ip:" + context.Path() + ":" + context.RealIP()
	timesStr, _ := redis.String(dataRedis.Do("GET", key))
	if timesStr == "" {
		dataRedis.Do("SETEX", key, 1, 60)
	} else {
		times, _ := strconv.Atoi(timesStr)
		if times > 10 {
			return false
		} else {
			dataRedis.Do("INCR", key)
		}
	}
	return true
}

func LimitTrafficWithEmail(context echo.Context) bool {
	dataRedis := utils.GetRedisConn("data")
	defer dataRedis.Close()
	key := "limit-traffic-with-email-" + context.FormValue("email")
	timesStr, _ := redis.String(dataRedis.Do("GET", key))
	if timesStr == "" {
		dataRedis.Do("Set", key, 1)
		dataRedis.Do("EXPIRE", key, 600)
	} else {
		times, _ := strconv.Atoi(timesStr)
		if times > 10 {
			return false
		} else {
			dataRedis.Do("INCR", key)
		}
	}
	return true
}

func treatLanguage(context echo.Context) {
	var language string
	var lqs []locale.LangQ
	if context.QueryParam("lang") != "" {
		lqs = locale.ParseAcceptLanguage(context.QueryParam("lang"))
	} else {
		lqs = locale.ParseAcceptLanguage(context.Request().Header.Get("Accept-Language"))
	}
	if lqs[0].Lang == "en" {
		language = "en"
	} else if lqs[0].Lang == "ja" {
		language = "ja"
	} else if lqs[0].Lang == "ko" {
		language = "ko"
	} else {
		language = "zh-CN"
	}
	context.Set("language", language)
}

func checkTimestamp(context echo.Context, params *map[string]string) bool {
	timestamp, _ := strconv.Atoi((*params)["timestamp"])
	now := time.Now()
	if int(now.Add(-time.Second*60*5).Unix()) <= timestamp && timestamp <= int(now.Add(time.Second*60*5).Unix()) {
		return true
	}
	return false
}

func IsRabbitMqConnected() bool {
	c := RabbitMqConnect
	ok := true
	if c.IsClosed() {
		fmt.Println("Connection state: closed")
		ok = false
	}
	return ok
}
