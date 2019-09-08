package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	newrelic "github.com/dafiti/echo-middleware"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	envConfig "github.com/oldfritter/goDCE/config"
	"github.com/oldfritter/goDCE/initializers"
	"github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/routes"
	"github.com/oldfritter/goDCE/utils"
)

func main() {
	initialize()
	e := echo.New()

	e.File("/web", "public/assets/index.html")
	e.File("/web/*", "public/assets/index.html")
	e.Static("/assets", "public/assets")

	if envConfig.CurrentEnv.Newrelic.AppName != "" && envConfig.CurrentEnv.Newrelic.LicenseKey != "" {
		e.Use(newrelic.NewRelic(envConfig.CurrentEnv.Newrelic.AppName, envConfig.CurrentEnv.Newrelic.LicenseKey))
	}
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(initializers.Auth)
	routes.SetV1Interfaces(e)
	e.HTTPErrorHandler = customHTTPErrorHandler
	e.HideBanner = true
	go func() {
		if err := e.Start(":9990"); err != nil {
			fmt.Println("start close echo")
			time.Sleep(500 * time.Millisecond)
			closeResource()
			fmt.Println("shutting down the server")
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("accepted signal")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	initializers.DeleteListeQueue()
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		fmt.Println("shutting down failed, err:" + err.Error())
		e.Logger.Fatal(err)
	}
}

func customHTTPErrorHandler(err error, context echo.Context) {
	language := context.Get("language").(string)
	if response, ok := err.(utils.Response); ok {
		response.Head["msg"] = fmt.Sprint(initializers.I18n.T(language, "error_code."+response.Head["code"]))
		context.JSON(http.StatusBadRequest, response)
	} else {
		panic(err)
	}
}

func initialize() {
	envConfig.InitEnv()
	utils.InitMainDB()
	utils.InitBackupDB()
	models.AutoMigrations()
	utils.InitRedisPools()
	utils.InitializeAmqpConfig()

	initializers.LoadInterfaces()
	initializers.InitI18n()
	initializers.LoadCacheData()

	err := ioutil.WriteFile("pids/api.pid", []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func closeResource() {
	utils.CloseAmqpConnection()
	utils.CloseRedisPools()
	utils.CloseMainDB()
}
