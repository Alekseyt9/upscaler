// Package s3store provides an interface and implementation for interacting
// with an S3-compatible storage service, such as Yandex Cloud Storage.
// It allows for operations such as generating presigned URLs, uploading files,
// and downloading files to temporary storage.
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

// S3Store defines the interface for S3 storage operations.
type S3Store interface {
	GetPresigned(count int) ([]Link, error)
	GetPresignedLoad(key string) (string, error)
	DownloadAndSaveTemp(url string, ext string) (string, error)
	Upload(url string, path string) error
}

// YOStorage implements the S3Store interface using AWS S3 SDK.
type YOStorage struct {
	client *s3.Client
	opts   S3Options
}

// S3Options contains the configuration options for connecting to the S3 service.
type S3Options struct {
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}

// Link represents a presigned URL and its associated key.
type Link struct {
	Url string
	Key string
}

// New creates a new YOStorage instance with the given S3Options.
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

// initStorage initializes the S3 client with the given S3Options.
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
		panic("Error loading configuration: " + err.Error())
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return client, nil
}

// GetPresigned generates a list of presigned URLs for uploading files.
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
			return nil, fmt.Errorf("presignClient.PresignPutObject %w", err)
		}

		objects = append(objects, Link{
			Url: req.URL,
			Key: key,
		})
	}

	return objects, nil
}

// GetPresignedLoad generates a presigned URL for downloading a file from S3.
func (s *YOStorage) GetPresignedLoad(key string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	req, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.opts.BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return "", fmt.Errorf("presignClient.PresignGetObject %w", err)
	}

	return req.URL, nil
}

// DownloadAndSaveTemp downloads a file from the given URL and saves it as a temporary file.
func (s *YOStorage) DownloadAndSaveTemp(url string, ext string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error downloading file: %w", err)
	}
	defer resp.Body.Close()

	tempFile, err := os.CreateTemp("", "downloaded-*"+ext)
	if err != nil {
		return "", fmt.Errorf("error creating temporary file: %w", err)
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error writing to temporary file: %w", err)
	}

	return tempFile.Name(), nil
}

// Upload uploads a file to the specified presigned URL.
func (s *YOStorage) Upload(url string, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	req, err := http.NewRequest(http.MethodPut, url, file)
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("file upload failed, status: %s", resp.Status)
	}

	return nil
}
