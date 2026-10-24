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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"example.com/widgetsvc/middleware"
	"example.com/widgetsvc/middleware/logger"
	"example.com/widgetsvc/services"
	"example.com/widgetsvc/spec"
	"example.com/widgetsvc/utils"
)

type WidgetController interface {
	ListWidgets(w http.ResponseWriter, r *http.Request)
	GetWidget(w http.ResponseWriter, r *http.Request)
	CreateWidget(w http.ResponseWriter, r *http.Request)
	DeleteWidget(w http.ResponseWriter, r *http.Request)
}

type widgetController struct {
	widgetService services.WidgetService
}

func NewWidgetController(widgetService services.WidgetService) WidgetController {
	return &widgetController{widgetService: widgetService}
}

func (c *widgetController) ListWidgets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLogger(ctx)
	orgID := middleware.OrgIDFromRequest(r)

	limit, err := queryInt(r, "limit", utils.DefaultLimit)
	if err != nil || limit < utils.MinLimit || limit > utils.MaxLimit {
		utils.WriteErrorResponse(w, http.StatusBadRequest,
			fmt.Sprintf("Invalid limit parameter: must be between %d and %d", utils.MinLimit, utils.MaxLimit))
		return
	}
	offset, err := queryInt(r, "offset", utils.DefaultOffset)
	if err != nil || offset < utils.MinOffset {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid offset parameter: must be >= 0")
		return
	}

	result, err := c.widgetService.ListWidgets(ctx, orgID, limit, offset)
	if err != nil {
		log.Error("Failed to list widgets", "error", err)
		handleCommonErrors(w, err, "Failed to list widgets")
		return
	}
	utils.WriteSuccessResponse(w, http.StatusOK, result)
}

func (c *widgetController) GetWidget(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLogger(ctx)
	orgID := middleware.OrgIDFromRequest(r)
	name := r.PathValue(utils.PathParamWidgetName)

	result, err := c.widgetService.GetWidget(ctx, orgID, name)
	if err != nil {
		log.Error("Failed to get widget", "error", err)
		handleCommonErrors(w, err, "Failed to get widget")
		return
	}
	utils.WriteSuccessResponse(w, http.StatusOK, result)
}

func (c *widgetController) CreateWidget(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLogger(ctx)
	orgID := middleware.OrgIDFromRequest(r)

	var payload spec.CreateWidgetRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if payload.Name == "" {
		utils.WriteValidationErrorResponse(w,
			utils.NewValidationError("Widget name is required", "name is empty"))
		return
	}

	result, err := c.widgetService.CreateWidget(ctx, orgID, &payload)
	if err != nil {
		log.Error("Failed to create widget", "error", err)
		handleCommonErrors(w, err, "Failed to create widget")
		return
	}
	log.Info("Created widget", "org", orgID, "widget", result.Name)
	utils.WriteSuccessResponse(w, http.StatusCreated, result)
}

func (c *widgetController) DeleteWidget(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLogger(ctx)
	orgID := middleware.OrgIDFromRequest(r)
	name := r.PathValue(utils.PathParamWidgetName)

	if err := c.widgetService.DeleteWidget(ctx, orgID, name); err != nil {
		log.Error("Failed to delete widget", "error", err)
		handleCommonErrors(w, err, "Failed to delete widget")
		return
	}
	log.Info("Deleted widget", "org", orgID, "widget", name)
	w.WriteHeader(http.StatusNoContent)
}

func queryInt(r *http.Request, key string, def int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def, nil
	}
	return strconv.Atoi(raw)
}
