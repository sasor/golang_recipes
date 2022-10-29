package envs

import (
	"github.com/joho/godotenv"
	"os"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err.Error())
	}
}

func Get(envVariable string, defaultValue string) string {
	value, ok := os.LookupEnv(envVariable)
	if ok {
		return value
	}

	return defaultValue
}
