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
	"context"
	"net/http"

	"github.com/google/uuid"

	"example.com/widgetsvc/utils"
)

const correlationIDHeader = "X-Correlation-ID"

// AddCorrelationID reuses the caller's correlation ID or mints one, and echoes
// it on the response so a client can quote it in a bug report.
func AddCorrelationID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(correlationIDHeader)
			if id == "" {
				id = uuid.NewString()
			}
			w.Header().Set(correlationIDHeader, id)
			ctx := context.WithValue(r.Context(), utils.CorrelationIDCtxKey(), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
