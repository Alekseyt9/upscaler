package s3store

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"os"
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

func TestYOStorage_GetPresignedLoad(t *testing.T) {
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

	// First, generate a presigned URL to upload a file
	objects, err := storage.GetPresigned(1)
	require.NoError(t, err, "Failed to generate presigned URLs")
	assert.Equal(t, 1, len(objects), "Expected number of presigned URLs does not match")

	// Upload the file using the generated URL
	url := objects[0].Url
	key := objects[0].Key
	testData := []byte("This is a test content for uploading to Yandex Object Storage.")

	req, err := http.NewRequest("PUT", url, bytes.NewReader(testData))
	require.NoError(t, err, "Failed to create new HTTP request")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to perform HTTP request")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected HTTP status 200 OK")

	// Now test GetPresignedLoad by generating a presigned download URL for the same file
	downloadURL, err := storage.GetPresignedLoad(key)
	require.NoError(t, err, "Failed to generate presigned download URL")

	// Check that the URL works by downloading the file
	resp, err = http.Get(downloadURL)
	require.NoError(t, err, "Failed to download file")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected HTTP status 200 OK")
}

func TestYOStorage_DownloadAndSaveTemp(t *testing.T) {
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

	// First, generate a presigned URL to upload a file
	objects, err := storage.GetPresigned(1)
	require.NoError(t, err, "Failed to generate presigned URLs")
	assert.Equal(t, 1, len(objects), "Expected number of presigned URLs does not match")

	// Upload the file using the generated URL
	url := objects[0].Url
	key := objects[0].Key
	testData := []byte("This is a test content for uploading to Yandex Object Storage.")

	req, err := http.NewRequest("PUT", url, bytes.NewReader(testData))
	require.NoError(t, err, "Failed to create new HTTP request")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to perform HTTP request")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected HTTP status 200 OK")

	// Test DownloadAndSaveTemp method
	downloadURL, err := storage.GetPresignedLoad(key)
	require.NoError(t, err, "Failed to generate presigned download URL")

	tempFilePath, err := storage.DownloadAndSaveTemp(downloadURL, ".txt")
	require.NoError(t, err, "Failed to download and save temporary file")

	assert.FileExists(t, tempFilePath, "Temporary file should exist")
}

func TestYOStorage_Upload(t *testing.T) {
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

	// Generate a presigned URL for uploading a file
	objects, err := storage.GetPresigned(1)
	require.NoError(t, err, "Failed to generate presigned URLs")

	url := objects[0].Url
	filePath := "./test_file.txt"

	// Create a temporary file to upload
	tempFile, err := os.Create(filePath)
	require.NoError(t, err, "Failed to create temp file")

	_, err = tempFile.WriteString("This is a test content for uploading to Yandex Object Storage.")
	require.NoError(t, err, "Failed to write to temp file")
	tempFile.Close()

	// Test Upload method
	err = storage.Upload(url, filePath)
	require.NoError(t, err, "Failed to upload file")

	// Clean up temporary file
	err = os.Remove(filePath)
	require.NoError(t, err, "Failed to remove temp file")
}
