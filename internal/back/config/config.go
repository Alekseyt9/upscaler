package config

type Config struct {
	BackAddress       string `env:"BACK_ADDRESS"`
	PgDataBaseDSN     string `env:"DATABASE_DSN"`
	S3AccessKeyID     string `env:"S3_ACCESSKEYID"`
	S3SecretAccessKey string `env:"S3_SECRETACCESSKEY"`
	S3BucketName      string `env:"S3_BUCKETNAME"`
	JWTSecret         string `env:"WT_SECRET"`
}
