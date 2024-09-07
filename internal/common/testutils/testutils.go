package testutils

import (
	"os"
	"regexp"

	"github.com/joho/godotenv"
)

const projectDirName = "upscaler"

// LoadEnv loads env vars from .env
func LoadEnv() error {
	re := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))

	err := godotenv.Load(string(rootPath) + `/.env`)
	if err != nil {
		return err
	}
	return nil
}
