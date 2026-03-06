# sql package

Lexes and parses SQL into an AST, then generates VDBE opcodes for execution. This is the frontend and planner for the engine.

## Files

- `lexer.go`: tokenization of SQL input.
- `parser.go`: token types, AST nodes, and parsing logic for statements and expressions.
- `generator.go`: bytecode generation to VDBE opcodes (SELECT/INSERT/etc.).
- `insert.go`, `update.go`, `delete.go`: statement-specific parsing helpers.
- `parser_test.go`: parser tests.
