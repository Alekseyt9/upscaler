package config

type Config struct {
	Address       string `env:"ADDRESS"`
	PgDataBaseDSN string `env:"PG_DATABASE_DSN"`
}
