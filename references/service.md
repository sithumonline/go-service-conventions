# Service layout and layering

## Directory layout

```text
main.go                 flags, build deployment-specific providers, call app.Run
app/                    Run(): logger, migrations, DB, DI, servers, signal-driven shutdown
config/                 Config struct, env reader, validate*() funcs
api/                    app.go (mux + middleware chain), <resource>_routes.go
controllers/            HTTP handlers, one interface + impl per resource
services/               business logic, one interface + impl per resource
repositories/           persistence; repomocks/ holds generated mocks
clients/<name>svc/      outbound HTTP clients; clientmocks/ holds generated mocks
models/                 DB entities and internal DTOs
spec/                   generated from docs/openapi.yaml, never edited
middleware/             auth, authz, correlation ID, CORS, panic recovery, logger/
rbac/                   permission constants, role → permission map
db/                     connection, ctx-carried transaction helpers
db_migrations/          numbered Go migrations (001_..., 002_...)
utils/                  sentinel errors, error codes, response writers, validation
wiring/                 DI provider sets (google/wire), AppParams
```

Files are `snake_case.go`, named `<resource>_<layer>.go` (`widget_controller.go`, `widget_service.go`, `widget_repository.go`, `widget_routes.go`).

## The constructor shape (every layer)

```go
type WidgetService interface {
	GetWidget(ctx context.Context, orgID, name string) (*models.WidgetResponse, error)
}

type widgetService struct {
	repo repositories.WidgetRepository
}

func NewWidgetService(repo repositories.WidgetRepository) WidgetService {
	return &widgetService{repo: repo}
}
```

- The interface is exported. The struct is unexported. The constructor returns the interface.
- Dependencies are interfaces. A service never takes `*gorm.DB` or a concrete client.
- Group interface methods with short comments (`// Kind CRUD`, `// Version management`) when the interface is long.
- Exported service methods get a one-line doc comment saying what they return or do.
- Load a resource once per request and pass it down. Don't re-fetch.

## Controller

One method per route, signature `func (c *xController) Op(w http.ResponseWriter, r *http.Request)`. Every handler has the same skeleton:

```go
func (c *widgetController) GetWidget(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.GetLogger(ctx)
	orgID := middleware.OrgIDFromRequest(r)           // from token claims
	name := r.PathValue(utils.PathParamWidgetName)    // constant, not a literal

	result, err := c.widgetService.GetWidget(ctx, orgID, name)
	if err != nil {
		log.Error("Failed to get widget", "error", err)
		handleCommonErrors(w, err, "Failed to get widget")
		return
	}
	utils.WriteSuccessResponse(w, http.StatusOK, result)
}
```

- Body: `json.NewDecoder(r.Body).Decode(&payload)` into a `spec.*Request`. On failure, 400 `"Invalid request body"`.
- Required-field and format checks happen in the controller before the service call. Use `utils.WriteValidationErrorResponse` for `ValidationError`s and log those at Debug.
- Pagination: `limit`/`offset` query params with `DefaultLimit`, `MinLimit`, `MaxLimit` constants. Reject out-of-range values with 400 and a message that names the range.
- Create → 201. Delete → 204 with `w.WriteHeader(http.StatusNoContent)`. Everything else → 200.
- List endpoints never return `null`. Return an empty slice.
- A comment in a controller explains a non-obvious status choice (for example, 200 with null for "not published yet").
- No business logic. If a controller starts branching on domain state, move it into the service.

## Service

- Map repository errors immediately after each call:

```go
w, err := s.repo.GetWidget(ctx, orgID, name)
if err != nil {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrWidgetNotFound
	}
	return nil, fmt.Errorf("failed to get widget: %w", err)
}
```

- Validation gates and preconditions (`ErrWidgetHasInstances`) return sentinels, not formatted strings.
- Convert models to response types with small `toXxxResponse(m *models.X) *models.XResponse` helpers in the same file.
- A deliberate trade-off (an unlocked check-then-act, a skipped re-fetch) gets a `NOTE:` comment on the interface method explaining why, so nobody "fixes" it.

## Repository

```go
//go:generate moq -rm -fmt goimports -skip-ensure -pkg repomocks -out repomocks/widget_repository_mock.go . WidgetRepository:WidgetRepositoryMock
type WidgetRepository interface { ... }

type widgetRepo struct{ db *gorm.DB }

func NewWidgetRepo(db *gorm.DB) WidgetRepository { return &widgetRepo{db: db} }
```

- Every query starts with `r.db.WithContext(ctx)` and filters on the tenant column first: `Where("org_id = ? AND name = ?", orgID, name)`.
- Return `gorm.ErrRecordNotFound` as-is. Mapping is the service's job.
- Update and delete check `RowsAffected == 0` and return `gorm.ErrRecordNotFound`.
- Assign `uuid.New()` on create when the ID is `uuid.Nil`. Normalize nil maps and slices before writing JSON columns.
- Multi-row writes go in `r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {...})`.
- When a service needs a transaction across repositories, inject a `TxRunner func(ctx context.Context, fn func(tx *gorm.DB) error) error` so a unit test can stub it.
- Prefer `ON CONFLICT` upserts to read-then-write.
- Lists return `(items, total, err)` and order deterministically (`created_at DESC`).

## Model

```go
type Widget struct {
	ID        uuid.UUID         `gorm:"column:id;primaryKey"`
	OrgID     string            `gorm:"column:org_id"`
	Name      string            `gorm:"column:name"`
	Labels    map[string]string `gorm:"column:labels;type:jsonb;serializer:json"`
	CreatedAt time.Time         `gorm:"column:created_at"`
	UpdatedAt time.Time         `gorm:"column:updated_at"`
}

func (Widget) TableName() string { return "widgets" }
```

Every field has an explicit `column:` tag. Schema changes are numbered migration files with raw SQL (`CREATE TABLE`, named constraints `uq_<table>_<cols>`, `fk_<table>_<ref>`), never `AutoMigrate`.

## Routes and authorization

```go
func registerWidgetRoutes(rr *middleware.RouteRegistrar, ctrl controllers.WidgetController) {
	rr.HandleFuncWithValidationAndAuthz("GET /orgs/{orgName}/widgets", rbac.WidgetRead, ctrl.ListWidgets)
	rr.HandleFuncWithValidationAndAuthz("POST /orgs/{orgName}/widgets", rbac.WidgetCreate, ctrl.CreateWidget)
	rr.HandleFuncWithValidationAndAuthz("DELETE /orgs/{orgName}/widgets/{widgetName}", rbac.WidgetDelete, ctrl.DeleteWidget)
}
```

- Stdlib `http.ServeMux` with Go 1.22 method patterns. No router framework.
- The registrar composes path-param validation → permission check (token scopes) → org-match (token org vs `{orgName}`). Variants exist for any-of permissions and request-dependent permissions.
- All `register*Routes` calls live in one function, so a test can register everything against a bare registrar and assert invariants (every mutating route is authorized and audited).
- Unauthenticated routes (`/health`, `/.well-known/...`) are registered on the root mux, outside the authenticated sub-mux.
- Middleware order on the API sub-mux, outermost first: panic recovery → CORS → correlation ID → request logger → auth → route-level authz.
- Permissions are typed constants, `resource:verb` with `create`/`read`/`update`/`delete`, or `manage` only where coarser grants already exist.

## Dependency injection

`google/wire` provider sets per layer (`configProviderSet`, `clientProviderSet`, `repositoryProviderSet`, `serviceProviderSet`, `controllerProviderSet`) feed one `AppParams` struct that the router reads. Constructors are plain `NewXxx` functions. Use `ProvideXxx` wrappers only when construction needs config or can fail. `wire.Bind` maps a concrete type to an interface when one struct serves several interfaces. A small service can wire by hand in one `wiring.Build` function with the same shape, as `scaffold/` does.

Deployment-specific behavior (auth provider, secret backend, extra routes) is injected into `app.Run(..., Options{...})` as interfaces or factory funcs. A nil option means the open-source default, documented on the field.
