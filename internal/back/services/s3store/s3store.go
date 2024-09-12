package s3store

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const (
	bucketName = "upscaler" // todo move
)

type S3Store interface {
	GetPresigned(count int) ([]Link, error)
}

type YOStorage struct {
	client *s3.Client
}

type YOKeys struct {
	AccessKeyID     string
	SecretAccessKey string
}

type Link struct {
	Url string
	Key string
}

func New(keys YOKeys) (*YOStorage, error) {
	c, err := initStorage(keys)
	if err != nil {
		return nil, err
	}
	inst := YOStorage{
		client: c,
	}

	return &inst, err
}

func (s *YOStorage) GetPresigned(count int) ([]Link, error) {
	var objects []Link
	presignClient := s3.NewPresignClient(s.client)

	for i := 0; i < count; i++ {
		key := uuid.New().String()

		req, err := presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
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

func initStorage(keys YOKeys) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("ru-central1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(keys.AccessKeyID, keys.SecretAccessKey, ""),
		),
	)

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.EndpointResolverV2 = &CustomS3EndpointResolverV2{}
	})

	/*
		result, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
		if err != nil {
			return nil, err
		}

		for _, bucket := range result.Buckets {
			log.Printf("bucket=%s creation time=%s", aws.ToString(bucket.Name), bucket.CreationDate.Format("2006-01-02 15:04:05 Monday"))
		}
	*/

	return client, nil
}
