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

package db

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"example.com/widgetsvc/config"
)

type slogWriter struct{}

func (slogWriter) Printf(format string, v ...any) {
	slog.Warn(fmt.Sprintf(format, v...))
}

// Open connects to Postgres. Repositories add the request context per call.
func Open(cfg config.POSTGRESQL) (*gorm.DB, error) {
	gdb, err := gorm.Open(postgres.Open(connString(cfg)), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger: gormlogger.New(slogWriter{}, gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: true,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("db.Open: %w", err)
	}
	return gdb, nil
}

func connString(p config.POSTGRESQL) string {
	s := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", p.Host, p.Port, p.User, p.Password, p.DBName)
	if p.SSLMode != "" {
		s += " sslmode=" + p.SSLMode
	}
	return s
}
