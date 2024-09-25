package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Alekseyt9/upscaler/internal/back/config"
	"github.com/Alekseyt9/upscaler/internal/back/handler/middleware/jwtcheker"
	"github.com/Alekseyt9/upscaler/internal/back/model"
	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/Alekseyt9/upscaler/internal/common/services/s3store"
	"github.com/gorilla/websocket"
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

	cookies, err := Login(t)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ServerHandler{
				s3:  s3s,
				log: slog.New(slog.NewTextHandler(os.Stdout, nil)),
			}

			handlerWithJWT := jwtcheker.WithJWTCheck(http.HandlerFunc(h.GetUploadURLs), "mysecret", h.log)
			req := httptest.NewRequest("GET", "/api/user/getuploadurls/?"+tt.queryParam, nil)
			for _, cookie := range cookies {
				req.AddCookie(cookie)
			}
			w := httptest.NewRecorder()
			handlerWithJWT.ServeHTTP(w, req)

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
	}

	cookies, err := Login(t)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ServerHandler{
				us:  &MockUserService{},
				log: slog.New(slog.NewTextHandler(os.Stdout, nil)),
			}

			handlerWithJWT := jwtcheker.WithJWTCheck(http.HandlerFunc(h.CompleteFilesUpload), "mysecret", h.log)
			req := httptest.NewRequest("POST", "/api/user/completefilesupload", strings.NewReader(tt.body))
			for _, cookie := range cookies {
				req.AddCookie(cookie)
			}
			w := httptest.NewRecorder()

			handlerWithJWT.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}

func TestServerHandler_GetState(t *testing.T) {
	tests := []struct {
		name         string
		expectedCode int
	}{
		{
			name:         "valid request",
			expectedCode: http.StatusOK,
		},
	}

	cookies, err := Login(t)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ServerHandler{
				store: &store.MemoryStore{},
				log:   slog.New(slog.NewTextHandler(os.Stdout, nil)),
			}

			handlerWithJWT := jwtcheker.WithJWTCheck(http.HandlerFunc(h.GetState), "mysecret", h.log)
			req := httptest.NewRequest("GET", "/api/user/getstate", nil)
			for _, cookie := range cookies {
				req.AddCookie(cookie)
			}
			w := httptest.NewRecorder()
			handlerWithJWT.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedCode == http.StatusOK {
				var result []model.ClientUserItem
				err := json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
			}
		})
	}
}

func Login(t *testing.T) ([]*http.Cookie, error) {
	handlerOptions := HandlerOptions{
		JWTSecret: "mysecret",
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	memStore := store.NewMemoryStore()

	h := ServerHandler{
		store: memStore,
		ws:    &MockWebSocketService{},
		opt:   handlerOptions,
		log:   log,
	}

	server := httptest.NewServer(http.HandlerFunc(h.Login2))
	defer server.Close()

	url := "ws" + server.URL[len("http"):] + "/login2"
	dialer := websocket.Dialer{}

	header := http.Header{}
	conn, resp, err := dialer.Dial(url, header)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		return nil, fmt.Errorf("unexpected status code: got %v want %v", resp.StatusCode, http.StatusSwitchingProtocols)
	}

	cookies := resp.Cookies()

	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return nil, fmt.Errorf("error sending close message: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	return cookies, nil
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
