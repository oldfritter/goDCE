package utils

import (
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"log"
	"strconv"
	"time"
)

type ConfigEnv struct {
	configFile *yaml.File
}

func getDatabaseConfig() *ConfigEnv {
	return getConfig("database")
}
func getRedisConfig() *ConfigEnv {
	return getConfig("redis")
}

func getConfig(name string) *ConfigEnv {
	filePath := fmt.Sprintf("config/%s.yml", name)
	return NewEnv(filePath)
}

func NewEnv(configFile string) *ConfigEnv {
	env := &ConfigEnv{
		configFile: yaml.ConfigFile(configFile),
	}
	if env.configFile == nil {
		panic("go-configenv failed to open configFile: " + configFile)
	}
	return env
}

func (env *ConfigEnv) Get(spec, defaultValue string) string {
	value, err := env.configFile.Get(spec)
	if err != nil {
		value = defaultValue
	}
	return value
}

func (env *ConfigEnv) GetInt(spec string, defaultValue int) int {
	str := env.Get(spec, "")
	if str == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		log.Panic("go-configenv GetInt failed Atoi", spec, str)
	}
	return val
}

func (env *ConfigEnv) GetDuration(spec string, defaultValue string) time.Duration {
	str := env.Get(spec, "")
	if str == "" {
		str = defaultValue
	}
	duration, err := time.ParseDuration(str)
	if err != nil {
		log.Panic("go-configenv GetDuration failed ParseDuration", spec, str)
	}
	return duration
}
