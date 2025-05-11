package util

import "os"

func AppEnv() string {
	return os.Getenv("ENV")
}

func IsEnvProd() bool {
	return AppEnv() == "production"
}
