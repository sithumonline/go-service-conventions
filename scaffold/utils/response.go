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
	"encoding/json"
	"net/http"

	"example.com/widgetsvc/spec"
)

func WriteSuccessResponse[T any](w http.ResponseWriter, statusCode int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if statusCode == http.StatusNoContent {
		return
	}
	_ = json.NewEncoder(w).Encode(data) // headers are already sent; nothing useful to do on failure
}

// WriteErrorResponse writes the error envelope with a code derived from the
// status. Use WriteErrorResponseWithReason to choose the code.
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	WriteErrorResponseWithReason(w, statusCode, message, "", "")
}

// WriteErrorResponseWithReason writes the error envelope.
// message is for people, reason is technical detail, code is for programs.
func WriteErrorResponseWithReason(w http.ResponseWriter, statusCode int, message, reason, code string) {
	if code == "" {
		code = statusCodeToErrorCode(statusCode)
	}
	payload := spec.ErrorResponse{Message: message, Code: code, Reason: nil}
	if reason != "" {
		payload.Reason = &reason
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload) // headers are already sent; nothing useful to do on failure
}

// WriteValidationErrorResponse answers 400, using a ValidationError's Message
// and Reason when err carries one.
func WriteValidationErrorResponse(w http.ResponseWriter, err error) {
	if ve := IsValidationError(err); ve != nil {
		WriteErrorResponseWithReason(w, http.StatusBadRequest, ve.Message, ve.Reason, ErrCodeValidation)
		return
	}
	WriteErrorResponseWithReason(w, http.StatusBadRequest, err.Error(), err.Error(), ErrCodeValidation)
}

func statusCodeToErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return ErrCodeBadRequest
	case http.StatusUnauthorized:
		return ErrCodeUnauthorized
	case http.StatusForbidden:
		return ErrCodeForbidden
	case http.StatusNotFound:
		return ErrCodeNotFound
	case http.StatusConflict:
		return ErrCodeConflict
	case http.StatusServiceUnavailable:
		return ErrCodeServiceUnavailable
	default:
		return ErrCodeInternalError
	}
}
