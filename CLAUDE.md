<!-- SPECKIT START -->
Active plan: [specs/002-claude-cli-backend/plan.md](specs/002-claude-cli-backend/plan.md)

For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
above. Supporting artifacts live alongside it under
`specs/002-claude-cli-backend/`:

- `spec.md` — feature specification
- `research.md` — Phase 0 research (`claude` CLI invocation, frame transport)
- `data-model.md` — entity additions (everything else inherits from 001)
- `contracts/claude-cli.md` — subprocess invocation contract (NEW)
- `contracts/claude-prompt.md` — prompt template + JSON schema (delta from 001)
- `contracts/handoff.md`, `contracts/cli.md` — references to unchanged 001 contracts
- `quickstart.md` — build & run instructions

Predecessor feature 001 (video picker screen) lives under
`specs/001-video-picker-screen/`; its plan and contracts are still load-bearing
for everything except how the Claude call is made.
<!-- SPECKIT END -->
