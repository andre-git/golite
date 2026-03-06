# test package

Integration-style tests for the engine. These tests exercise the public API and the internal subsystems together.

## Files

- `btree_test.go`: B-Tree creation and cursor behavior.
- `pager_test.go`: pager cache and persistence tests.
- `sql_dml_test.go`: INSERT/UPDATE/DELETE behavior.
- `sql_query_test.go`: SELECT and basic join scenarios.
- `vdbe_test.go`: VDBE opcode execution unit tests.
- `sql_dml_test.db`, `sql_select_test.db`: test fixtures produced during runs.
