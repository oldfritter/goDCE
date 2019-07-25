package config

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Env struct {
	Model string `yaml:"model"`

	Matching struct {
		Node string `yaml:"node"`
	} `yaml:"matching"`

	TradeTreat struct {
		Node string `yaml:"node"`
	} `yaml:"trade_treat"`

	OrderCancel struct {
		Node string `yaml:"node"`
	} `yaml:"order_cancel"`

	Newrelic struct {
		AppName    string `yaml:"app_name"`
		LicenseKey string `yaml:"license_key"`
	} `yaml:"newrelic"`

	Host string `yaml:"HOST"`

	Email struct {
		SmtpPort       int    `yaml:"SMTP_PORT"`
		SmtpDomain     string `yaml:"SMTP_DOMAIN"`
		SmtpAddress    string `yaml:"SMTP_ADDRESS"`
		SmtpUsername   string `yaml:"SMTP_USERNAME"`
		SmtpPassword   string `yaml:"SMTP_PASSWORD"`
		SystemMailFrom string `yaml:"SYSTEM_MAIL_FROM"`
	} `yaml:"EMAIL"`

	Sms struct {
		SmsUser        string `yaml:"SMS_USER"`
		SmsKey         string `yaml:"SMS_KEY"`
		ValidateCode   string `yaml:"VALIDATE_CODE"`
		ApiUrl         string `yaml:"API_URL"`
		Sms253Account  string `yaml:"SMS_253_ACCOUNT"`
		Sms253Password string `yaml:"SMS_253_PASSWORD"`
		VoiceApiUrl    string `yaml:"VOICE_API_URL"`
	} `yaml:"SMS"`

	GarnetClient struct {
		ApiBaseUrl  string `yaml:"ApiBaseUrl"`
		TenantId    string `yaml:"TenantId"`
		TenantToken string `yaml:"TenantToken"`
	} `yaml:"GarnetClient"`
}

var CurrentEnv Env

func InitEnv() {
	path_str, _ := filepath.Abs("config/env.yml")
	content, err := ioutil.ReadFile(path_str)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = yaml.Unmarshal(content, &CurrentEnv)
	if err != nil {
		log.Fatal(err)
	}
}
