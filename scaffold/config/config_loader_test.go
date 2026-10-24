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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead_CollectsEveryMissingRequiredVar(t *testing.T) {
	for _, k := range []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME"} {
		t.Setenv(k, "")
	}
	r := &configReader{}
	read(r)
	assert.Len(t, r.errors, 4)
}

func TestValidateAuthConfig_RequiresAllTogetherOutsideLocalDev(t *testing.T) {
	r := &configReader{}
	validateAuthConfig(&Config{Auth: AuthConfig{JWKSURL: "https://idp/jwks"}}, r)
	assert.Len(t, r.errors, 1)

	r = &configReader{}
	validateAuthConfig(&Config{IsLocalDevEnv: true}, r)
	assert.Empty(t, r.errors)
}
