package s3store

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"testing"

	"github.com/Alekseyt9/upscaler/internal/back/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYOStorage_GetPresigned(t *testing.T) {
	cfg, err := config.LoadConfig()
	require.NoError(t, err, "Failed to load config")

	if cfg.S3AccessKeyID == "" || cfg.S3SecretAccessKey == "" {
		t.Fatal("AccessKeyID and SecretAccessKey must be provided")
	}

	storage, err := New(S3Options{
		AccessKeyID:     cfg.S3AccessKeyID,
		SecretAccessKey: cfg.S3SecretAccessKey,
		BucketName:      cfg.S3BucketName,
	})
	require.NoError(t, err, "Failed to initialize storage")

	count := 1
	objects, err := storage.GetPresigned(count)
	require.NoError(t, err, "Failed to generate presigned URLs")

	assert.Equal(t, count, len(objects), "Expected number of presigned URLs does not match")

	url := objects[0].Url
	testData := []byte("This is a test content for uploading to Yandex Object Storage.")

	req, err := http.NewRequest("PUT", url, bytes.NewReader(testData))
	require.NoError(t, err, "Failed to create new HTTP request")

	requestDump, err := httputil.DumpRequestOut(req, true)
	require.NoError(t, err, "Failed to dump request")
	t.Logf("Request:\n%s", string(requestDump))

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to perform HTTP request")
	defer resp.Body.Close()

	responseDump, err := httputil.DumpResponse(resp, true)
	require.NoError(t, err, "Failed to dump response")
	t.Logf("Response:\n%s", string(responseDump))

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected HTTP status 200 OK")
}
