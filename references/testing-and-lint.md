# Testing and lint

## Two tiers, chosen by the build tag on line 1

| Tier | Tag | Touches | Covers |
|---|---|---|---|
| Unit | none | nothing external | service logic: error mapping, validation gates, branching, transformation, fan-out |
| Integration | `//go:build integration` | real Postgres | repository SQL, transactions, constraints |

- Never unit-test a repository by mocking it.
- Code with no seam (tickers calling a global DB, concrete vault client) gets unit tests for the branches before the boundary and integration tests for the rest.
- Unit test files are named `<subject>_unit_test.go` to make the tier obvious.
- `go build -tags=integration ./...` must compile in CI, because helper names can collide across tiers.

## Mocks

- Generate with `moq` from a `//go:generate` line above each interface. Pin the moq version in the Makefile and check it there, since versions reformat imports. moq v0.5.x cannot load packages on Go 1.27+. Use v0.8.1 or later there.
- Mocks live in `repositories/repomocks/` and `clients/clientmocks/`. Fields are `<Method>Func`.
- An unconfigured method panics. Leave a func nil to assert that path is never reached, and say so in a comment.
- In-package interfaces without generated mocks get a hand-written stub in the test file with the same `<Method>Func` shape.

## Service unit test shape

```go
func TestWidgetService_GetWidget(t *testing.T) {
	const org, name = "acme", "w1"

	t.Run("maps record-not-found to ErrWidgetNotFound", func(t *testing.T) {
		repo := &repomocks.WidgetRepositoryMock{
			GetWidgetFunc: func(_ context.Context, _, _ string) (*models.Widget, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}
		_, err := NewWidgetService(repo).GetWidget(context.Background(), org, name)
		assert.ErrorIs(t, err, utils.ErrWidgetNotFound)
	})

	t.Run("does not mask a real error as not-found", func(t *testing.T) {
		boom := errors.New("connection reset")
		repo := &repomocks.WidgetRepositoryMock{
			GetWidgetFunc: func(_ context.Context, _, _ string) (*models.Widget, error) { return nil, boom },
		}
		_, err := NewWidgetService(repo).GetWidget(context.Background(), org, name)
		assert.ErrorIs(t, err, boom)
		assert.NotErrorIs(t, err, utils.ErrWidgetNotFound)
	})
}
```

- `TestType_Method` with `t.Run` subtests named as sentences of expected behavior.
- testify: `require` for preconditions that make the rest meaningless, `assert` otherwise. `assert.ErrorIs` for sentinels.
- Inject mocks through the real constructor, never by poking struct fields.
- Every service method that maps not-found has a test proving a real error is not masked.
- Shared helpers (`strPtr`, `discardLogger`) are declared once per package. Reuse them.
- One file per package may be marked as the reference test in its header comment so new tests copy its shape.

## HTTP handler tests

Use `httptest.NewRecorder()` and `httptest.NewRequest`, set path values with `req.SetPathValue`, and assert status plus the decoded `code` field.

## Invariant tests

Encode cross-cutting rules as tests over the real registration code rather than docs:

- every mutating route declares a permission
- every permission is granted by a role
- every audited action has a schema

These run in the unit tier, so the build fails when someone forgets.

## Lint (`golangci-lint` v2)

Enable at least: `errorlint`, `nilerr`, `nilnil`, `nolintlint`, `goheader`, `grouper`, `reassign`, `wastedassign`, `exhaustruct` (for models, spec, wiring; excluded for repositories and `_test.go`). Formatters: `gofmt`, `gofumpt`, `goimports` with `local-prefixes` set to the module path. See `scaffold/.golangci.yaml`. golangci-lint 2.13+ warns that `exhaustruct` is deprecated in favor of `exhaustruct_v5`. The old linter still runs.

Rules the linters enforce:

- Never `return nil, nil`. Return an empty typed value. Suppress only with a reason.
- `//nolint:<linter> // <reason>`. A bare `//nolint` fails.
- Every `.go` file starts with the license header.
- Imports in three groups: stdlib, third-party, then this module.

Run the exact CI command with the CI config file. Don't trust a local `make lint` that uses a different config.
