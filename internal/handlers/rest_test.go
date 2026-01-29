package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRestSaveService struct {
	mock.Mock
}

func (m *MockRestSaveService) CreatePromoWithAudit(ctx context.Context, modelToRepo model.PromoCode, auditLog audit.Log) error {
	args := m.Called(ctx, modelToRepo, auditLog)
	return args.Error(0)
}

func TestOneTimePromoHandler_GeneratePromo(t *testing.T) {
	type testCase struct {
		name       string
		method     string
		body       string
		setupMock  func(*MockRestSaveService)
		wantStatus int
		wantBody   string
	}

	tests := []testCase{
		{
			name:   "Error: Wrong HTTP Method",
			method: http.MethodGet,
			body:   "",
			setupMock: func(m *MockRestSaveService) {
			},
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "Method not allowed",
		},
		{
			name:   "Error: Invalid JSON Syntax",
			method: http.MethodPost,
			body:   `{"code": "TEST", "bonus_length": 10`,
			setupMock: func(m *MockRestSaveService) {
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "unexpected EOF",
		},
		{
			name:   "Error: Validation (Empty Code)",
			method: http.MethodPost,
			body:   `{"code": "", "bonus_length": 10, "capacity": 1}`,
			setupMock: func(m *MockRestSaveService) {
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Validation failed",
		},
		{
			name:   "Error: Validation (Negative Capacity)",
			method: http.MethodPost,
			body:   `{"code": "TEST", "bonus_length": 10, "capacity": -5}`,
			setupMock: func(m *MockRestSaveService) {
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Validation failed",
		},
		{
			name:   "Error: Service Failure (DB Error)",
			method: http.MethodPost,
			body:   `{"code": "DB_FAIL", "bonus_length": 10, "capacity": 1}`,
			setupMock: func(m *MockRestSaveService) {
				m.On("CreatePromoWithAudit",
					mock.Anything,
					mock.MatchedBy(func(p model.PromoCode) bool {
						return p.Code == "DB_FAIL" && p.BonusLength == 10
					}),
					mock.MatchedBy(func(a audit.Log) bool {
						return a.Code == "DB_FAIL" && a.Action == "create"
					}),
				).Return(errors.New("database connection lost"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Promo creation failed",
		},
		{
			name:   "Success: Promo Created",
			method: http.MethodPost,
			body:   `{"code": "SUCCESS", "bonus_length": 5, "capacity": 100}`,
			setupMock: func(m *MockRestSaveService) {
				m.On("CreatePromoWithAudit",
					mock.Anything,
					mock.MatchedBy(func(p model.PromoCode) bool {
						return p.Code == "SUCCESS" && p.BonusLength == 5 && p.Capacity == 100
					}),
					mock.MatchedBy(func(a audit.Log) bool {
						return a.Code == "SUCCESS" && a.Action == "create" && a.By == "auto"
					}),
				).Return(nil)
			},
			wantStatus: http.StatusCreated,
			wantBody:   `{"code":"SUCCESS","status":"ok"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 3. Initialization updated
			mockService := new(MockRestSaveService)

			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			h := NewOneTimePromoHandler(mockService)
			handlerFunc := h.GeneratePromo()

			req := httptest.NewRequest(tt.method, "/promo/generate", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handlerFunc(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, "Status code should match expected value")
			require.NotEmpty(t, w.Body.String(), "Response body should not be empty")
			assert.Contains(t, w.Body.String(), tt.wantBody, "Body should contain expected content")

			mockService.AssertExpectations(t)
		})
	}
}
