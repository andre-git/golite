# record package

Encodes and decodes SQLite record payloads using varints and serial types. Also defines runtime `Value` types used during record serialization.

## Files

- `record.go`: value types, serial type computation, record encoding, and column decoding.
- `record_test.go`: record encoding/decoding tests.
