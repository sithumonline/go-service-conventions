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

package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"example.com/widgetsvc/models"
)

//go:generate moq -rm -fmt goimports -skip-ensure -pkg repomocks -out repomocks/widget_repository_mock.go . WidgetRepository:WidgetRepositoryMock
type WidgetRepository interface {
	// CreateWidget inserts w. It returns created=false, without error, when a
	// widget with the same org and name already exists.
	CreateWidget(ctx context.Context, w *models.Widget) (created bool, err error)
	GetWidget(ctx context.Context, orgID, name string) (*models.Widget, error)
	ListWidgets(ctx context.Context, orgID string, limit, offset int) ([]models.Widget, int64, error)
	DeleteWidget(ctx context.Context, orgID, name string) error
}

type widgetRepo struct {
	db *gorm.DB
}

func NewWidgetRepo(db *gorm.DB) WidgetRepository {
	return &widgetRepo{db: db}
}

// CreateWidget uses ON CONFLICT DO NOTHING rather than check-then-insert, so
// two concurrent creates cannot both succeed.
func (r *widgetRepo) CreateWidget(ctx context.Context, w *models.Widget) (bool, error) {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	if w.Labels == nil {
		w.Labels = map[string]string{}
	}
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(w)
	return result.RowsAffected > 0, result.Error
}

func (r *widgetRepo) GetWidget(ctx context.Context, orgID, name string) (*models.Widget, error) {
	var w models.Widget
	if err := r.db.WithContext(ctx).
		Where("org_id = ? AND name = ?", orgID, name).
		First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *widgetRepo) ListWidgets(ctx context.Context, orgID string, limit, offset int) ([]models.Widget, int64, error) {
	var widgets []models.Widget
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Widget{}).Where("org_id = ?", orgID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	result := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&widgets)
	return widgets, total, result.Error
}

func (r *widgetRepo) DeleteWidget(ctx context.Context, orgID, name string) error {
	result := r.db.WithContext(ctx).
		Where("org_id = ? AND name = ?", orgID, name).
		Delete(&models.Widget{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
