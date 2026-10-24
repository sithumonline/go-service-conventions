---
extends: feature, refactoring, bug-fix
when: Building, reshaping, or fixing a Go backend service or Go CLI that follows the go-service-conventions skill.
---

# Go service conventions on top of pstack

Resolve the absolute path of the **go-service-conventions** skill's `SKILL.md` once (usually `~/.claude/skills/go-service-conventions/SKILL.md` or `.claude/skills/go-service-conventions/SKILL.md`). Every brief below names that path.

- **Before** "`how` over the affected subsystem."
  Load the **go-service-conventions** skill. Pick the reference its table names for this task. A brand-new service starts as a copy of its `scaffold/`, renamed, with `make check` green before any feature work.
- **In** "Delegate code-writing to a subagent"
  The brief names the conventions `SKILL.md` path and the one reference file that applies, and requires the delegate to read both before writing code. The data shape the brief names is the model, the service interface, and the sentinel errors, chosen before any handler code.
- **After** "Verify on the matching surface."
  Run the conventions' done checklist: unit tests, `go build -tags=integration ./...`, the CI lint command with the CI config file, and regenerated spec and mocks committed. A missing check is a failed verify, not a skip.
- **In** "Name the target shape."
  The target shape is the controller → service → repository layering in the conventions' `references/service.md`. Name which layer each moved function lands in.
- **In** "Prove behavior is unchanged on the real artifact"
  For HTTP surfaces, the pin includes status codes and the error `code` field for every sentinel the reshaped code can return.
- **In** "Plan the fix."
  Check first whether the bug is a mis-mapped error (real error answered as not-found, sentinel compared with `==`, detail leaked in a 500). The fix lands in the layer the conventions assign that mapping to, with a service unit test that fails before the fix.
- **Before** "Run **Opening a PR**."
  Walk the conventions' done checklist in `SKILL.md` and paste each item's result into the PR description.
