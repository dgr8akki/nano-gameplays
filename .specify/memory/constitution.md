<!--
Sync Impact Report
==================
Version change: (uninitialized template) → 1.0.0
Modified principles:
  - [PRINCIPLE_1_NAME] → I. CLI Interface
  - [PRINCIPLE_2_NAME] → II. Beautiful TUI
  - [PRINCIPLE_3_NAME] → III. Intelligent Claude Integration
  - [PRINCIPLE_4_NAME] → IV. Simplicity of Use
  - [PRINCIPLE_5_NAME] → V. Virality-Aware Output
Added sections:
  - Core Principles (5 principles defined)
  - Performance Standards (Section 2)
  - Development Workflow & Quality Gates (Section 3)
  - Governance
Removed sections: none
Templates requiring updates:
  - .specify/templates/plan-template.md ✅ (no edits required; Constitution Check
    block is generic and binds dynamically to this file)
  - .specify/templates/spec-template.md ✅ (no edits required; no
    principle-specific guidance hardcoded)
  - .specify/templates/tasks-template.md ✅ (no edits required; no
    principle-specific task categorization hardcoded)
  - .specify/templates/commands/*.md ✅ (no agent-name hardcoding detected)
Follow-up TODOs:
  - README.md ⚠ (does not yet exist; create alongside first feature spec and
    link to this constitution)
  - docs/quickstart.md ⚠ (deferred until first release artifacts exist)
-->

# nano-gameplays Constitution

## Core Principles

### I. CLI Interface

The CLI is the primary, first-class entry point for every capability of
nano-gameplays. Every user-facing operation MUST be executable as a single
command-line invocation with stable, documented flags. Commands MUST follow a
text-in / text-out contract: arguments and stdin define inputs, stdout
produces results, stderr carries diagnostics, and exit codes signal success or
failure unambiguously. Both human-readable and machine-readable (`--json`)
output modes MUST be supported for every command that emits structured data.

**Rationale:** A scriptable CLI guarantees automation, reproducibility, and
composability with other tooling (cron jobs, CI pipelines, batch runs over a
library of gameplay captures) without forcing users into a GUI.

### II. Beautiful TUI

When run interactively, nano-gameplays MUST present a polished terminal user
interface that makes status, progress, and decisions legible at a glance.
Long-running operations (transcoding, analysis, render queues) MUST surface
progress, ETA, and the current pipeline stage. The TUI MUST degrade
gracefully: in non-interactive contexts (pipes, CI) it MUST fall back to plain
text, and it MUST NOT block automation paths defined by Principle I.

**Rationale:** PlayStation gameplay captures are large and processing is
slow; a clear TUI converts opaque waits into trustable feedback, while strict
fallback rules prevent the TUI from compromising scriptability.

### III. Intelligent Claude Integration

Editorial decisions — moment selection, hook crafting, captioning, titling,
hashtags, and pacing — MUST be driven by Claude. Every prompt sent to Claude
MUST be reproducible: prompt template, model ID, and relevant inputs MUST be
recorded with each generated short so a result can be re-derived or audited.
Failures from the model MUST surface explicit errors rather than silently
falling back to heuristics; if a heuristic fallback is added it MUST be
opt-in via an explicit flag.

**Rationale:** The product's differentiator is editorial intelligence, not
clip slicing. Reproducibility is what separates an "AI feature" from a
non-debuggable black box, especially when iterating on virality.

### IV. Simplicity of Use

The default path from a raw gameplay file to publishable shorts MUST require
the smallest reasonable number of decisions from the user — ideally a single
command with sensible defaults. Configuration MUST be optional and layered:
zero-config defaults, then a project config file, then per-invocation flags,
in that order of precedence (later wins). New features MUST justify any added
required flag; if a flag is required, the feature is not done.

**Rationale:** The target user wants shorts produced, not a video pipeline
configured. Friction in the default path is the single biggest threat to
adoption.

### V. Virality-Aware Output

Every generated short MUST be evaluated against an explicit virality rubric
covering at minimum: hook strength in the opening seconds, retention shape,
caption/title fit for YouTube Shorts conventions, and platform-format
correctness (vertical aspect ratio, duration ceiling, audio levels). The
rubric MUST be versioned alongside this constitution; rubric changes are
governance changes and follow the amendment procedure below. Output metadata
MUST record the rubric version used.

**Rationale:** "Smart" is unfalsifiable without a rubric. Versioning the
rubric makes virality a measurable, reviewable property of the system rather
than a marketing claim.

## Performance Standards

nano-gameplays MUST be capable of ingesting a long-form PlayStation gameplay
capture and emitting multiple high-quality YouTube Shorts in a single run.
The pipeline MUST:

- Support batching: a single invocation MUST be able to produce N candidate
  shorts from one source video without re-decoding the source N times.
- Preserve quality: output MUST be at or above the source's effective visual
  quality at the chosen target resolution and bitrate; visible re-encoding
  artifacts on default settings are a defect.
- Parallelize where safe: independent stages (analysis, scoring, rendering of
  distinct candidates) SHOULD run concurrently; shared-resource stages (e.g.
  source decode, GPU encode) MUST be serialized to avoid contention.
- Be measurable: every run MUST emit a machine-readable performance record
  (per-stage wall time, per-short render time, total throughput) so
  regressions are detectable.

A performance regression — measured against the most recent tagged release on
the same input — is a release blocker unless explicitly justified in the PR.

## Development Workflow & Quality Gates

- **Spec-driven flow.** Non-trivial work MUST go through the Spec Kit
  pipeline (`/speckit-specify` → `/speckit-clarify` → `/speckit-plan` →
  `/speckit-tasks` → `/speckit-implement`). Trivial fixes (typos, doc-only
  edits) MAY skip the spec stage but MUST still be reviewed.
- **Constitution Check gate.** Every implementation plan MUST include a
  Constitution Check section that lists, per principle, how the plan
  complies. Violations MUST be enumerated in the plan's Complexity Tracking
  with a justified rationale, or the plan MUST be revised.
- **Reproducibility gate.** Any change touching Claude prompts, the virality
  rubric, or default render settings MUST update the corresponding versioned
  artifact and note the bump in the PR description.
- **CLI/TUI parity gate.** Any new capability MUST ship the CLI surface
  first; TUI exposure is additive and MUST NOT introduce capabilities
  unavailable on the CLI.
- **Performance gate.** Changes to the rendering or analysis pipeline MUST
  publish before/after numbers from the performance record on a reference
  input.
- **Review.** All PRs MUST be reviewed against this constitution. Reviewers
  MUST cite the principle(s) any concern relates to.

## Governance

This constitution supersedes ad-hoc practice. Where another document
conflicts with it, this document wins until the conflict is resolved by
amendment.

**Amendment procedure.** Amendments MUST be proposed as a PR that edits
this file, includes the updated Sync Impact Report, and updates any
dependent templates in the same PR. At least one reviewer other than the
proposer MUST approve. The PR description MUST state the version bump and
its rationale.

**Versioning policy.** This constitution uses semantic versioning:

- **MAJOR** — a principle is removed, redefined in a backward-incompatible
  way, or governance rules are materially relaxed.
- **MINOR** — a new principle or section is added, or guidance is materially
  expanded.
- **PATCH** — clarifications, wording, typo fixes, and non-semantic
  refinements.

The `Last Amended` date MUST be updated whenever this file changes; the
`Ratified` date MUST NOT change after initial adoption.

**Compliance review.** Compliance is verified at two points: (1) at plan
review via the Constitution Check gate above, and (2) at release tagging,
where the release notes MUST confirm no open principle violations exist
(or list them with linked justifications).

**Version**: 1.0.0 | **Ratified**: 2026-05-01 | **Last Amended**: 2026-05-01
