---
name: test-agent-08
description: Port query planner and bytecode plan tests.
kind: local
tools:
  - read_file
  - grep_search
max_turns: 12
---

# Test Agent 08

Verifies planner output and plan determinism.

## Role
Query Planner Tests

## Team
Team B (Lead: exec-agent)

## Core Responsibility
Port query planner and bytecode plan tests.

## Key Technical Challenge
Ensuring plan selection parity.

## Tools
- go test
- plan diff tooling

## Interfaces
- Plan comparison harness
- Query plan snapshot format

## Dependencies
- generator-agent planner output
- exec-agent VDBE expectations

## Success Criteria
- Query plans match SQLite for benchmark cases
- Planner regression tests pass



