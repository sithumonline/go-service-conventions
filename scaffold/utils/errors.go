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

package utils

import (
	"errors"
	"fmt"
)

// ValidationError carries a user-facing Message and a technical Reason for a
// client mistake. Error() returns Reason so logs keep the detail.
type ValidationError struct {
	// Message is shown in the UI: short, non-technical, actionable.
	Message string
	// Reason is for developers: field names and the rule that failed.
	Reason string
	// sentinel lets errors.Is(err, ErrInvalidInput) still match. Unexported so
	// a constructor is the only way to set it.
	sentinel error
}

func (e *ValidationError) Error() string { return e.Reason }

func (e *ValidationError) Unwrap() error { return e.sentinel }

func NewValidationError(message, reason string) *ValidationError {
	return &ValidationError{Message: message, Reason: reason, sentinel: nil}
}

func NewValidationErrorf(message, reasonFmt string, args ...any) *ValidationError {
	return NewValidationError(message, fmt.Sprintf(reasonFmt, args...))
}

// NewInvalidInputError is NewValidationError for a failure that must also
// match ErrInvalidInput under errors.Is.
func NewInvalidInputError(message, reason string) *ValidationError {
	return &ValidationError{Message: message, Reason: reason, sentinel: ErrInvalidInput}
}

// IsValidationError returns the ValidationError in err's chain, or nil.
func IsValidationError(err error) *ValidationError {
	var ve *ValidationError
	if errors.As(err, &ve) {
		return ve
	}
	return nil
}

var (
	// Resource not found errors
	ErrWidgetNotFound = errors.New("widget not found")

	// Conflict errors
	ErrWidgetAlreadyExists = errors.New("widget already exists")

	// Input errors
	ErrInvalidInput = errors.New("invalid input")
)
