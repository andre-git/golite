package util

type Pgno uint32
type ErrorCode int

const (
	OK    ErrorCode = 0
	ERROR ErrorCode = 1
	// ... all SQLITE_* codes
)
