package s3stor

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type YOStorage struct {
	client *s3.Client
}

type YOKeys struct {
	AccessKeyID     string
	SecretAccessKey string
}

var (
	instance YOStorage
	once     sync.Once
	err      error
)

func Get(keys YOKeys) (*YOStorage, error) {
	once.Do(func() {
		var c *s3.Client
		c, err = initStorage(keys)
		if err != nil {
			return
		}
		instance = YOStorage{
			client: c,
		}
	})
	return &instance, err
}

func initStorage(keys YOKeys) (*s3.Client, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				PartitionID: "yc",
				URL:         "https://storage.yandexcloud.net",
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	})

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(
			&credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID:     keys.AccessKeyID,
					SecretAccessKey: keys.SecretAccessKey,
				},
			},
		),
		config.WithRegion("auto"),
	)

	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}
