package main

import (
	"log"

	"github.com/Alekseyt9/upscaler/internal/proc/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Println("config load error: " + err.Error())
	}

	log.Println(cfg)

	/*
		err = run.Run(cfg)
		if err != nil {
			log.Println("server startup error: " + err.Error())
		}
	*/
}
