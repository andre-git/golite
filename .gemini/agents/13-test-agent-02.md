---
name: test-agent-02
description: Validate opcode semantics and register behavior.
kind: local
tools:
  - read_file
  - grep_search
max_turns: 12
---

# Test Agent 02

Focuses on VDBE correctness and opcode coverage.

## Role
VDBE Opcode Tests

## Team
Team B (Lead: exec-agent)

## Core Responsibility
Validate opcode semantics and register behavior.

## Key Technical Challenge
Covering rare opcodes and edge-case flags.

## Tools
- go test
- fuzzing for opcode inputs

## Interfaces
- Opcode test fixtures
- Register state snapshot format

## Dependencies
- exec-agent opcode definitions
- generator-agent bytecode plans

## Success Criteria
- All opcode semantics verified
- No regressions in register state behavior



