package main

import (
	"log"

	"github.com/Alekseyt9/upscaler/internal/back/config"
	"github.com/Alekseyt9/upscaler/internal/back/run"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Println("config load error: " + err.Error())
	}

	err = run.Run(cfg)
	if err != nil {
		log.Println("server startup error: " + err.Error())
	}
}
