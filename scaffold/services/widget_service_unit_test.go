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

// Reference service unit test: inject generated mocks through the real
// constructor, drive each branch, assert on returned values and sentinels.
// An unconfigured mock method panics, so an unexpected call fails loudly.
package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"example.com/widgetsvc/models"
	"example.com/widgetsvc/repositories/repomocks"
	"example.com/widgetsvc/spec"
	"example.com/widgetsvc/utils"
)

func TestWidgetService_GetWidget(t *testing.T) {
	const org, name = "org-1", "w1"

	t.Run("maps record-not-found to ErrWidgetNotFound", func(t *testing.T) {
		repo := &repomocks.WidgetRepositoryMock{
			GetWidgetFunc: func(_ context.Context, _, _ string) (*models.Widget, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}
		_, err := NewWidgetService(repo).GetWidget(context.Background(), org, name)
		assert.ErrorIs(t, err, utils.ErrWidgetNotFound)
	})

	t.Run("does not mask a real error as not-found", func(t *testing.T) {
		boom := errors.New("connection reset")
		repo := &repomocks.WidgetRepositoryMock{
			GetWidgetFunc: func(_ context.Context, _, _ string) (*models.Widget, error) { return nil, boom },
		}
		_, err := NewWidgetService(repo).GetWidget(context.Background(), org, name)
		assert.ErrorIs(t, err, boom)
		assert.NotErrorIs(t, err, utils.ErrWidgetNotFound)
	})

	t.Run("scopes the lookup to the caller's org", func(t *testing.T) {
		repo := &repomocks.WidgetRepositoryMock{
			GetWidgetFunc: func(_ context.Context, orgID, n string) (*models.Widget, error) {
				return &models.Widget{OrgID: orgID, Name: n}, nil
			},
		}
		got, err := NewWidgetService(repo).GetWidget(context.Background(), org, name)
		require.NoError(t, err)
		assert.Equal(t, name, got.Name)
		assert.Equal(t, org, repo.GetWidgetCalls()[0].OrgID)
	})
}

func TestWidgetService_CreateWidget(t *testing.T) {
	t.Run("reports a duplicate as ErrWidgetAlreadyExists", func(t *testing.T) {
		repo := &repomocks.WidgetRepositoryMock{
			CreateWidgetFunc: func(_ context.Context, _ *models.Widget) (bool, error) { return false, nil },
		}
		_, err := NewWidgetService(repo).CreateWidget(context.Background(), "org-1", &spec.CreateWidgetRequest{Name: "w1"})
		assert.ErrorIs(t, err, utils.ErrWidgetAlreadyExists)
	})

	t.Run("rejects a missing org without touching the repository", func(t *testing.T) {
		repo := &repomocks.WidgetRepositoryMock{} // CreateWidgetFunc nil: a call panics
		_, err := NewWidgetService(repo).CreateWidget(context.Background(), "", &spec.CreateWidgetRequest{Name: "w1"})
		assert.ErrorIs(t, err, utils.ErrInvalidInput)
	})

	t.Run("defaults display name to name", func(t *testing.T) {
		repo := &repomocks.WidgetRepositoryMock{
			CreateWidgetFunc: func(_ context.Context, _ *models.Widget) (bool, error) { return true, nil },
		}
		got, err := NewWidgetService(repo).CreateWidget(context.Background(), "org-1", &spec.CreateWidgetRequest{Name: "w1"})
		require.NoError(t, err)
		assert.Equal(t, "w1", got.DisplayName)
	})
}

func TestWidgetService_ListWidgets(t *testing.T) {
	t.Run("returns an empty slice, not nil, for no rows", func(t *testing.T) {
		repo := &repomocks.WidgetRepositoryMock{
			ListWidgetsFunc: func(_ context.Context, _ string, _, _ int) ([]models.Widget, int64, error) {
				return []models.Widget{}, 0, nil
			},
		}
		got, err := NewWidgetService(repo).ListWidgets(context.Background(), "org-1", 20, 0)
		require.NoError(t, err)
		assert.NotNil(t, got.Widgets)
	})
}
