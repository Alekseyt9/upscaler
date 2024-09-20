package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Alekseyt9/upscaler/internal/back/config"
	"github.com/Alekseyt9/upscaler/internal/back/model"
	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/Alekseyt9/upscaler/internal/common/services/s3store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerHandler_GetUploadURLs(t *testing.T) {
	tests := []struct {
		name         string
		queryParam   string
		expectedCode int
	}{
		{
			name:         "valid request with one link",
			queryParam:   "count=1",
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid count parameter",
			queryParam:   "count=invalid",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "count exceeds max limit",
			queryParam:   "count=11",
			expectedCode: http.StatusBadRequest,
		},
	}

	cfg, err := config.LoadConfig()
	require.NoError(t, err, "Failed to load config")
	if cfg.S3AccessKeyID == "" || cfg.S3SecretAccessKey == "" {
		t.Fatal("AccessKeyID and SecretAccessKey must be provided")
	}
	s3s, err := s3store.New(s3store.S3Options{
		AccessKeyID:     cfg.S3AccessKeyID,
		SecretAccessKey: cfg.S3SecretAccessKey,
		BucketName:      cfg.S3BucketName,
	})
	require.NoError(t, err, "Failed to load s3stor")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ServerHandler{s3: s3s}
			req := httptest.NewRequest("GET", "/?"+tt.queryParam, nil)
			w := httptest.NewRecorder()

			h.GetUploadURLs(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedCode == http.StatusOK {
				var result []s3store.Link
				err := json.NewDecoder(resp.Body).Decode(&result)
				assert.NoError(t, err)
				assert.Equal(t, 1, len(result), "Expected 1 presigned link")
			}
		})
	}
}

func TestServerHandler_CompleteFilesUpload(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		expectedCode int
		userIDValid  bool
	}{
		{
			name:         "valid request with file upload data",
			body:         `[{"FileName": "file1.png"}, {"FileName": "file2.jpg"}]`,
			expectedCode: http.StatusOK,
			userIDValid:  true,
		},
		{
			name:         "invalid JSON body",
			body:         `invalid json`,
			expectedCode: http.StatusInternalServerError,
			userIDValid:  true,
		},
		{
			name:         "missing user ID",
			body:         `[{"FileName": "file1.png"}]`,
			expectedCode: http.StatusInternalServerError,
			userIDValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ServerHandler{
				us: &MockUserService{},
			}
			req := httptest.NewRequest("POST", "/complete", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.CompleteFilesUpload(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}

func TestServerHandler_GetState(t *testing.T) {
	tests := []struct {
		name         string
		userIDValid  bool
		expectedCode int
	}{
		{
			name:         "valid request with valid user ID",
			userIDValid:  true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "missing user ID",
			userIDValid:  false,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ServerHandler{
				store: &store.MemoryStore{},
			}
			req := httptest.NewRequest("GET", "/state", nil)
			w := httptest.NewRecorder()

			h.GetState(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedCode == http.StatusOK {
				var result []model.ClientUserItem
				err := json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// MockUserService simulates user service behavior for tests.
type MockUserService struct{}

func (m *MockUserService) CreateTasks(ctx context.Context, fileInfos []model.UploadedFile, userID int64) error {
	return nil
}

// FinishTasks simulates the task completion logic.
func (m *MockUserService) FinishTasks(ctx context.Context, msgs []cmodel.BrokerMessageResult) error {
	return nil
}
