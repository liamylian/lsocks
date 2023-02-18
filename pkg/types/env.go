package types

import (
	"os"
)

// Env 获取环境变量
func Env(name string) StrValue {
	val, _ := os.LookupEnv(name)
	return StrValue(val)
}

// EnvDefault 获取环境变量
func EnvDefault(name string, defaultValue string) StrValue {
	val, ok := os.LookupEnv(name)
	if !ok || val == "" {
		return StrValue(defaultValue)
	}

	return StrValue(val)
}
