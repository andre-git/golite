# lib package

Contains built-in scalar and date/time functions used by SQL evaluation. Values are returned as VDBE-compatible types.

## Files

- `functions.go`: scalar functions (e.g., `abs`, `lower`, `upper`).
- `date.go`: date/time parsing and functions (`date`, `time`, `datetime`, `julianday`, `strftime`).
