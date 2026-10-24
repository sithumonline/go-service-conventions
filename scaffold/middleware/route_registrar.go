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

package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"example.com/widgetsvc/rbac"
	"example.com/widgetsvc/utils"
)

// RouteInfo is what the registrar recorded for one route, so invariant tests
// can inspect every route without starting a server.
type RouteInfo struct {
	Pattern    string
	Permission rbac.Permission
}

// RouteRegistrar is the only way authenticated routes are added, so every
// route gets path validation and its own authorization.
type RouteRegistrar struct {
	mux    *http.ServeMux
	Routes []RouteInfo
}

func NewRouteRegistrar(mux *http.ServeMux) *RouteRegistrar {
	return &RouteRegistrar{mux: mux, Routes: nil}
}

// HandleFuncWithValidationAndAuthz composes path-param validation, the
// permission check, and the org match (when the path has {orgName}).
func (rr *RouteRegistrar) HandleFuncWithValidationAndAuthz(pattern string, perm rbac.Permission, h http.HandlerFunc) {
	if perm == "" {
		panic(fmt.Sprintf("route %q registered without a permission", pattern))
	}
	// Wrapped inside-out, so requests run: validate → permission → org match.
	if strings.Contains(pattern, "{"+utils.PathParamOrgName+"}") {
		h = RequireOrgMatch()(h)
	}
	h = RequirePermission(perm)(h)
	h = validatePathParams(h)
	rr.Routes = append(rr.Routes, RouteInfo{Pattern: pattern, Permission: perm})
	rr.mux.HandleFunc(pattern, h)
}

var pathParamPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,62}$`)

func validatePathParams(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, name := range []string{utils.PathParamOrgName, utils.PathParamWidgetName} {
			if v := r.PathValue(name); v != "" && !pathParamPattern.MatchString(v) {
				utils.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid path parameter: %s", name))
				return
			}
		}
		next(w, r)
	}
}
