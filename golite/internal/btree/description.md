# btree package

Implements a simplified SQLite-style B-Tree on top of the pager. It exposes a `Btree` interface and cursor operations for table traversal and mutation.

## Files

- `btree.go`: core B-Tree implementation, page header parsing, table creation, and cursor navigation.
- `btree_split_test.go`: tests for B-Tree split behavior.
