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

package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"example.com/widgetsvc/models"
	"example.com/widgetsvc/repositories"
	"example.com/widgetsvc/spec"
	"example.com/widgetsvc/utils"
)

type WidgetService interface {
	CreateWidget(ctx context.Context, orgID string, req *spec.CreateWidgetRequest) (*spec.WidgetResponse, error)
	GetWidget(ctx context.Context, orgID, name string) (*spec.WidgetResponse, error)
	ListWidgets(ctx context.Context, orgID string, limit, offset int) (*spec.WidgetListResponse, error)
	DeleteWidget(ctx context.Context, orgID, name string) error
}

type widgetService struct {
	repo repositories.WidgetRepository
	now  func() time.Time
}

func NewWidgetService(repo repositories.WidgetRepository) WidgetService {
	return &widgetService{repo: repo, now: time.Now}
}

// CreateWidget creates a widget, defaulting the display name to the name.
func (s *widgetService) CreateWidget(ctx context.Context, orgID string, req *spec.CreateWidgetRequest) (*spec.WidgetResponse, error) {
	if orgID == "" {
		return nil, fmt.Errorf("create widget: missing org: %w", utils.ErrInvalidInput)
	}
	now := s.now()
	w := &models.Widget{
		ID:          uuid.New(),
		OrgID:       orgID,
		Name:        req.Name,
		DisplayName: req.Name,
		Labels:      map[string]string{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if req.DisplayName != nil && *req.DisplayName != "" {
		w.DisplayName = *req.DisplayName
	}
	if req.Labels != nil {
		w.Labels = *req.Labels
	}

	created, err := s.repo.CreateWidget(ctx, w)
	if err != nil {
		return nil, fmt.Errorf("failed to create widget: %w", err)
	}
	if !created {
		return nil, utils.ErrWidgetAlreadyExists
	}
	return toWidgetResponse(w), nil
}

// GetWidget returns one widget in the caller's org.
func (s *widgetService) GetWidget(ctx context.Context, orgID, name string) (*spec.WidgetResponse, error) {
	w, err := s.repo.GetWidget(ctx, orgID, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrWidgetNotFound
		}
		return nil, fmt.Errorf("failed to get widget: %w", err)
	}
	return toWidgetResponse(w), nil
}

// ListWidgets returns one page of the org's widgets, newest first.
func (s *widgetService) ListWidgets(ctx context.Context, orgID string, limit, offset int) (*spec.WidgetListResponse, error) {
	widgets, total, err := s.repo.ListWidgets(ctx, orgID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list widgets: %w", err)
	}
	items := make([]spec.WidgetResponse, len(widgets))
	for i := range widgets {
		items[i] = *toWidgetResponse(&widgets[i])
	}
	return &spec.WidgetListResponse{Widgets: items, Total: total, Limit: limit, Offset: offset}, nil
}

// DeleteWidget deletes one widget in the caller's org.
func (s *widgetService) DeleteWidget(ctx context.Context, orgID, name string) error {
	err := s.repo.DeleteWidget(ctx, orgID, name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.ErrWidgetNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to delete widget: %w", err)
	}
	return nil
}

func toWidgetResponse(w *models.Widget) *spec.WidgetResponse {
	labels := w.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	return &spec.WidgetResponse{
		Name:        w.Name,
		DisplayName: w.DisplayName,
		Labels:      labels,
		CreatedAt:   w.CreatedAt,
	}
}
