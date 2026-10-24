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

### Apply it in every session (recommended)

The skill loads when a task matches its description. To make Claude apply it every time, and pass it to pstack's subagents, add a short block to your global `~/.claude/CLAUDE.md`:

```bash
~/.claude/skills/go-service-conventions/pstack/install-global.sh
```

Running the script again replaces the block instead of adding a second copy. The block text is in `pstack/claude-md-snippet.md`.

## Use with pstack

[pstack](https://github.com/michael-denyer/pstack-claude) routes work through playbooks (Feature, Refactoring, Bug fix). Its project playbooks let a repo add steps to those without forking pstack. `pstack/go-service.md` is one. It makes pstack:

- load this skill before `how` on a Feature, and start new services from `scaffold/`
- name this skill's `SKILL.md` in every code-writing delegate's brief
- run this skill's done checklist as part of verify
- use the controller → service → repository layering as the refactoring target shape
- check error mapping first when fixing a bug

The global `CLAUDE.md` block above already applies the conventions in pstack runs. A project playbook ties each rule to a specific pstack step instead, and pstack's checker catches drift after an upgrade. Use both for repos you work in often.

Install it into a Go repo that uses pstack (the script also runs pstack's playbook checker):

```bash
~/.claude/skills/go-service-conventions/pstack/install.sh /path/to/your/go/repo
```

Commit the resulting `.agents/playbooks/go-service.md` so the team gets it. Rerun the script after a pstack upgrade. The checker fails if pstack reworded a step this playbook anchors on.

Project playbooks need a pstack build that has them (the Claude Code port 0.9.86 does). The `backnotprop/pstack` mirror does not yet.

## Layout

```text
SKILL.md                         rules, golden paths, done checklist
references/service.md            layout, layers, routes, DI
references/errors-and-responses.md
references/platform.md           config, logging, bootstrap, DB, outbound clients
references/testing-and-lint.md
references/cli.md                cobra CLI conventions
scaffold/                        compilable reference service
pstack/                          pstack project playbook, global CLAUDE.md block, installers
```

## Verify the scaffold

```bash
cd scaffold
make check                       # build, unit tests, lint, integration-tag build
TEST_DATABASE_DSN=postgres://... make test-integration
```

## License

Apache-2.0, matching the codebase it was derived from.
