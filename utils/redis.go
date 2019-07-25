package utils

import (
	"github.com/gomodule/redigo/redis"
	"time"
	"fmt"
)

var (
	RailsCachePool *redis.Pool
	DatePool       *redis.Pool
	TickerPool     *redis.Pool
	KLinePool      *redis.Pool
	LimitPool      *redis.Pool
)

func InitRedisPools() {
	RailsCachePool = newRedisPool("cache")
	DatePool = newRedisPool("data")
	TickerPool = newRedisPool("ticker")
	KLinePool = newRedisPool("k")
	LimitPool = newRedisPool("limit")
}

func CloseRedisPools() {
	RailsCachePool.Close()
	DatePool.Close()
	TickerPool.Close()
	KLinePool.Close()
	LimitPool.Close()
}

func GetRedisConn(redisName string) redis.Conn {
	if redisName == "cache" {
		return RailsCachePool.Get()
	} else if redisName == "data" {
		return DatePool.Get()
	} else if redisName == "ticker" {
		return TickerPool.Get()
	} else if redisName == "k" {
		return KLinePool.Get()
	} else if redisName == "limit" {
		return LimitPool.Get()
	}
	return nil
}

func newRedisPool(redisName string) *redis.Pool {
	config := getRedisConfig()
	capacity := config.GetInt(redisName+".pool", 10)
	maxCapacity := config.GetInt(redisName+".maxopen", 0)
	idleTimout := config.GetDuration(redisName+".timeout", "4m")
	maxConnLifetime := config.GetDuration(redisName+".life_time", "2m")
	network := config.Get(redisName+".network", "tcp")
	server := config.Get(redisName+".server", "localhost:6379")
	db := config.Get(redisName+".db", "")
	password := config.Get(redisName+".password", "")

	return &redis.Pool{
		MaxIdle:         capacity,
		MaxActive:       maxCapacity,
		IdleTimeout:     idleTimout,
		MaxConnLifetime: maxConnLifetime,
		Wait: true,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(network, server)
			if err != nil {
				fmt.Println("redis can't dial:" + err.Error())
				return nil, err
			}

			if password != "" {
				_, err := conn.Do("AUTH", password)
				if err != nil {
					fmt.Println("redis can't AUTH:" + err.Error())
					conn.Close()
					return nil, err
				}
			}

			if db != "" {
				_, err := conn.Do("SELECT", db)
				if err != nil {
					fmt.Println("redis can't SELECT:" + err.Error())
					conn.Close()
					return nil, err
				}
			}
			return conn, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
            if err != nil {
            	fmt.Println("redis can't ping, err:" + err.Error())
			}
			return err
		},
	}
}
