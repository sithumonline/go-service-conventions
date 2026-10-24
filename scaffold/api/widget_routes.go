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
	"example.com/widgetsvc/controllers"
	"example.com/widgetsvc/middleware"
	"example.com/widgetsvc/rbac"
)

func registerWidgetRoutes(rr *middleware.RouteRegistrar, ctrl controllers.WidgetController) {
	rr.HandleFuncWithValidationAndAuthz("GET /orgs/{orgName}/widgets", rbac.WidgetRead, ctrl.ListWidgets)
	rr.HandleFuncWithValidationAndAuthz("POST /orgs/{orgName}/widgets", rbac.WidgetCreate, ctrl.CreateWidget)
	rr.HandleFuncWithValidationAndAuthz("GET /orgs/{orgName}/widgets/{widgetName}", rbac.WidgetRead, ctrl.GetWidget)
	rr.HandleFuncWithValidationAndAuthz("DELETE /orgs/{orgName}/widgets/{widgetName}", rbac.WidgetDelete, ctrl.DeleteWidget)
}
