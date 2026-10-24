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

package wiring

import (
	"gorm.io/gorm"

	"example.com/widgetsvc/controllers"
	"example.com/widgetsvc/middleware"
	"example.com/widgetsvc/repositories"
	"example.com/widgetsvc/services"
)

// AppParams is everything the router needs. With google/wire this struct is
// the injector's output (wire.Struct(new(AppParams), "*")); here it is built
// by hand with the same layer order.
type AppParams struct {
	WidgetController controllers.WidgetController
	TokenVerifier    middleware.TokenVerifier
}

func Build(db *gorm.DB, verifier middleware.TokenVerifier) *AppParams {
	// Repositories
	widgetRepo := repositories.NewWidgetRepo(db)

	// Services
	widgetService := services.NewWidgetService(widgetRepo)

	// Controllers
	return &AppParams{
		WidgetController: controllers.NewWidgetController(widgetService),
		TokenVerifier:    verifier,
	}
}
