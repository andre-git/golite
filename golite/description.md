# golite architecture overview

This folder contains a minimal SQLite-like engine written in Go. The top-level API (`Open`, `Prepare`, `Exec`, `Stmt`) wires together storage, SQL parsing, bytecode generation, and the virtual machine.

## High-level flow

1. `Open` (golite.go) creates an OS-backed VFS, a pager, and a B-Tree, then initializes an in-memory schema.
2. `Prepare` (db.go) tokenizes and parses SQL, then generates VDBE opcodes for the first statement.
3. `Stmt.Step` executes the VDBE program, which uses the B-Tree, record encoding, and pager to read/write pages.

## Main files

- `golite.go`: Database initialization and backend wiring.
- `db.go`: `DB`, `Stmt`, prepare/exec pipeline, and schema updates for CREATE TABLE.
- `example_test.go`, `detailed_example_test.go`: end-to-end usage examples.

## Component map

- `internal/sql`: lexer, parser (AST), and opcode generator.
- `internal/vdbe`: virtual machine executing opcodes.
- `internal/btree`: B-Tree storage and cursor operations on pages.
- `internal/pager`: page cache and transaction boundary management; WAL types.
- `internal/record`: SQLite record encoding/decoding.
- `internal/schema`: in-memory schema model.
- `internal/vfs`: file system abstraction and OS implementation.
- `internal/lib`: scalar and date/time functions used by SQL evaluation.
- `internal/util`: shared primitives (varints, locks, page numbers).
- `internal/testutil` and `test`: test helpers and integration tests.
