---
name: test-agent-09
description: Validate file format compatibility at the binary level.
kind: local
tools:
  - read_file
  - grep_search
max_turns: 12
---

# Test Agent 09

Ensures binary file format parity with SQLite.

## Role
File Format Compatibility Tests

## Team
Team B (Lead: exec-agent)

## Core Responsibility
Validate file format compatibility at the binary level.

## Key Technical Challenge
Matching page layout, headers, and WAL formats exactly.

## Tools
- go test
- binary diff utilities

## Interfaces
- File format validators
- Golden database files

## Dependencies
- pager-agent, tree-agent, bridge-agent
- validation-agent guidance

## Success Criteria
- Binary compatibility verified for stored databases
- WAL files match SQLite reference outputs



