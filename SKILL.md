---
name: go-service-conventions
description: Portable rules plus a default stack for writing Go backend services and Go CLIs in a layered controller → service → repository style with sentinel errors, a stable JSON error envelope, per-route authorization, env-driven config validated at startup, slog request loggers, moq mocks, and a cobra CLI built on a Factory + Options + runX(ctx, opts) shape. Repo-documented conventions always take precedence; only the portable rules apply in repos with their own framework. Use when starting a new Go service or CLI, rewriting one into this style, adding an endpoint/resource/command, reviewing Go code for consistency with these conventions, or when the user says "follow the go service conventions", "scaffold a service like agent-manager", or "write it the same way".
---

# Go service conventions

Distilled from a production control-plane API, its trace API, and its CLI. Copy shapes from `scaffold/`, which compiles and passes its tests.

## Precedence: check the repo first

Before applying anything here, read the repo's own guides (`AGENTS.md`, `CLAUDE.md`, `.claude/rules/`, a lint config with `depguard`, framework skills).

- **The repo names its own framework, layout, or packages** (RonyKit, Gin, Echo, sqlc, fx, a mandated errors or logging package): the repo wins on every point it covers. Apply only the **portable rules** below, written in the repo's own packages. Do not import this skill's stack, folders, scaffold, or golden paths.
- **The repo states a rule that contradicts a portable rule** (for example "optional dependencies never crash startup"): the repo wins. Say which rule you skipped and why.
- **New repo, or an existing repo with no stated conventions**: apply everything here, starting from `scaffold/`.

## Pick the reference

| Task | Read |
|---|---|
| New service, or a new resource/endpoint | `references/service.md`, then copy from `scaffold/` |
| Errors, status codes, response bodies | `references/errors-and-responses.md` |
| Config, logging, bootstrap, shutdown, DB, outbound HTTP clients | `references/platform.md` |
| Tests, mocks, lint | `references/testing-and-lint.md` |
| A cobra CLI or a new command | `references/cli.md` |

## Starting a new service from `scaffold/`

`scaffold/` is a working widget CRUD service (module `example.com/widgetsvc`). To start a new one, copy it, then replace the module path in `go.mod` and every import, the org name in `.github/copyright_header.tmpl` and every file header, and the `local-prefixes` in `.golangci.yaml`. Rename the widget resource last. `make check` must pass before you add anything.

It wires DI by hand in `wiring/wiring.go`. Switch to `google/wire` once there are more than a handful of resources. `middleware.TokenVerifier` has no production implementation. Inject a JWKS verifier in `main.go`.

## Running under pstack

When a pstack playbook delegates code-writing, put this file's absolute path and the one matching reference in the brief, and tell the delegate to read both first. This file's done checklist is part of verify, and gets walked again before opening a PR. A repo's `.agents/playbooks/go-service.md` (installed by `pstack/install.sh`) ties these to specific pstack steps and takes precedence.

## Portable rules (any Go repo, written in its own packages)

1. **Not-found and real errors stay distinct.** Map the persistence layer's not-found to a domain sentinel in the business layer and wrap everything else with `%w` and context. Never turn an unexpected error into not-found. Never fall back to a default on a real error. Compare with `errors.Is` / `errors.As` only. A test proves a real error is not masked.
2. **Stable machine-readable error codes.** Whatever the error type, clients branch on a code that is added, never renamed or removed. Internal detail never reaches a 500 response body.
3. **Tenant from verified identity, never from the request.** The service also checks the tenant against the resource it loads. A missing tenant is an error, not a wildcard.
4. **Every operation declares its authorization.** No unauthenticated endpoint without a written reason.
5. **`context.Context` first on every I/O method, propagated all the way.** Never pass `nil`.
6. **Concurrency.** Never hold a lock across I/O. Atomic upserts, not read-then-write. Serialize expensive side effects per key, not globally.
7. **Config problems surface early and together.** Check co-dependent values together. Secrets have no working defaults.
8. **Logs carry correlation context** (request ID, tenant, resource ID). Never log a secret or a request body.
9. **Generated code is never hand-edited.** Regenerate and commit.
10. **Thin transport, logic in the business layer, persistence behind an interface** so business logic is unit-testable with mocks.

## Default stack (only when the repo states none)

1. **Three layers, each behind an interface.** Controller (HTTP only) → service (business logic) → repository (persistence only). Each layer receives the interface of the layer below through a `NewXxx(deps...) Xxx` constructor that returns the interface and hides an unexported struct.
2. **Contract first.** The OpenAPI spec is the source of truth. Request and response types are generated into `spec/` and never hand-edited. Regenerate and commit.
3. **Sentinel errors, mapped once per layer.** The repository returns the driver error (`gorm.ErrRecordNotFound`). The service maps it to a domain sentinel (`utils.ErrWidgetNotFound`) and wraps everything else with `fmt.Errorf("failed to <verb> <noun>: %w", err)`. The controller maps sentinels to status codes in one `handleCommonErrors` switch. Never turn an unexpected error into not-found. Never fall back to a default on a real error. Compare with `errors.Is` and `errors.As` only.
4. **One error envelope.** Every error response is `{"message","code","reason"}` with a stable machine-readable `code`. Codes are a wire contract: add, never rename or remove.
5. **Authorization is declared at route registration.** Each route names its own `resource:verb` permission. No route is unauthenticated without a written reason. Every permission is granted by at least one role.
6. **Tenant from the token, never from the request.** The org/tenant ID comes from verified claims. The service also checks it against the loaded resource. A missing tenant is an error, not a wildcard. Every tenant-scoped table has the same tenant column.
7. **`context.Context` first on every I/O method, propagated all the way.** Repositories call `db.WithContext(ctx)`. Clients use `http.NewRequestWithContext`. Never pass `nil`.
8. **Config is env vars, read and validated once at startup.** Collect every problem, log them together, exit non-zero. Check co-dependent values together. Secrets have no defaults.
9. **Request-scoped `slog` logger from the context.** `log := logger.GetLogger(ctx)` with method, path and correlation ID already attached. Debug for hot paths and client mistakes, Info for rare events, Error for failures and destructive operations.
10. **Concurrency.** Never hold a lock across I/O. Use atomic upserts (`ON CONFLICT`), not read-then-write. Serialize expensive side effects per key, not globally.
11. **Generated code is never hand-edited.** `spec/`, mocks, and DI output are regenerated by a make target and committed.
12. **Every file carries the license header.** The linter enforces it.

## Golden path: add a resource to a service

1. Edit the OpenAPI spec. Regenerate `spec/`.
2. Add the permission constants (`widget:create`, `widget:read`, ...) and grant them to roles.
3. Model in `models/` with explicit `gorm:"column:..."` tags and a `TableName()` method. Add a numbered migration.
4. Repository interface with a `//go:generate moq` directive, then the impl.
5. Service interface and impl. Map errors here.
6. Controller interface and impl. Parse, validate, call, write.
7. `register<Resource>Routes(rr, ctrl)` in `api/`, one line per route with its permission. Call it from the single route-list function.
8. Add constructors to the DI provider sets. Regenerate DI and mocks.
9. Service unit test with generated mocks covering each error branch. Repository test with the `integration` build tag against a real DB.
10. Run the done checklist.

## Golden path: add a CLI command

`NewXxxCmd(f *cmdutil.Factory)` builds `XxxOptions` from Factory fields as direct references. `RunE` resolves flags and scope, then calls `runXxx(cmd.Context(), opts)`. `runXxx` never sees cobra. All output goes through `IOStreams`. Every error returns through `render.Error(opts.IO, scope, err)`. Test `runXxx` with a hand-built options struct. Details in `references/cli.md`.

## Done checklist

When the repo has its own gate (`make verify`, a verify hook), run that gate instead of the stack-specific items here, then check the portable ones.

- [ ] Unit tests pass. `go build -tags=integration ./...` compiles.
- [ ] The CI lint command passes with the CI config, not a local default.
- [ ] `gofmt`/`gofumpt`/`goimports` clean.
- [ ] Generated files regenerated and committed.
- [ ] New routes declare a permission. New permissions are granted by a role.
- [ ] New config is validated at startup.
- [ ] New I/O takes and propagates `ctx`. New errors wrap with `%w`.
- [ ] Not-found and real errors stay distinct, and a test asserts it. (portable)
- [ ] Tenant comes from verified identity and is checked against the loaded resource. (portable)
