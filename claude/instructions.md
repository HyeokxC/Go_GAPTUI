# Go_GAPTUI Migration Instructions

## Objective
- Migrate Rust engine responsibilities to Go while keeping Python Textual frontend.
- Preserve existing behavior from `kimchiCEX_arbitrage` documents before adding new features.

## Baseline Scope (must be preserved)
- Real-time ticker ingestion across supported exchanges.
- Orderbook-based slippage/real-gap calculations.
- Scenario detection threads: `GapThreshold`, `DomesticGap`, `FutBasis`.
- Transfer pipeline with 6-step state machine and per-exchange API handling.
- Snapshot-based UI data publishing model and 100ms-class refresh target.

## Migration Principles
- Behavior parity first, optimization second.
- Deterministic numeric handling for money/rates/percentages.
- Keep module boundaries clear: config, exchanges, monitor, scenario, transfer, ipc.
- Keep failure paths explicit (timeout, retry, partial failure, blocked D/W).

## Mandatory Inputs from `claude/` docs
- `spec.md`: architecture/module boundaries/risk controls.
- `task.md`: phase status and unfinished work boundaries.
- `scenario.md`: scenario trigger rules and thread semantics.
- `orderbook_slippage.md`: real-gap/slippage logic and pending scenario integration.
- `Gap_Strategy.md`: hedge matching and strategy entry/exit parameters.
- `performance_improvements.md`: virtualization/snapshot/lock/perf constraints.
- `ui_improvements.md` + `DESIGN_SPEC.md`: frontend rendering intent and visual hierarchy.
- `claude.md`: project summary and strategic constraints.

## Deliverable Rule
- Do not mark migration complete until `test.md` checklist passes.
