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

	"example.com/widgetsvc/config"
	"example.com/widgetsvc/middleware"
	"example.com/widgetsvc/middleware/logger"
	"example.com/widgetsvc/wiring"
)

// registerAPIRoutes registers every authenticated /api/v1 route. Keeping the
// list in one function lets route_authz_invariant_test.go check all of them.
func registerAPIRoutes(rr *middleware.RouteRegistrar, params *wiring.AppParams) {
	registerWidgetRoutes(rr, params.WidgetController)
}

func MakeHTTPHandler(cfg *config.Config, params *wiring.AppParams) http.Handler {
	mux := http.NewServeMux()
	registerHealthCheck(mux)

	apiMux := http.NewServeMux()
	registerAPIRoutes(middleware.NewRouteRegistrar(apiMux), params)

	// Applied inside-out: the last wrap runs first.
	apiHandler := http.Handler(apiMux)
	apiHandler = middleware.Authenticate(params.TokenVerifier)(apiHandler)
	apiHandler = logger.RequestLogger()(apiHandler)
	apiHandler = middleware.AddCorrelationID()(apiHandler)
	apiHandler = middleware.CORS(cfg.CORSAllowedOrigin)(apiHandler)
	apiHandler = middleware.RecovererOnPanic()(apiHandler)

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiHandler))
	return mux
}

func registerHealthCheck(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
