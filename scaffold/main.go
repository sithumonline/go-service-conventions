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

package main

import (
	"flag"

	"example.com/widgetsvc/app"
	"example.com/widgetsvc/config"
)

func main() {
	serverFlag := flag.Bool("server", true, "start the HTTP server")
	migrateFlag := flag.Bool("migrate", false, "migrate the database")
	flag.Parse()

	cfg := config.Load()

	app.Run(cfg, app.Options{
		Server:  *serverFlag,
		Migrate: *migrateFlag,
		// Inject a JWKS-backed middleware.TokenVerifier here.
		TokenVerifier: nil,
	})
}
