package main

import (
	"context"
	"log"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/spf13/pflag"
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
	var keys string
	pflag.StringVarP(&keys, "secrets", "s", "", "Access Key ID and Secret Access Key separated by ';'")
	pflag.Parse()
	if keys == "" {
		log.Fatal("Keys must be provided")
	}
	keyParts := strings.Split(keys, ";")
	if len(keyParts) != 2 {
		log.Fatal("Invalid format for keys, expected format: 'AccessKeyID;SecretAccessKey'")
	}
	accessKeyID := keyParts[0]
	secretAccessKey := keyParts[1]

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("ru-central1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		),
	)

	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.EndpointResolverV2 = &CustomS3EndpointResolverV2{}
	})

	result, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatal(err)
	}

	for _, bucket := range result.Buckets {
		log.Printf("bucket=%s creation time=%s", aws.ToString(bucket.Name), bucket.CreationDate.Format("2006-01-02 15:04:05 Monday"))
	}
}
