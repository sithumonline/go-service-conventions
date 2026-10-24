# Config, logging, bootstrap, DB, outbound clients

## Config

A `config` package exposes a typed `Config` struct with nested sub-structs (`POSTGRESQL`, `OTELConfig`, `ObserverConfig`) and a getter `config.GetConfig() *Config`.

Read env vars through a small reader that collects errors instead of failing on the first:

```go
type configReader struct{ errors []error }

func (c *configReader) readRequiredString(name string) string
func (c *configReader) readOptionalString(name, def string) string
func (c *configReader) readOptionalInt64(name string, def int64) int64
func (c *configReader) readOptionalBool(name string, def bool) bool
func (c *configReader) readNullableInt64(name string) *int64            // nil when unset
func (c *configReader) readOptionalStringList(name, def string) []string // comma-separated, trimmed
func (c *configReader) logAndExitIfErrorsFound()                        // slog.Error all, os.Exit(1)
```

- Load order: optional `.env` via `ENV_FILE_PATH` → read every var → run `validateXxx(cfg, r)` functions → `logAndExitIfErrorsFound()`.
- One `validateXxx` per concern (`validateHTTPServerConfigs`, `validateObserverURLs`). Co-dependent values are checked together: "if URL is set, client ID and secret are required."
- Secrets are `readRequiredString` or have an empty default that fails validation. Never a working default.
- A default that only makes sense locally is chosen from an `IS_LOCAL_DEV_ENV` flag, never silently in production.
- Durations are int seconds in env vars (`HTTP_READ_TIMEOUT_SECONDS`) converted to `time.Duration` at use.
- In tests, set env with `t.Setenv`.

## Logging

- `log/slog` with a JSON handler, level from `LOG_LEVEL`.
- `middleware/logger` provides `WithLogger(ctx, l)` and `GetLogger(ctx)` with an unexported `loggerKey struct{}` context key. `GetLogger` falls back to `slog.Default()`.
- The request logger middleware attaches `method`, `path`, `correlation_id`. Add resource identifiers at the call site as key-value pairs: `log.Info("Deleted widget", "org", orgID, "widget", name)`.
- Messages start with a capital letter and name the operation: `"Failed to list widgets"`. Keys are snake_case or lowerCamel, consistently per service.
- Background workers build their own logger with `slog.Default().With("component", "reconciler")`.

## Bootstrap (`main.go` → `app.Run`)

`main.go` stays tiny: parse flags (`-server`, `-migrate`), build deployment-specific providers, call `app.Run(providers..., app.Options{...})`.

`app.Run`:

1. `setupLogger(cfg)`, `automaxprocs` if enabled.
2. Run migrations when `-migrate`. Return if not serving.
3. Open DB, build DI graph. Any construction error → `slog.Error(...)`; `os.Exit(1)`.
4. Start background workers with a cancellable context.
5. Start the HTTP server(s) in goroutines with read, write, idle timeouts and `MaxHeaderBytes` from config.
6. Block on a signal channel (`SIGINT`, `SIGTERM`). Cancel workers, then `server.Shutdown(ctxWithTimeout)`.

Package-level setters for process-wide singletons are allowed only before the server starts, with a comment saying so.

## Database

- GORM on Postgres via pgx. Pool settings from config (`MaxIdleCount`, `MaxOpenCount`, lifetimes) as nullable ints.
- GORM's slow-query logger routed into slog. `SkipDefaultTransaction: true`.
- Transactions travel in the context: `db.CtxWithTx(ctx, tx)` and `db.DB(ctx)` returns the tx if present, else the pool with a default operation timeout.
- Tenant column name is the same on every tenant-scoped table.

## Outbound HTTP clients (`clients/<name>svc/`)

```go
//go:generate moq -rm -fmt goimports -skip-ensure -pkg clientmocks -out ../clientmocks/widget_client_fake.go . WidgetSvcClient:WidgetSvcClientMock
type WidgetSvcClient interface {
	GetWidget(ctx context.Context, org, name string) (*models.Widget, error)
}

type Config struct {
	BaseURL      string
	AuthProvider AuthProvider
	RetryConfig  requests.RequestRetryConfig
}

func NewWidgetClient(cfg *Config) (WidgetSvcClient, error) // validates cfg, returns error on missing fields
```

- The constructor validates its config and returns an error. It never panics.
- Trim the trailing `/` from `BaseURL` once.
- Build query strings with `url.Values`. Escape path segments.
- On a 401, invalidate the cached token and retry once.
- Wrap errors with `pkgname.Method:` prefix.
- A shared `requests` package provides a retrying `HttpClient` and a typed `*HttpError{StatusCode}` that callers detect with `errors.As`.
- Token caches refresh before expiry with a buffer (30s to 5m).

## Concurrent fan-out

Bound every fan-out with a semaphore. When a fan-out nests (items, then sub-items per item), use one semaphore per level so outer workers cannot starve inner ones. Give the whole operation a deadline, collect partial failures, and report them in the response (`failedIds`, `truncated`) instead of failing the request.
