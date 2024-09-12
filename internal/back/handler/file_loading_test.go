package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alekseyt9/upscaler/internal/back/config"
	"github.com/Alekseyt9/upscaler/internal/back/services/s3stor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockS3Service struct {
	mock.Mock
}

func (m *MockS3Service) GetPresigned(count int) ([]s3stor.Link, error) {
	args := m.Called(count)
	return args.Get(0).([]s3stor.Link), args.Error(1)
}

func TestFrontHandler_GetPresignedURLs(t *testing.T) {
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

	//err := envutils.LoadEnv()
	//require.NoError(t, err, "Failed to load env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err, "Failed to load config")
	if cfg.S3AccessKeyID == "" || cfg.S3SecretAccessKey == "" {
		t.Fatal("AccessKeyID and SecretAccessKey must be provided")
	}
	s3s, err := s3stor.New(s3stor.YOKeys{
		AccessKeyID:     cfg.S3AccessKeyID,
		SecretAccessKey: cfg.S3SecretAccessKey,
	})
	require.NoError(t, err, "Failed to load s3stor")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			h := &FrontHandler{s3: s3s}
			req := httptest.NewRequest("GET", "/?"+tt.queryParam, nil)
			w := httptest.NewRecorder()

			h.GetRequisites(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedCode == http.StatusOK {
				var result []s3stor.Link
				err := json.NewDecoder(resp.Body).Decode(&result)
				assert.NoError(t, err)
			}
		})
	}
}
