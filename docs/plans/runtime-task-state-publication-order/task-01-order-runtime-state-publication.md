---
id: "01-order-runtime-state-publication"
title: "Order runtime state publication"
status: pending
wave: 1
depends_on: []
plan: "plan.md"
spec: "../../specs/tasks/runtime-state-publication-order.md"
---

# Task 01: Order runtime state publication

## Acceptance

- A successful `WAITING_FOR_INPUT -> RUNNING` session transition reconciles an
  eligible task to `IN_PROGRESS` before publishing
  `session.state_changed(RUNNING)`.
- Existing archive, Office, clarification/cancellation, terminal-state, and
  session-state CAS guards remain intact; reconciliation failure does not hide
  the running session.
- Already-`RUNNING` stream churn performs no redundant task-state writes or
  state events, and the executor-success reconciliation remains available.

## Verification

```bash
cd apps/backend && go test \
  -run 'TestSetSessionRunning_(PublishesTaskStateBeforeRunningSession|NoRedundantTaskWrites|WritesOnTransition)' \
  ./internal/orchestrator
cd apps/backend && go test -race \
  -run 'Test(SetSessionRunning_PublishesTaskStateBeforeRunningSession|ReconcileTaskStateForRuntime_)' \
  ./internal/orchestrator
cd apps/backend && go test \
  -run 'TestUpdateTaskStateIfSessionState_' \
  ./internal/task/repository/sqlite
```

## Files likely touched

- `apps/backend/internal/orchestrator/event_handlers_streaming.go`
- `apps/backend/internal/orchestrator/event_handlers_streaming_test.go`

## Dependencies

None.

## Parallelism

Sequential. The production ordering and its regression test share the same
orchestrator lifecycle seam.

## Inputs

- Spec `What`, `Failure modes`, and `Scenarios`.
- ADR
  `docs/decisions/2026-07-30-runtime-task-state-before-running-event.md`.
- Existing `updateTaskSessionStateWithHook` pre-publication seam.
- Existing `reconcileTaskStateForRuntimeLocked` guarded task-state writer.
- Existing no-redundant-write and runtime-state race tests.

## Output contract

Report the RED failure reason, GREEN results, files changed, exact commands and
results, remaining risks, and the task/plan status updates. Do not change
frontend grouping or synthesize task state from session events.
