package testutil

import (
	"golite/internal/vfs"
	"os"
)

func NewTestVFS() vfs.VFS {
	return vfs.NewOSVFS()
}

func CreateTempDBFile() (string, func()) {
	f, err := os.CreateTemp("", "golite_test_*.db")
	if err != nil {
		panic(err)
	}
	name := f.Name()
	f.Close()
	return name, func() {
		os.Remove(name)
	}
}
