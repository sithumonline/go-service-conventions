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

package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"example.com/widgetsvc/controllers"
	"example.com/widgetsvc/middleware"
	"example.com/widgetsvc/rbac"
	"example.com/widgetsvc/wiring"
)

func TestEveryRouteHasAGrantedPermission(t *testing.T) {
	rr := middleware.NewRouteRegistrar(http.NewServeMux())
	registerAPIRoutes(rr, &wiring.AppParams{
		WidgetController: controllers.NewWidgetController(nil),
		TokenVerifier:    nil,
	})
	require.NotEmpty(t, rr.Routes)

	granted := map[rbac.Permission]bool{}
	for _, perms := range rbac.PredefinedRolePermissions {
		for _, p := range perms {
			granted[p] = true
		}
	}
	for _, route := range rr.Routes {
		assert.True(t, granted[route.Permission], "%s needs %s, which no role grants", route.Pattern, route.Permission)
	}
}
