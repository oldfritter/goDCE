package initializers

import (
	i18n "github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"
)

var I18n *i18n.I18n

func InitI18n() {
	I18n = i18n.New(yaml.New("initializers/locales"))
}
