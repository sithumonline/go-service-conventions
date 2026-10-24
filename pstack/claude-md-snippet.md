<!-- go-service-conventions:start -->
## Go services

In any Go service or Go CLI repo, load the go-service-conventions skill before planning.
The repo's own AGENTS.md / CLAUDE.md / rules win: in a repo with its own framework or
layout, apply only the skill's portable rules, never its default stack or scaffold.
When a pstack playbook delegates code-writing, name ~/.claude/skills/go-service-conventions/SKILL.md
in the brief. Treat its done checklist as part of verify and walk it before opening a PR.
If the repo has .agents/playbooks/go-service.md, that playbook's steps take precedence.
<!-- go-service-conventions:end -->
