package test

import (
	"bytes"
	"golite/internal/pager"
	"golite/internal/testutil"
	"testing"
)

func TestPagerCacheLifecycle(t *testing.T) {
	dbPath, cleanup := testutil.CreateTempDBFile()
	defer cleanup()

	vfs := testutil.NewTestVFS()
	file, err := vfs.Open(dbPath, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	pageSize := uint32(4096)
	p := pager.New(file, pageSize)

	p.Begin(true, false)
	pg1, err := p.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	p.Write(pg1)
	p.Commit()
	p.Release(pg1)

	for i := 0; i < 10; i++ {
		p.Begin(false, false)
		pg, err := p.Get(1)
		if err != nil {
			t.Fatal(err)
		}
		p.Release(pg)
	}
}

func TestPagerJournalBaseline(t *testing.T) {
	dbPath, cleanup := testutil.CreateTempDBFile()
	defer cleanup()

	vfs := testutil.NewTestVFS()
	file, _ := vfs.Open(dbPath, 0)
	defer file.Close()

	p := pager.New(file, 4096)

	p.Begin(true, false)
	pg, _ := p.Get(1)
	testData := []byte("persistence test")
	copy(pg.Data(), testData)
	p.Write(pg)
	p.Commit()
	p.Release(pg)

	file2, _ := vfs.Open(dbPath, 0)
	defer file2.Close()
	p2 := pager.New(file2, 4096)
	pg2, err := p2.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(pg2.Data(), testData) {
		t.Errorf("expected %s, got %s", testData, pg2.Data()[:len(testData)])
	}
	p2.Release(pg2)
}
