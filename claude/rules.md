# Go_GAPTUI Rules

## Engineering Rules
- Preserve existing semantics from source docs; no speculative behavior changes.
- Implement smallest valid unit per step; validate before next step.
- Keep logs structured and actionable for data-feed, scenario, and transfer failures.
- Separate read path (snapshot/view) from write path (collectors/processors/executors).

## Strategy and Risk Rules
- Follow hedge method split from strategy doc:
  - transfer involved -> amount matching
  - no transfer -> quantity matching
- Enforce pre-trade checks: balance, D/W status, network availability, minimum amount.
- Enforce circuit breaker handling for stuck transfers/timeouts.

## Scenario Rules
- Keep implemented scenarios active: `GapThreshold`, `DomesticGap`, `FutBasis`.
- Keep planned scenarios excluded unless explicitly added in future step.
- Preserve per-key thread behavior (create, sub-entry update, close).

## Performance Rules
- Maintain snapshot-swap style data publication.
- Minimize lock contention and unnecessary allocations.
- Keep refresh/publish cadence aligned with documented 100ms/200ms expectations.

## Execution Rules
- Work step-by-step using `sequence.md`.
- After each step is verified, remove that step line from `sequence.md`.
- Final validation must follow `test.md` exactly.
