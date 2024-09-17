package main

import (
	"context"
	"log"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	smithyendpoints "github.com/aws/smithy-go/endpoints"

	cfg "github.com/Alekseyt9/upscaler/internal/back/config"
	"github.com/Alekseyt9/upscaler/internal/common/services/s3store"
)

type CustomS3EndpointResolverV2 struct{}

func (r *CustomS3EndpointResolverV2) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (smithyendpoints.Endpoint, error) {
	if *params.Region == "ru-central1" {
		uri, err := url.Parse("https://storage.yandexcloud.net")
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}

		return smithyendpoints.Endpoint{
			URI:        *uri,
			Headers:    nil,
			Properties: smithy.Properties{},
		}, nil
	}

	return smithyendpoints.Endpoint{}, &aws.EndpointNotFoundError{}
}

func main() {
	c, err := cfg.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if c.S3AccessKeyID == "" || c.S3SecretAccessKey == "" {
		log.Fatal("AccessKeyID and SecretAccessKey must be provided")
	}

	log.Println("c.S3BucketName", c.S3BucketName)

	s3, err := s3store.New(s3store.S3Options{
		AccessKeyID:     c.S3AccessKeyID,
		SecretAccessKey: c.S3SecretAccessKey,
		BucketName:      c.S3BucketName,
	})
	if err != nil {
		log.Panicln(err)
	}

	links, err := s3.GetPresigned(1)
	if err != nil {
		log.Panicln(err)
	}

	log.Println(links[0].Url)
}
