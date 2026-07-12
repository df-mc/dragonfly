# Redstone Burnout Test Cleanup Design

## Goal

Reduce duplication in `server/block/redstone_test.go` without removing any distinct redstone-torch burnout behavior or changing production code.

## Design

- Introduce focused test helpers for creating a synchronous loaded world, forcing a torch into burnout, waiting for recoverability, and capturing torch/dust state.
- Convert recovery tests that share the same arrange/act/assert flow into table-driven subtests.
- Convert rejection tests that share the same arrange/act/assert flow into table-driven subtests.
- Preserve every existing scenario name, spatial relationship, timing boundary, and assertion as a named subtest.
- Keep the full scheduler-driven loop tests separate because they validate integration behavior rather than helper-level recovery policy.
- Leave production redstone code unchanged.

## Verification

- Run the focused burnout tests before and after the refactor.
- Run `go test ./server/block ./server/world`.
- Run `go test ./...`.
- Confirm the branch diff contains test-only structural changes and no scenario deletions.
