# internal package architecture

This directory contains the engine internals split into packages that mirror SQLite subsystems. Each package is responsible for one layer, and the `golite` API composes them.

## Package roles

- `btree`: B-Tree storage and cursor logic built on pager pages.
- `pager`: page cache, transaction lifecycle, and WAL scaffolding.
- `record`: record serialization for row payloads.
- `schema`: in-memory schema structures (tables, columns, indexes).
- `sql`: lexer, parser (AST), and bytecode generation.
- `vdbe`: virtual database engine executing opcodes.
- `vfs`: file system abstraction and OS-backed implementation.
- `lib`: scalar/date functions used by SQL evaluation.
- `util`: shared primitives (varints, locks, error codes).
- `testutil`: testing helpers for file setup and mock pages.
