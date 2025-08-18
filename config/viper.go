package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

func New() *viper.Viper {
	var v = viper.New()

	if vv := os.Getenv("VIPER_ENV_PREFIX"); vv != "" {
		v.SetEnvPrefix(vv)
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	for _, vv := range getAvailablePath() {
		v.AddConfigPath(vv)
	}

	if vv := os.Getenv("VIPER_CONFIG_NAME"); vv != "" {
		v.SetConfigName(vv)
	} else {
		v.SetConfigName("config.yaml")
	}

	if vv := os.Getenv("VIPER_CONFIG_TYPE"); vv != "" {
		v.SetConfigType(vv)
	} else {
		v.SetConfigType("yaml")
	}

	_ = v.MergeInConfig()

	return v
}
