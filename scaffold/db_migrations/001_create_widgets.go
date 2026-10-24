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

import "gorm.io/gorm"

var migration001 = migration{
	ID: 1,
	Migrate: func(db *gorm.DB) error {
		return db.Exec(`
			CREATE TABLE widgets (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				org_id VARCHAR(255) NOT NULL,
				name VARCHAR(255) NOT NULL,
				display_name VARCHAR(255) NOT NULL,
				labels JSONB NOT NULL DEFAULT '{}',
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

				CONSTRAINT uq_widgets_org_name UNIQUE (org_id, name)
			);
		`).Error
	},
}
