package run

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/Alekseyt9/upscaler/internal/common/services/s3store"
	"github.com/Alekseyt9/upscaler/internal/proc/config"
	"github.com/Alekseyt9/upscaler/internal/proc/services/consumer"
	"github.com/Alekseyt9/upscaler/internal/proc/services/fileprocessor"
	"github.com/Alekseyt9/upscaler/internal/proc/services/idcheck"
	"github.com/Alekseyt9/upscaler/internal/proc/services/processor"
	"github.com/Alekseyt9/upscaler/internal/proc/services/producer"
	"github.com/Alekseyt9/upscaler/pkg/workerpool"
)

func Run(cfg *config.Config, log *slog.Logger) error {
	s3, err := s3store.New(s3store.S3Options{
		AccessKeyID:     cfg.S3AccessKeyID,
		SecretAccessKey: cfg.S3SecretAccessKey,
		BucketName:      cfg.S3BucketName,
	})
	if err != nil {
		return fmt.Errorf("s3store.New %w", err)
	}

	wp := workerpool.New(1) // there is no point in running upscale in parallel
	fproc := fileprocessor.NewFileProcessor("")
	idcheck := idcheck.NewIdCheckService(cfg.RedisAddr, 24*time.Hour)

	producer, err := producer.NewProducer(log, model.BrokerOptions{
		Topic:        cfg.KafkaTopicResult,
		KafkaBrokers: []string{cfg.KafkaAddr},
	})
	if err != nil {
		return fmt.Errorf("producer.NewProducer %w", err)
	}

	proc := processor.New(wp, s3, log, fproc, idcheck, producer)
	cons, err := consumer.NewConsumer(proc, log, model.BrokerOptions{
		Topic:         cfg.KafkaTopic,
		KafkaBrokers:  []string{cfg.KafkaAddr},
		ConsumerGroup: cfg.KafkeCunsumerGroup,
	})
	if err != nil {
		return fmt.Errorf("consumer.NewConsumer %w", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Info("Shutting down gracefully...")

	cons.Close()
	producer.Close()

	return nil
}
