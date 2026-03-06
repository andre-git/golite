package pager

import (
	"bytes"
	"golite/internal/testutil"
	"os"
	"testing"
)

func TestWALRecovery(t *testing.T) {
	dbPath, cleanup := testutil.CreateTempDBFile()
	defer cleanup()

	vfs := testutil.NewTestVFS()
	
	file, err := vfs.Open(dbPath, os.O_RDWR|os.O_CREATE)
	if err != nil {
		t.Fatalf("failed to open db file: %v", err)
	}
	p := New(file, 4096)
	
	p.Begin(true, false)
	pg1, err := p.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	copy(pg1.Data(), []byte("initial data page 1"))
	p.Write(pg1)
	p.Commit()
	file.Close()

	file, err = vfs.Open(dbPath, os.O_RDWR)
	if err != nil {
		t.Fatal(err)
	}
	p = New(file, 4096)
	
	p.Begin(true, false)
	pg1, _ = p.Get(1)
	copy(pg1.Data(), []byte("updated data page 1"))
	p.Write(pg1)
	
	p.Commit() 
	file.Close()

	file2, err := vfs.Open(dbPath, os.O_RDWR)
	if err != nil {
		t.Fatal(err)
	}
	defer file2.Close()
	p2 := New(file2, 4096)
	
	pg1_recovered, err := p2.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	
	expected := "updated data page 1"
	actual := string(bytes.TrimRight(pg1_recovered.Data()[:len(expected)], "\x00"))
	if actual != expected {
		t.Errorf("Recovery failed: expected %q, got %q", expected, actual)
	}
}
