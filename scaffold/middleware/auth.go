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
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"example.com/widgetsvc/middleware/logger"
	"example.com/widgetsvc/rbac"
	"example.com/widgetsvc/utils"
)

// Claims is the verified identity of the caller. OrgID is the only source of
// tenant identity; request paths and bodies are never trusted for it.
type Claims struct {
	Subject string
	OrgID   string
	OrgName string
	Scopes  []string
}

// TokenVerifier turns a bearer token into verified Claims. Deployments inject
// their own (JWKS, introspection); the scaffold ships none for production.
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (*Claims, error)
}

var ErrInvalidToken = errors.New("invalid token")

type claimsKey struct{}

func WithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, c)
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey{}).(*Claims)
	return c, ok
}

// OrgIDFromRequest returns the tenant ID from verified claims. Routes reach a
// controller only through Authenticate, so a missing value is a wiring bug.
func OrgIDFromRequest(r *http.Request) string {
	if c, ok := ClaimsFromContext(r.Context()); ok {
		return c.OrgID
	}
	return ""
}

func Authenticate(v TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				utils.WriteErrorResponse(w, http.StatusUnauthorized, "Missing bearer token")
				return
			}
			claims, err := v.Verify(r.Context(), token)
			if err != nil {
				logger.GetLogger(r.Context()).Debug("Authenticate: token rejected", "error", err)
				utils.WriteErrorResponse(w, http.StatusUnauthorized, "Invalid token")
				return
			}
			if claims.OrgID == "" {
				utils.WriteErrorResponse(w, http.StatusUnauthorized, "Token carries no organization")
				return
			}
			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}

func RequirePermission(perm rbac.Permission) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			c, ok := ClaimsFromContext(r.Context())
			if !ok || !slices.Contains(c.Scopes, string(perm)) {
				utils.WriteErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
				return
			}
			next(w, r)
		}
	}
}

// RequireOrgMatch rejects a request whose {orgName} differs from the token's
// org. Services still check the org against the resource they load.
func RequireOrgMatch() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			c, ok := ClaimsFromContext(r.Context())
			if !ok || c.OrgName != r.PathValue(utils.PathParamOrgName) {
				slog.Default().Debug("RequireOrgMatch: org mismatch")
				utils.WriteErrorResponse(w, http.StatusForbidden, "Organization mismatch")
				return
			}
			next(w, r)
		}
	}
}
