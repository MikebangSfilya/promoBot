package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kozalosev/goSadTgBot/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOneTimePromoHandler_GeneratePromo_Validation(t *testing.T) {
	type testCase struct {
		name       string
		method     string
		body       string
		wantStatus int
		wantBody   string
	}
	tests := []testCase{
		{
			name:       "Error: Wrong HTTP Method",
			method:     http.MethodGet,
			body:       "",
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "Method not allowed",
		},
		{
			name:       "Error: Invalid JSON Syntax (Missing Brace)",
			method:     http.MethodPost,
			body:       `{"code": "TEST", "bonus_length": 10`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "unexpected EOF",
		},
		{
			name:       "Error: Validation (Empty Code)",
			method:     http.MethodPost,
			body:       `{"code": "", "bonus_length": 10, "capacity": 1}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Validation failed",
		},
		{
			name:       "Error: Validation (Negative Capacity)",
			method:     http.MethodPost,
			body:       `{"code": "TEST", "bonus_length": 10, "capacity": -5}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Validation failed",
		},
		{
			name:       "Error: Type Mismatch (String instead of Int)",
			method:     http.MethodPost,
			body:       `{"code": "TEST", "bonus_length": "ten", "capacity": 1}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "cannot unmarshal string",
		},
	}

	// Initialize the handler with minimal dependencies.
	mockEnv := &base.ApplicationEnv{Ctx: context.Background()}

	h := NewOneTimePromoHandler(mockEnv, nil, nil, nil)
	handlerFunc := h.GeneratePromo()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fake HTTP request
			req := httptest.NewRequest(tt.method, "/promo/generate", strings.NewReader(tt.body))

			// Create a ResponseRecorder to capture the response
			w := httptest.NewRecorder()

			// Execute the handler
			handlerFunc(w, req)

			// 1. Check if the status code matches expected
			assert.Equal(t, tt.wantStatus, w.Code, "Status code should match expected value")

			// 2. Check if the response body contains the expected error message.
			// require.NotEmpty ensures we don't proceed if the body is unexpectedly empty.
			require.NotEmpty(t, w.Body.String(), "Response body should not be empty")
			assert.Contains(t, w.Body.String(), tt.wantBody, "Body should contain the expected error message")
		})
	}
}
