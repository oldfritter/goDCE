package initializers

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"github.com/labstack/echo"
	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
	"gopkg.in/yaml.v2"
)

type ApiInterface struct {
	Method                string `yaml:"method"`
	Path                  string `yaml:"path"`
	Auth                  bool   `yaml:"auth"`
	Sign                  bool   `yaml:"sign"`
	CheckTimestamp        bool   `yaml:"check_timestamp"`
	LimitTrafficWithIp    bool   `yaml:"limit_traffic_with_ip"`
	LimitTrafficWithEmail bool   `yaml:"limit_traffic_with_email"`
}

var GlobalApiInterfaces []ApiInterface

func LoadInterfaces() {
	path_str, _ := filepath.Abs("config/interfaces.yml")
	content, err := ioutil.ReadFile(path_str)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = yaml.Unmarshal(content, &GlobalApiInterfaces)
	if err != nil {
		log.Fatal(err)
	}
}

func Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(context echo.Context) error {

		treatLanguage(context)

		params := make(map[string]string)

		for k, v := range context.QueryParams() {
			params[k] = v[0]
		}
		values, _ := context.FormParams()
		for k, v := range values {
			params[k] = v[0]
		}
		context.Set("params", params)
		var currentApiInterface ApiInterface
		for _, apiInterface := range GlobalApiInterfaces {
			if context.Path() == apiInterface.Path && context.Request().Method == apiInterface.Method {
				currentApiInterface = apiInterface
				// limit_traffic_with_email
				if currentApiInterface.LimitTrafficWithEmail && LimitTrafficWithEmail(context) != true {
					return utils.BuildError("3043")
				}
				// limit_traffic_with_ip
				if currentApiInterface.LimitTrafficWithIp && LimitTrafficWithIp(context) != true {
					return utils.BuildError("3043")
				}
				if apiInterface.Auth != true {
					return next(context)
				}
			}
		}

		if currentApiInterface.Path == "" {
			return utils.BuildError("3025")
		}
		if context.Request().Header.Get("Authorization") == "" {
			return utils.BuildError("4001")
		}
		if currentApiInterface.CheckTimestamp && checkTimestamp(context, &params) == false {
			return utils.BuildError("3050")
		}

		db := utils.MainDbBegin()
		defer db.DbRollback()

		var user User
		var token Token
		var apiToken ApiToken
		var device Device
		var err error
		if context.Param("platform") == "client" {
			user, apiToken, err = robotAuth(context, &params, db)
			if currentApiInterface.Sign && checkSign(context, apiToken.SecretKey, &params) == false {
				return utils.BuildError("4005")
			}
		} else if context.Param("platform") == "mobile" {
			user, device, err = mobileAuth(context, &params, db)
		} else if context.Param("platform") == "web" {
			user, token, err = webAuth(context, &params, db)
		}
		if err != nil {
			return err
		}

		db.DbCommit()
		context.Set("current_user", user)
		context.Set("current_token", token.Token)
		context.Set("current_user_id", user.Id)
		context.Set("current_device", device.Token)
		context.Set("current_api_token", apiToken.AccessKey)
		return next(context)
	}
}

func robotAuth(context echo.Context, params *map[string]string, db *utils.GormDB) (user User, apiToken ApiToken, err error) {
	if db.Where("access_key = ?", (*params)["access_key"]).First(&apiToken).RecordNotFound() {
		return user, apiToken, utils.BuildError("2008")
	}
	if db.Where("id = ?", apiToken.UserId).First(&user).RecordNotFound() {
		return user, apiToken, utils.BuildError("2016")
	}
	return
}

func mobileAuth(context echo.Context, params *map[string]string, db *utils.GormDB) (user User, device Device, err error) {
	if db.Where("token = ?", context.Request().Header.Get("Authorization")).First(&device).RecordNotFound() {
		return user, device, utils.BuildError("4016")
	}
	if db.Where("id = ?", device.UserId).First(&user).RecordNotFound() {
		return user, device, utils.BuildError("4016")
	}
	return
}

func webAuth(context echo.Context, params *map[string]string, db *utils.GormDB) (user User, token Token, err error) {
	if db.Where("token = ? AND ? < expire_at", context.Request().Header.Get("Authorization"), time.Now()).First(&token).RecordNotFound() {
		return user, token, utils.BuildError("4016")
	}
	if db.Where("id = ?", token.UserId).First(&user).RecordNotFound() {
		return user, token, utils.BuildError("4016")
	}
	return
}
