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

package dbmigrations

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

type migration struct {
	ID      int
	Migrate func(db *gorm.DB) error
}

// migrations is append-only. Never edit a migration that has shipped.
var migrations = []migration{
	migration001,
}

// Migrate applies every migration newer than the recorded version, each in
// its own transaction, so a crash leaves the schema at a known version.
func Migrate(db *gorm.DB) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (id INT PRIMARY KEY)`).Error; err != nil {
		return fmt.Errorf("dbmigrations.Migrate: %w", err)
	}
	for _, m := range migrations {
		err := db.Transaction(func(tx *gorm.DB) error {
			res := tx.Exec(`INSERT INTO schema_migrations (id) VALUES (?) ON CONFLICT DO NOTHING`, m.ID)
			if res.Error != nil || res.RowsAffected == 0 {
				return res.Error
			}
			slog.Info("Applying migration", "id", m.ID)
			return m.Migrate(tx)
		})
		if err != nil {
			return fmt.Errorf("dbmigrations.Migrate: migration %d: %w", m.ID, err)
		}
	}
	return nil
}
