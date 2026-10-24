# go-service-conventions

A Claude Code skill that teaches Claude one consistent way to write Go backend services and Go CLIs. It was distilled from the [WSO2 Agent Manager](https://github.com/wso2/agent-manager) codebase (control-plane API, trace API, and the `amctl` CLI).

It covers:

- controller → service → repository layering behind interfaces, wired by constructors
- contract-first OpenAPI with generated request and response types
- sentinel errors, a `ValidationError` type, and a stable `{message, code, reason}` error envelope
- per-route `resource:verb` authorization, with the tenant taken from the token
- env-var config validated once at startup, request-scoped `slog` loggers, graceful shutdown
- unit tests with `moq` mocks, integration tests behind a build tag, and invariant tests
- a strict `golangci-lint` v2 config
- a cobra CLI built on Factory + Options + `runX(ctx, opts)`, with a JSON output contract

`scaffold/` is a small, working service that follows every rule. Its unit tests pass and it lints clean.

## Install

Personal (every project on your machine):

```bash
git clone https://github.com/sithumonline/go-service-conventions ~/.claude/skills/go-service-conventions
```

One project (shared with the team through the repo):

```bash
git clone https://github.com/sithumonline/go-service-conventions .claude/skills/go-service-conventions
rm -rf .claude/skills/go-service-conventions/.git
```

Claude loads it when you start or change a Go service or CLI. You can also invoke it with `/go-service-conventions`.

## Layout

```text
SKILL.md                         rules, golden paths, done checklist
references/service.md            layout, layers, routes, DI
references/errors-and-responses.md
references/platform.md           config, logging, bootstrap, DB, outbound clients
references/testing-and-lint.md
references/cli.md                cobra CLI conventions
scaffold/                        compilable reference service
```

## Verify the scaffold

```bash
cd scaffold
make check                       # build, unit tests, lint, integration-tag build
TEST_DATABASE_DSN=postgres://... make test-integration
```

## License

Apache-2.0, matching the codebase it was derived from.
