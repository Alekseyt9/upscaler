package s3store

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
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
