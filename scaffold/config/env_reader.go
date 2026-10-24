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
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// configReader collects every problem so startup reports them all at once.
type configReader struct {
	errors []error
}

func (c *configReader) logAndExitIfErrorsFound() {
	if len(c.errors) == 0 {
		return
	}
	msgs := make([]string, 0, len(c.errors))
	for _, err := range c.errors {
		msgs = append(msgs, err.Error())
	}
	slog.Error("configReader: errors found while reading config", "errors", msgs)
	os.Exit(1)
}

func (c *configReader) readRequiredString(name string) string {
	v := os.Getenv(name)
	if v == "" {
		c.errors = append(c.errors, fmt.Errorf("required environment variable %s not found", name))
	}
	return v
}

func (c *configReader) readOptionalString(name, defaultValue string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return defaultValue
}

func (c *configReader) readOptionalInt64(name string, defaultValue int64) int64 {
	v := os.Getenv(name)
	if v == "" {
		return defaultValue
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		c.errors = append(c.errors, fmt.Errorf("environment variable %s is not a valid integer: %w", name, err))
		return 0
	}
	return n
}

func (c *configReader) readOptionalBool(name string, defaultValue bool) bool {
	v := os.Getenv(name)
	if v == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		c.errors = append(c.errors, fmt.Errorf("environment variable %s is not a valid boolean: %w", name, err))
		return false
	}
	return b
}

func (c *configReader) readOptionalStringList(name, defaultValue string) []string {
	raw := c.readOptionalString(name, defaultValue)
	out := []string{}
	for _, part := range strings.Split(raw, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
