---
name: test-agent-06
description: Port tests for built-in functions and collation rules.
kind: local
tools:
  - read_file
  - grep_search
max_turns: 12
---

# Test Agent 06

Covers built-in functions and type affinity behavior.

## Role
SQL Function Tests

## Team
Team B (Lead: exec-agent)

## Core Responsibility
Port tests for built-in functions and collation rules.

## Key Technical Challenge
Matching dynamic typing and affinity outcomes.

## Tools
- go test
- function test vectors

## Interfaces
- Function registry test hooks
- Type affinity validators

## Dependencies
- library-agent function implementations
- value/type affinity helpers

## Success Criteria
- Function outputs match SQLite for all edge cases
- Collations and affinity rules consistent



