package delete_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"URL-shortener/internal/http-server/handlers/url/delete"
	"URL-shortener/internal/http-server/handlers/url/delete/mocks"
	"github.com/go-chi/chi/v5"

	"github.com/stretchr/testify/require"
)

func TestDeleteHandler(t *testing.T) {
	mockDeleter := new(mocks.URLDeleter)
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	tests := []struct {
		name           string
		alias          string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:  "successfully deletes URL",
			alias: "test-alias",
			mockSetup: func() {
				mockDeleter.On("DeleteURL", "test-alias").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"url deleted successfully"}`,
		},
		{
			name:           "alias is empty",
			alias:          "",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name:  "URL not found",
			alias: "not-found-alias",
			mockSetup: func() {
				mockDeleter.On("DeleteURL", "not-found-alias").Return(errors.New("not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to delete url"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка моков
			mockDeleter.ExpectedCalls = nil
			tt.mockSetup()

			handler := delete.New(logger, mockDeleter)
			req := httptest.NewRequest(http.MethodDelete, "/url/"+tt.alias, nil)
			w := httptest.NewRecorder()

			if tt.alias != "" {
				routeContext := chi.NewRouteContext()
				routeContext.URLParams.Add("alias", tt.alias)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext))
			}

			handler(w, req)

			res := w.Result()
			defer res.Body.Close()

			body := w.Body.String()

			require.Equal(t, tt.expectedStatus, res.StatusCode)
			require.JSONEq(t, tt.expectedBody, body)

			mockDeleter.AssertExpectations(t)
		})
	}
}
