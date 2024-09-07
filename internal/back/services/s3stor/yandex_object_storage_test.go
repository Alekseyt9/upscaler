package s3stor

import (
	"testing"

	"github.com/Alekseyt9/upscaler/internal/back/config"
	"github.com/Alekseyt9/upscaler/internal/common/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYOStorage_GetPresigned(t *testing.T) {
	err := testutils.LoadEnv()
	require.NoError(t, err, "Failed to load env")

	cfg, err := config.LoadConfig()
	require.NoError(t, err, "Failed to load config")

	if cfg.S3AccessKeyID == "" || cfg.S3SecretAccessKey == "" {
		t.Fatal("AccessKeyID and SecretAccessKey must be provided")
	}

	storage, err := New(YOKeys{
		AccessKeyID:     cfg.S3AccessKeyID,
		SecretAccessKey: cfg.S3SecretAccessKey,
	})
	require.NoError(t, err, "Failed to initialize storage")

	count := 5
	objects, err := storage.GetPresigned(count)
	require.NoError(t, err, "Failed to generate presigned URLs")

	assert.Equal(t, count, len(objects), "Expected number of presigned URLs does not match")

	for _, obj := range objects {
		//t.Log("url", obj.Url)
		assert.NotEmpty(t, obj.Url, "Presigned URL should not be empty")
		assert.NotEmpty(t, obj.Key, "Key should not be empty")
	}
}
