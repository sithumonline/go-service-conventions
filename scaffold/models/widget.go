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

package models

import (
	"time"

	"github.com/google/uuid"
)

type Widget struct {
	ID          uuid.UUID         `gorm:"column:id;primaryKey"`
	OrgID       string            `gorm:"column:org_id"`
	Name        string            `gorm:"column:name"`
	DisplayName string            `gorm:"column:display_name"`
	Labels      map[string]string `gorm:"column:labels;type:jsonb;serializer:json"`
	CreatedAt   time.Time         `gorm:"column:created_at"`
	UpdatedAt   time.Time         `gorm:"column:updated_at"`
}

func (Widget) TableName() string { return "widgets" }
