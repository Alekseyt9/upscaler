package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alekseyt9/upscaler/internal/front/services/s3stor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		mockResponse []s3stor.Link
		mockError    error
		expectedCode int
		expectedBody []string
	}{
		{
			name:       "valid request with one link",
			queryParam: "count=1",
			mockResponse: []s3stor.Link{
				{Url: "http://example.com/1"},
			},
			expectedCode: http.StatusOK,
			expectedBody: []string{"http://example.com/1"},
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
		{
			name:         "error from GetPresigned",
			queryParam:   "count=1",
			mockError:    assert.AnError,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := new(MockS3Service)
			if tt.mockResponse != nil || tt.mockError != nil {
				mockS3.On("GetPresigned", mock.AnythingOfType("int")).Return(tt.mockResponse, tt.mockError)
			}

			h := &FrontHandler{s3: mockS3}
			req := httptest.NewRequest("GET", "/?"+tt.queryParam, nil)
			w := httptest.NewRecorder()

			h.GetPresignedURLs(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedCode == http.StatusOK {
				var result []string
				err := json.NewDecoder(resp.Body).Decode(&result)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, result)
			}
		})
	}
}
