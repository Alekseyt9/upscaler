package main

import (
	goflag "flag"

	"github.com/Alekseyt9/upscaler/internal/front/config"
	"github.com/caarlos0/env"
	flag "github.com/spf13/pflag"
)

func ParseFlags(cfg *config.Config) {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	address := flag.StringP("address", "a", "localhost:8080", "Address and port to run server")
	pg_database := flag.StringP("database_postgres", "p", "", "Postgres database connection string")

	flag.Parse()

	cfg.Address = *address
	cfg.PgDataBaseDSN = *pg_database
}

func SetEnv(cfg *config.Config) {
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}
}
