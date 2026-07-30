---
id: "05-public-documentation"
title: "Public documentation"
status: done
wave: 4
depends_on: ["02-desktop-warning-surfaces", "03-mobile-warning-surfaces"]
plan: "plan.md"
spec: "../../specs/ui/ws-connectivity-warning.md"
---

# Task 05: Public documentation

## Acceptance

- Operations docs explain the 3-second yellow and 10-second red states, where the warning appears
  with App status bar enabled/disabled, and that red means live data may be stale.
- Configuration docs keep `features.app_status_bar` accurate: it controls the general-purpose
  surface but not urgent connectivity fallback visibility.
- Public-doc validation passes with no new page or navigation entry.

## Verification

```bash
node --test scripts/validate-public-docs.test.mjs
node scripts/validate-public-docs.mjs
```

## Files likely touched

- `docs/public/operations.md`
- `docs/public/configuration.md`

## Dependencies

Tasks 02 and 03.

## Parallelism

Sequential so terminology matches the landed UI.

## Inputs

- Spec: `What`, `Failure modes`, and `Out of scope`.
- Plan: `Public documentation`.
- Existing Feature toggles and troubleshooting sections in `docs/public/operations.md`.

## Risks

- Do not imply that `features.app_status_bar=false` mounts plugins or metrics during an outage.
- Keep the warning distinct from `/health`, System Status, executor health, and integration health.

## Output contract

Report public docs changed, exact validation results, blockers/risks, then mark this task `done` and
update its checkbox in `plan.md`.
