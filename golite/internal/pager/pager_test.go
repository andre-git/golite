package pager

import (
	"bytes"
	"golite/internal/testutil"
	"testing"
)

func TestPagerGet(t *testing.T) {
	dbPath, cleanup := testutil.CreateTempDBFile()
	defer cleanup()

	vfs := testutil.NewTestVFS()
	file, err := vfs.Open(dbPath, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	pageSize := uint32(4096)
	p := New(file, pageSize)

	pg1, err := p.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	if pg1.Pgno() != 1 {
		t.Errorf("expected pgno 1, got %d", pg1.Pgno())
	}
	if len(pg1.Data()) != int(pageSize) {
		t.Errorf("expected page size %d, got %d", pageSize, len(pg1.Data()))
	}

	pg1_2, err := p.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	if pg1 != pg1_2 {
		t.Error("expected same page object for same pgno")
	}
	p.Release(pg1)
	p.Release(pg1_2)
}

func TestPagerWriteAndCommit(t *testing.T) {
	dbPath, cleanup := testutil.CreateTempDBFile()
	defer cleanup()

	vfs := testutil.NewTestVFS()
	file, err := vfs.Open(dbPath, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	pageSize := uint32(4096)
	p := New(file, pageSize)

	pg, err := p.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	err = p.Write(pg)
	if err == nil {
		t.Error("expected error when writing without transaction")
	}

	if err := p.Begin(true, false); err != nil {
		t.Fatal(err)
	}

	data := pg.Data()
	copy(data, []byte("hello world"))
	if err := p.Write(pg); err != nil {
		t.Fatal(err)
	}

	if err := p.Commit(); err != nil {
		t.Fatal(err)
	}
	p.Release(pg)

	file2, _ := vfs.Open(dbPath, 0)
	defer file2.Close()
	p2 := New(file2, pageSize)
	pg2, err := p2.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(pg2.Data(), []byte("hello world")) {
		t.Errorf("expected 'hello world' in page 1, got %s", pg2.Data()[:20])
	}
	p2.Release(pg2)
}

func TestPagerRollback(t *testing.T) {
	dbPath, cleanup := testutil.CreateTempDBFile()
	defer cleanup()

	vfs := testutil.NewTestVFS()
	file, _ := vfs.Open(dbPath, 0)
	defer file.Close()

	pageSize := uint32(4096)
	p := New(file, pageSize)

	p.Begin(true, false)
	pg, _ := p.Get(1)
	copy(pg.Data(), []byte("initial"))
	p.Write(pg)
	p.Commit()
	p.Release(pg)

	p.Begin(true, false)
	pg, _ = p.Get(1)
	copy(pg.Data(), []byte("modified"))
	p.Write(pg)
	
	if err := p.Rollback(); err != nil {
		t.Fatal(err)
	}
	p.Release(pg)

	pg, _ = p.Get(1)
	if bytes.HasPrefix(pg.Data(), []byte("modified")) {
		t.Error("rollback did not revert in-memory changes")
	}
	if !bytes.HasPrefix(pg.Data(), []byte("initial")) {
		t.Errorf("expected 'initial', got %s", pg.Data()[:10])
	}
	p.Release(pg)
}

func TestPagerPageSize(t *testing.T) {
	dbPath, cleanup := testutil.CreateTempDBFile()
	defer cleanup()

	vfs := testutil.NewTestVFS()
	file, _ := vfs.Open(dbPath, 0)
	defer file.Close()

	p := New(file, 4096)
	if p.PageSize() != 4096 {
		t.Errorf("expected 4096, got %d", p.PageSize())
	}

	if err := p.SetPageSize(8192); err != nil {
		t.Fatal(err)
	}
	if p.PageSize() != 8192 {
		t.Errorf("expected 8192, got %d", p.PageSize())
	}

	pg, _ := p.Get(1)
	if err := p.SetPageSize(4096); err == nil {
		t.Error("expected error changing page size with pages in cache")
	}
	p.Release(pg)
}
