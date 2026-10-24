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

package controllers

import (
	"errors"
	"net/http"

	"example.com/widgetsvc/utils"
)

// handleCommonErrors maps service errors to responses. Unknown errors answer
// 500 with fallbackMsg; the caller has already logged the detail.
func handleCommonErrors(w http.ResponseWriter, err error, fallbackMsg string) {
	switch {
	// Keep FIRST: a ValidationError may also match ErrInvalidInput below,
	// which would replace its specific Message with a generic one.
	case utils.IsValidationError(err) != nil:
		utils.WriteValidationErrorResponse(w, err)

	// Not found errors
	case errors.Is(err, utils.ErrWidgetNotFound):
		utils.WriteErrorResponseWithReason(w, http.StatusNotFound,
			"Widget not found", err.Error(), utils.ErrCodeWidgetNotFound)

	// Conflict errors
	case errors.Is(err, utils.ErrWidgetAlreadyExists):
		utils.WriteErrorResponseWithReason(w, http.StatusConflict,
			"Widget already exists", err.Error(), utils.ErrCodeWidgetAlreadyExists)

	// Input errors
	case errors.Is(err, utils.ErrInvalidInput):
		utils.WriteErrorResponseWithReason(w, http.StatusBadRequest,
			"Invalid input provided", err.Error(), utils.ErrCodeBadRequest)

	default:
		utils.WriteErrorResponse(w, http.StatusInternalServerError, fallbackMsg)
	}
}
