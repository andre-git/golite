---
name: test-agent-07
description: Port locking, busy handler, and concurrency stress tests.
kind: local
tools:
  - read_file
  - grep_search
max_turns: 12
---

# Test Agent 07

Focuses on concurrency and locking stress cases.

## Role
Concurrency and Busy Tests

## Team
Team B (Lead: exec-agent)

## Core Responsibility
Port locking, busy handler, and concurrency stress tests.

## Key Technical Challenge
Reproducing contention without flakiness.

## Tools
- go test
- race detector

## Interfaces
- Concurrency stress harness
- Busy handler test hooks

## Dependencies
- concurrency-agent policies
- bridge-agent VFS locking

## Success Criteria
- Busy handler behavior matches SQLite
- No deadlocks or data races in tests



