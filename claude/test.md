# Go_GAPTUI Test Checklist

Run tests in order. Stop on first failure.

## T1. Repository and document setup
- Command:
  - `test -d /Users/hyeokx/git/Go_GAPTUI/.git`
  - `test -f /Users/hyeokx/git/Go_GAPTUI/instructions.md`
  - `test -f /Users/hyeokx/git/Go_GAPTUI/rules.md`
  - `test -f /Users/hyeokx/git/Go_GAPTUI/sequence.md`
  - `test -f /Users/hyeokx/git/Go_GAPTUI/test.md`
- Pass criteria: all commands exit code 0.

## T2. Source-reference coverage checks
- Command:
  - `grep -q "spec.md" /Users/hyeokx/git/Go_GAPTUI/instructions.md`
  - `grep -q "scenario.md" /Users/hyeokx/git/Go_GAPTUI/instructions.md`
  - `grep -q "orderbook_slippage.md" /Users/hyeokx/git/Go_GAPTUI/instructions.md`
  - `grep -q "Gap_Strategy.md" /Users/hyeokx/git/Go_GAPTUI/instructions.md`
  - `grep -q "performance_improvements.md" /Users/hyeokx/git/Go_GAPTUI/instructions.md`
  - `grep -q "ui_improvements.md" /Users/hyeokx/git/Go_GAPTUI/instructions.md`
  - `grep -q "DESIGN_SPEC.md" /Users/hyeokx/git/Go_GAPTUI/instructions.md`
  - `grep -q "claude.md" /Users/hyeokx/git/Go_GAPTUI/instructions.md`
- Pass criteria: all commands exit code 0.

## T3. Sequence integrity checks
- Command:
  - `grep -q "STEP 01" /Users/hyeokx/git/Go_GAPTUI/sequence.md`
  - `grep -q "STEP 11" /Users/hyeokx/git/Go_GAPTUI/sequence.md`
  - `grep -q "delete that step line" /Users/hyeokx/git/Go_GAPTUI/sequence.md`
- Pass criteria: all commands exit code 0.

## T4. Rules integrity checks
- Command:
  - `grep -q "amount matching" /Users/hyeokx/git/Go_GAPTUI/rules.md`
  - `grep -q "quantity matching" /Users/hyeokx/git/Go_GAPTUI/rules.md`
  - `grep -q "snapshot" /Users/hyeokx/git/Go_GAPTUI/rules.md`
  - `grep -q "test.md" /Users/hyeokx/git/Go_GAPTUI/rules.md`
- Pass criteria: all commands exit code 0.

## Result Template
- T1: PASS/FAIL
- T2: PASS/FAIL
- T3: PASS/FAIL
- T4: PASS/FAIL
- Final: PASS only if all tests PASS.
