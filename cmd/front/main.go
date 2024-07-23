package main

import (
	"github.com/Alekseyt9/upscaler/internal/front/config"
	"github.com/Alekseyt9/upscaler/internal/front/run"
)

func main() {
	cfg := &config.Config{}
	ParseFlags(cfg)
	SetEnv(cfg)

	err := run.Run(cfg)
	if err != nil {
		panic("Ошибка запуска сервера: " + err.Error())
	}
}
