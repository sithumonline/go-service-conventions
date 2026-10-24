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

//go:build integration

package repositories

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	dbmigrations "example.com/widgetsvc/db_migrations"
	"example.com/widgetsvc/models"
)

// Runs against the Postgres named by TEST_DATABASE_DSN; `make test-integration` provides one.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set")
	}
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, dbmigrations.Migrate(gdb))
	t.Cleanup(func() { gdb.Exec("DELETE FROM widgets") })
	return gdb
}

func TestWidgetRepo_CreateIsIdempotentPerOrgAndName(t *testing.T) {
	repo := NewWidgetRepo(openTestDB(t))
	ctx := context.Background()

	created, err := repo.CreateWidget(ctx, &models.Widget{OrgID: "org-1", Name: "w1", DisplayName: "w1"})
	require.NoError(t, err)
	assert.True(t, created)

	created, err = repo.CreateWidget(ctx, &models.Widget{OrgID: "org-1", Name: "w1", DisplayName: "w1"})
	require.NoError(t, err)
	assert.False(t, created)
}

func TestWidgetRepo_GetIsTenantScoped(t *testing.T) {
	repo := NewWidgetRepo(openTestDB(t))
	ctx := context.Background()
	_, err := repo.CreateWidget(ctx, &models.Widget{OrgID: "org-1", Name: "w1", DisplayName: "w1"})
	require.NoError(t, err)

	_, err = repo.GetWidget(ctx, "org-2", "w1")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
