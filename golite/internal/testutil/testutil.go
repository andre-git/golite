package testutil

import (
	"encoding/binary"
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

func SetupMockDB(path string, pageSize int, pages map[uint32][]byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	maxPg := uint32(1)
	for pgno := range pages {
		if pgno > maxPg {
			maxPg = pgno
		}
	}

	for i := uint32(1); i <= maxPg; i++ {
		data, ok := pages[i]
		if !ok {
			data = make([]byte, pageSize)
		}
		
		if i == 1 && len(data) >= 100 && string(data[:16]) != "SQLite format 3\x00" {
			copy(data[:16], "SQLite format 3\x00")
			binary.BigEndian.PutUint16(data[16:], uint16(pageSize))
		}

		if _, err := f.WriteAt(data, int64(i-1)*int64(pageSize)); err != nil {
			return err
		}
	}
	return nil
}

func WriteBTreeLeafPage(pageSize int, isIntKey bool, cells []byte) []byte {
	pg := make([]byte, pageSize)
	if isIntKey {
		pg[0] = 0x0D 
	} else {
		pg[0] = 0x0A 
	}
	binary.BigEndian.PutUint16(pg[5:], uint16(pageSize)) 
	return pg
}
