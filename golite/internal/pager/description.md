# pager package

Provides page caching, transactional read/write boundaries, and page allocation. The pager is the bridge between B-Tree logic and the VFS file API.

## Files

- `pager.go`: page cache, `Get/Write/Release`, transaction state, and file IO.
- `wal.go`: WAL headers and frame layout, plus read helpers (write/ckpt are stubbed).
- `pager_test.go`: pager lifecycle and persistence tests.
