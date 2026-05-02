# Specification Quality Checklist: Use Local `claude` CLI Instead of Anthropic SDK

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-05-02
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Success criteria are technology-agnostic (no implementation details)
- [x] All acceptance scenarios are defined
- [x] Edge cases are identified
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Feature meets measurable outcomes defined in Success Criteria
- [x] No implementation details leak into specification

## Notes

- The spec names the `claude` command-line tool by name. This is not an
  implementation choice that leaks into the spec — it is the explicit
  user request and the entire point of the feature, so it is treated as
  product behaviour rather than a technical detail.
- "Anthropic SDK" and `ANTHROPIC_API_KEY` are referenced because the
  feature is *removing* them; describing what is being removed without
  naming it would be unintelligible. They appear only in
  removal-context functional requirements, not as architecture.
- A reasonable default was applied to the frame-transport question
  (per-user temp directory, deleted before goroutine exit) and recorded
  in Assumptions rather than as a [NEEDS CLARIFICATION] marker, because
  that default preserves the prior feature's FR-023 intent and no other
  reasonable interpretation materially changes scope or UX.
- Items marked incomplete require spec updates before `/speckit-clarify`
  or `/speckit-plan`.
