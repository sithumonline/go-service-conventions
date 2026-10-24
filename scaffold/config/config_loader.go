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
	"fmt"
)

var config *Config

func GetConfig() *Config {
	return config
}

// Load reads every environment variable, validates the result, and exits the
// process when anything is wrong. Call it once from main before anything else.
func Load() *Config {
	r := &configReader{}
	cfg := read(r)
	validate(cfg, r)
	r.logAndExitIfErrorsFound()
	config = cfg
	return cfg
}

func read(r *configReader) *Config {
	cfg := &Config{
		ServerPort:          int(r.readOptionalInt64("SERVER_PORT", 8080)),
		LogLevel:            r.readOptionalString("LOG_LEVEL", "INFO"),
		IsLocalDevEnv:       r.readOptionalBool("IS_LOCAL_DEV_ENV", false),
		CORSAllowedOrigin:   r.readOptionalString("CORS_ALLOWED_ORIGIN", "http://localhost:3000"),
		ReadTimeoutSeconds:  int(r.readOptionalInt64("HTTP_READ_TIMEOUT_SECONDS", 10)),
		WriteTimeoutSeconds: int(r.readOptionalInt64("HTTP_WRITE_TIMEOUT_SECONDS", 90)),
		IdleTimeoutSeconds:  int(r.readOptionalInt64("HTTP_IDLE_TIMEOUT_SECONDS", 60)),
		MaxHeaderBytes:      int(r.readOptionalInt64("HTTP_MAX_HEADER_BYTES", 64*1024)),
		ShutdownSeconds:     int(r.readOptionalInt64("SHUTDOWN_TIMEOUT_SECONDS", 30)),
		POSTGRESQL: POSTGRESQL{
			Host:     r.readRequiredString("DB_HOST"),
			Port:     int(r.readOptionalInt64("DB_PORT", 5432)),
			User:     r.readRequiredString("DB_USER"),
			Password: r.readRequiredString("DB_PASSWORD"),
			DBName:   r.readRequiredString("DB_NAME"),
			SSLMode:  r.readOptionalString("DB_SSL_MODE", ""),
		},
		Auth: AuthConfig{
			JWKSURL:  r.readOptionalString("KEY_MANAGER_JWKS_URL", ""),
			Issuer:   r.readOptionalStringList("KEY_MANAGER_ISSUER", ""),
			Audience: r.readOptionalStringList("KEY_MANAGER_AUDIENCE", ""),
		},
	}
	return cfg
}

func validate(cfg *Config, r *configReader) {
	validateHTTPServerConfigs(cfg, r)
	validateAuthConfig(cfg, r)
}

func validateHTTPServerConfigs(cfg *Config, r *configReader) {
	if cfg.ServerPort <= 0 || cfg.ServerPort > 65535 {
		r.errors = append(r.errors, fmt.Errorf("SERVER_PORT must be between 1 and 65535, got %d", cfg.ServerPort))
	}
	for name, v := range map[string]int{
		"HTTP_READ_TIMEOUT_SECONDS":  cfg.ReadTimeoutSeconds,
		"HTTP_WRITE_TIMEOUT_SECONDS": cfg.WriteTimeoutSeconds,
		"HTTP_IDLE_TIMEOUT_SECONDS":  cfg.IdleTimeoutSeconds,
		"SHUTDOWN_TIMEOUT_SECONDS":   cfg.ShutdownSeconds,
	} {
		if v <= 0 {
			r.errors = append(r.errors, fmt.Errorf("%s must be positive, got %d", name, v))
		}
	}
}

// validateAuthConfig checks the token settings together: outside local dev
// they are all required, since a partial set fails on the first request.
func validateAuthConfig(cfg *Config, r *configReader) {
	if cfg.IsLocalDevEnv {
		return
	}
	if cfg.Auth.JWKSURL == "" || len(cfg.Auth.Issuer) == 0 || len(cfg.Auth.Audience) == 0 {
		r.errors = append(r.errors, fmt.Errorf("KEY_MANAGER_JWKS_URL, KEY_MANAGER_ISSUER and KEY_MANAGER_AUDIENCE are required unless IS_LOCAL_DEV_ENV=true"))
	}
}
