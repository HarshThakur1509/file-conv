package initializers

import "github.com/joho/godotenv"

func LoadEnv() {
	godotenv.Load()
}
