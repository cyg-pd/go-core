package config

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func Parse() {
	var v *viper.Viper = Default()

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	v.BindPFlags(pflag.CommandLine)

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
}

func Default() *viper.Viper                        { return viper.GetViper() }
func BindPFlag(key string, flag *pflag.Flag) error { return Default().BindPFlag(key, flag) }
func Get(key string) any                           { return Default().Get(key) }
func GetString(key string) string                  { return Default().GetString(key) }
func GetBool(key string) bool                      { return Default().GetBool(key) }
func GetInt(key string) int                        { return Default().GetInt(key) }
func GetInt32(key string) int32                    { return Default().GetInt32(key) }
func GetInt64(key string) int64                    { return Default().GetInt64(key) }
func GetUint8(key string) uint8                    { return Default().GetUint8(key) }
func GetUint(key string) uint                      { return Default().GetUint(key) }
func GetUint16(key string) uint16                  { return Default().GetUint16(key) }
func GetUint32(key string) uint32                  { return Default().GetUint32(key) }
func GetUint64(key string) uint64                  { return Default().GetUint64(key) }
func GetFloat64(key string) float64                { return Default().GetFloat64(key) }
func GetTime(key string) time.Time                 { return Default().GetTime(key) }
func GetDuration(key string) time.Duration         { return Default().GetDuration(key) }
func GetIntSlice(key string) []int                 { return Default().GetIntSlice(key) }
func GetStringSlice(key string) []string           { return Default().GetStringSlice(key) }
func GetStringMap(key string) map[string]any       { return Default().GetStringMap(key) }
func GetStringMapString(key string) map[string]string {
	return Default().GetStringMapString(key)
}
func GetStringMapStringSlice(key string) map[string][]string {
	return Default().GetStringMapStringSlice(key)
}
func GetSizeInBytes(key string) uint   { return Default().GetSizeInBytes(key) }
func Set(key string, value any)        { Default().Set(key, value) }
func SetDefault(key string, value any) { Default().SetDefault(key, value) }
func UnmarshalKey(key string, rawVal any, opts ...viper.DecoderConfigOption) error {
	return Default().UnmarshalKey(key, rawVal, opts...)
}
func Unmarshal(rawVal any, opts ...viper.DecoderConfigOption) error {
	return Default().Unmarshal(rawVal, opts...)
}
func UnmarshalExact(rawVal any, opts ...viper.DecoderConfigOption) error {
	return Default().UnmarshalExact(rawVal, opts...)
}

func Reset() {
	viper.Reset()
	Parse()
}
