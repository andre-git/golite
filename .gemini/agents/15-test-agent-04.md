---
name: test-agent-04
description: Port pager and WAL tests with crash recovery checks.
kind: local
tools:
  - read_file
  - grep_search
max_turns: 12
---

# Test Agent 04

Ensures pager, cache, and WAL correctness.

## Role
Pager and WAL Tests

## Team
Team B (Lead: exec-agent)

## Core Responsibility
Port pager and WAL tests with crash recovery checks.

## Key Technical Challenge
Modeling power loss and partial writes.

## Tools
- go test
- fault injection helpers

## Interfaces
- WAL test fixtures
- Crash recovery harness

## Dependencies
- pager-agent and bridge-agent implementations
- concurrency-agent lock policies

## Success Criteria
- Recovery tests pass across platforms
- WAL checkpointing matches SQLite



