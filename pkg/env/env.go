package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func MustGet(key string) string {
	val := os.Getenv(key)
	if val == "" && key != "PORT" {
		panic("env key " + val + "cannot found")
	}
	return val
}

func CheckDotEnv(path string) {
	err := godotenv.Load(path)
	if err != nil {
		log.Fatalln("Error in loading .env file:", err)
	}
}
