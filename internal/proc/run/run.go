package run

import (
	"log/slog"
	"os"
	"time"

	"github.com/Alekseyt9/upscaler/internal/common/services/s3store"
	"github.com/Alekseyt9/upscaler/internal/proc/config"
	"github.com/Alekseyt9/upscaler/internal/proc/services/consumer"
	"github.com/Alekseyt9/upscaler/internal/proc/services/fileprocessor"
	"github.com/Alekseyt9/upscaler/internal/proc/services/idcheck"
	"github.com/Alekseyt9/upscaler/internal/proc/services/processor"
	"github.com/Alekseyt9/upscaler/internal/proc/services/producer"
	"github.com/Alekseyt9/upscaler/pkg/workerpool"
)

func Run(cfg *config.Config) error {
	s3, err := s3store.New(s3store.S3Options{
		AccessKeyID:     cfg.S3AccessKeyID,
		SecretAccessKey: cfg.S3SecretAccessKey,
		BucketName:      cfg.S3BucketName,
	})
	if err != nil {
		return err
	}

	wp := workerpool.New(1) // features of running the utility for upscale, there is no point in running it in parallel
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	fproc := fileprocessor.NewFileProcessor("")
	idcheck := idcheck.NewIdCheckService(cfg.RedisAddr, 24*time.Hour)
	producer, err := producer.NewProducer([]string{cfg.KafkaAddr}, cfg.KafkaTopicResult)
	if err != nil {
		return err
	}

	proc := processor.New(wp, s3, log, fproc, idcheck, producer)
	_, err = consumer.NewConsumer([]string{cfg.KafkaAddr}, cfg.KafkaTopic, cfg.KafkeCunsumerGroup, proc)
	if err != nil {
		return err
	}

	return nil
}
