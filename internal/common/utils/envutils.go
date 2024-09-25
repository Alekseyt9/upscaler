package utils

import (
	"github.com/joho/godotenv"
)

const projectDirName = "upscaler"

// LoadEnv loads env vars from .env
func LoadEnv() error {
	err := godotenv.Load(GetProcDir() + `/.env`)
	if err != nil {
		return err
	}
	return nil
}
