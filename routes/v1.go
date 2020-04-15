package routes

import (
	"github.com/labstack/echo"
	. "github.com/oldfritter/goDCE/api/v1"
)

func SetV1Interfaces(e *echo.Echo) {

	e.GET("/api/:platform/v1/currencies", V1GetCurrencies)
	e.GET("/api/:platform/v1/k", V1GetK)
	e.GET("/api/:platform/v1/chart", V1GetChart)
	e.GET("/api/:platform/v1/markets", V1GetMarkets)
	e.POST("/api/:platform/v1/users/login", V1PostUsersLogin)
	e.GET("/api/:platform/v1/users/me", V1GetUsersMe)
	e.GET("/api/:platform/v1/users/accounts/:currency", V1GetUsersAccountsCurrency)
	e.GET("/api/:platform/v1/users/accounts", V1GetUsersAccounts)
	e.GET("/api/:platform/v1/depth", V1Getdepth)
	e.GET("/api/:platform/v1/order", V1GetOrder)
	e.GET("/api/:platform/v1/orders", V1GetOrders)
	e.POST("/api/:platform/v1/orders", V1PostOrders)
	e.POST("/api/:platform/v1/order/delete", V1PostOrderDelete)
	e.POST("/api/:platform/v1/orders/clear", V1PostOrdersClear)
	e.GET("/api/:platform/v1/tickers", V1GetTickers)
	e.GET("/api/:platform/v1/tickers/:market", V1GetTickersMarket)
	e.GET("/api/:platform/v1/trades", V1GetTrades)
	e.GET("/api/:platform/v1/trades/my", V1GetTradesMy)

}
