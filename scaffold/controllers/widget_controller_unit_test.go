// Copyright (c) 2026, Example Org.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"example.com/widgetsvc/middleware"
	"example.com/widgetsvc/spec"
	"example.com/widgetsvc/utils"
)

// stubWidgetService is a hand-written func-field stub, the shape to use for
// interfaces that have no generated mock.
type stubWidgetService struct {
	CreateWidgetFunc func(ctx context.Context, orgID string, req *spec.CreateWidgetRequest) (*spec.WidgetResponse, error)
	GetWidgetFunc    func(ctx context.Context, orgID, name string) (*spec.WidgetResponse, error)
	ListWidgetsFunc  func(ctx context.Context, orgID string, limit, offset int) (*spec.WidgetListResponse, error)
	DeleteWidgetFunc func(ctx context.Context, orgID, name string) error
}

func (s *stubWidgetService) CreateWidget(ctx context.Context, orgID string, req *spec.CreateWidgetRequest) (*spec.WidgetResponse, error) {
	return s.CreateWidgetFunc(ctx, orgID, req)
}

func (s *stubWidgetService) GetWidget(ctx context.Context, orgID, name string) (*spec.WidgetResponse, error) {
	return s.GetWidgetFunc(ctx, orgID, name)
}

func (s *stubWidgetService) ListWidgets(ctx context.Context, orgID string, limit, offset int) (*spec.WidgetListResponse, error) {
	return s.ListWidgetsFunc(ctx, orgID, limit, offset)
}

func (s *stubWidgetService) DeleteWidget(ctx context.Context, orgID, name string) error {
	return s.DeleteWidgetFunc(ctx, orgID, name)
}

func newRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return req.WithContext(middleware.WithClaims(req.Context(), &middleware.Claims{OrgID: "org-1", OrgName: "acme"}))
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) spec.ErrorResponse {
	t.Helper()
	var body spec.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	return body
}

func TestWidgetController_GetWidget(t *testing.T) {
	t.Run("answers 404 with a stable code for ErrWidgetNotFound", func(t *testing.T) {
		svc := &stubWidgetService{GetWidgetFunc: func(_ context.Context, _, _ string) (*spec.WidgetResponse, error) {
			return nil, utils.ErrWidgetNotFound
		}}
		req := newRequest(http.MethodGet, "/orgs/acme/widgets/w1", "")
		req.SetPathValue(utils.PathParamWidgetName, "w1")
		rec := httptest.NewRecorder()

		NewWidgetController(svc).GetWidget(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, utils.ErrCodeWidgetNotFound, decodeError(t, rec).Code)
	})

	t.Run("hides internal error detail behind the fallback message", func(t *testing.T) {
		svc := &stubWidgetService{GetWidgetFunc: func(_ context.Context, _, _ string) (*spec.WidgetResponse, error) {
			return nil, errors.New("dial tcp 10.0.0.5:5432: connection refused")
		}}
		rec := httptest.NewRecorder()

		NewWidgetController(svc).GetWidget(rec, newRequest(http.MethodGet, "/orgs/acme/widgets/w1", ""))

		body := decodeError(t, rec)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "Failed to get widget", body.Message)
		assert.Nil(t, body.Reason)
	})
}

func TestWidgetController_CreateWidget(t *testing.T) {
	t.Run("rejects an empty name before calling the service", func(t *testing.T) {
		rec := httptest.NewRecorder()
		NewWidgetController(&stubWidgetService{}).CreateWidget(rec, newRequest(http.MethodPost, "/orgs/acme/widgets", `{"name":""}`))

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, utils.ErrCodeValidation, decodeError(t, rec).Code)
	})
}

func TestWidgetController_ListWidgets(t *testing.T) {
	t.Run("rejects a limit above the maximum", func(t *testing.T) {
		rec := httptest.NewRecorder()
		NewWidgetController(&stubWidgetService{}).ListWidgets(rec, newRequest(http.MethodGet, "/orgs/acme/widgets?limit=1000", ""))

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
