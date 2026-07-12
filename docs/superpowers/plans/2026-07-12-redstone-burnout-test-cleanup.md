# Redstone Burnout Test Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Reduce redstone-torch burnout test duplication while preserving every existing behavioral scenario.

**Architecture:** Keep scheduler integration tests standalone. Add test-only fixtures for world loading, burnout state, recovery waiting, and snapshots, then express cases with identical control flow as named table rows.

**Tech Stack:** Go testing package, Dragonfly synchronous worlds.

## Global Constraints

- Do not modify production code.
- Preserve every existing scenario and assertion.
- Keep each former test name visible as a subtest name or standalone test.

---

### Task 1: Capture the baseline

**Files:**
- Test: `server/block/redstone_test.go`

- [x] **Step 1:** Record the current burnout-related test names with `go test ./server/block -list 'RedstoneTorch|BurnedOut'`.
- [x] **Step 2:** Run `go test ./server/block -run 'RedstoneTorch|BurnedOut' -count=1` and require a passing baseline.

### Task 2: Extract shared burnout fixtures

**Files:**
- Modify: `server/block/redstone_test.go`

**Interfaces:**
- Produces: test-only helpers that create/close a loaded synchronous world, force burnout, wait for recoverability, and capture state.

- [x] **Step 1:** Replace repeated loaded-world setup with one helper while retaining each test's dimension and loader radius.
- [x] **Step 2:** Replace repeated manual burnout loops with a helper that performs the identical sequence.
- [x] **Step 3:** Run the focused burnout tests and require them to pass.

### Task 3: Table-drive equivalent recovery-policy scenarios

**Files:**
- Modify: `server/block/redstone_test.go`

**Interfaces:**
- Consumes: Task 2 fixture helpers.
- Produces: named table rows corresponding one-to-one with the former standalone recovery/rejection tests.

- [x] **Step 1:** Group local recovery scenarios with identical execution flow into a table-driven test.
- [x] **Step 2:** Group distant/unrelated rejection scenarios with identical execution flow into a table-driven test.
- [x] **Step 3:** Keep scheduler loop, stale-state, zero-position, and source-context tests standalone where their mechanics differ.
- [x] **Step 4:** Compare the resulting named scenario inventory against Task 1 and confirm none were removed.
- [x] **Step 5:** Run focused burnout tests and require them to pass.

### Task 4: Verify and publish

**Files:**
- Modify: `server/block/redstone_test.go`
- Add: `docs/superpowers/specs/2026-07-12-redstone-burnout-test-cleanup-design.md`
- Add: `docs/superpowers/plans/2026-07-12-redstone-burnout-test-cleanup.md`

- [x] **Step 1:** Run `gofmt -w server/block/redstone_test.go`.
- [x] **Step 2:** Run `git diff --check` and require no errors.
- [x] **Step 3:** Run `go test ./server/block ./server/world` and require both packages to pass.
- [x] **Step 4:** Run `go test ./...` and require the full repository to pass.
- [x] **Step 5:** Confirm `git diff --stat` shows only test cleanup and the approved planning documents.
- [x] **Step 6:** Commit and push the branch, then wait for PR checks.
