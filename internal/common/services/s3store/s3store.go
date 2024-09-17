package s3store

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Store interface {
	GetPresigned(count int) ([]Link, error)
	DownloadAndSaveTemp(url string) (string, error)
	Upload(url string, path string) error
}

type YOStorage struct {
	client *s3.Client
	opts   S3Options
}

type S3Options struct {
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}

type Link struct {
	Url string
	Key string
}

func New(opt S3Options) (*YOStorage, error) {
	c, err := initStorage(opt)
	if err != nil {
		return nil, err
	}
	inst := YOStorage{
		client: c,
		opts:   opt,
	}

	return &inst, err
}

func initStorage(opts S3Options) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("ru-central1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(opts.AccessKeyID, opts.SecretAccessKey, "")),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           "https://storage.yandexcloud.net",
					SigningRegion: region,
				}, nil
			})),
	)
	if err != nil {
		panic("Ошибка загрузки конфигурации: " + err.Error())
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return client, nil
}

func (s *YOStorage) GetPresigned(count int) ([]Link, error) {
	var objects []Link
	presignClient := s3.NewPresignClient(s.client)

	for i := 0; i < count; i++ {
		key := uuid.New().String()

		req, err := presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(s.opts.BucketName),
			Key:    aws.String(key),
		}, s3.WithPresignExpires(time.Hour*24*30))

		if err != nil {
			return nil, err
		}

		objects = append(objects, Link{
			Url: req.URL,
			Key: key,
		})
	}

	return objects, nil
}

func (s *YOStorage) DownloadAndSaveTemp(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("ошибка при скачивании файла: %w", err)
	}
	defer resp.Body.Close()

	tempFile, err := os.CreateTemp("", "downloaded-*.tmp")
	if err != nil {
		return "", fmt.Errorf("ошибка создания временного файла: %w", err)
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка записи во временный файл: %w", err)
	}

	return tempFile.Name(), nil
}

func (s *YOStorage) Upload(url string, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	req, err := http.NewRequest(http.MethodPut, url, file)
	if err != nil {
		return fmt.Errorf("ошибка создания HTTP-запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка отправки HTTP-запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("неудачная загрузка файла, статус: %s", resp.Status)
	}

	return nil
}
