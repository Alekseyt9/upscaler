package s3stor

import (
	"flag"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	accessKeyID     string
	secretAccessKey string
)

var sflag = flag.String("secrets", "", "Access Key ID and Secret Access Key separated by ';'")

func TestYOStorage_GetPresigned(t *testing.T) {
	keys := *sflag
	if keys != "" {
		keyParts := strings.Split(keys, ";")
		if len(keyParts) == 2 {
			accessKeyID = keyParts[0]
			secretAccessKey = keyParts[1]
		}
	}

	if accessKeyID == "" || secretAccessKey == "" {
		t.Fatal("AccessKeyID and SecretAccessKey must be provided")
	}

	storage, err := New(YOKeys{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
	})
	require.NoError(t, err, "Failed to initialize storage")

	count := 5
	objects, err := storage.GetPresigned(count)
	require.NoError(t, err, "Failed to generate presigned URLs")

	assert.Equal(t, count, len(objects), "Expected number of presigned URLs does not match")

	for _, obj := range objects {
		t.Log("url", obj.Url)
		assert.NotEmpty(t, obj.Url, "Presigned URL should not be empty")
		assert.NotEmpty(t, obj.Key, "Key should not be empty")
	}
}
