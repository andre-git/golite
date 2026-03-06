package test

import (
	"golite/internal/btree"
	"golite/internal/pager"
	"golite/internal/testutil"
	"testing"
)

func TestBtreeCore(t *testing.T) {
	path, cleanup := testutil.CreateTempDBFile()
	defer cleanup()

	vfs := testutil.NewTestVFS()
	file, err := vfs.Open(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	p := pager.New(file, 1024)
	bt := btree.New(p)

	t.Run("CreateTable", func(t *testing.T) {
		err := bt.BeginTrans(true)
		if err != nil {
			t.Fatalf("BeginTrans failed: %v", err)
		}
		
		pgno, err := bt.CreateTable(btree.BTREE_INTKEY)
		if err != nil {
			bt.Rollback()
			t.Fatalf("CreateTable failed: %v", err)
		}
		
		if pgno == 0 {
			t.Errorf("expected valid pgno, got 0")
		}

		err = bt.Commit()
		if err != nil {
			t.Fatalf("Commit failed: %v", err)
		}
	})

	t.Run("CursorOperations", func(t *testing.T) {
		err := bt.BeginTrans(false)
		if err != nil {
			t.Fatalf("BeginTrans failed: %v", err)
		}
		defer bt.Rollback()

		cur, err := bt.Cursor(2, false)
		if err != nil {
			t.Fatalf("Cursor creation failed: %v", err)
		}
		defer cur.Close()

		if err := cur.First(); err != nil {
			t.Errorf("First() failed: %v", err)
		}
		
		if err := cur.Last(); err != nil {
			t.Errorf("Last() failed: %v", err)
		}

		if err := cur.Next(); err == nil {
			t.Errorf("Next() on empty table should return error/EOF")
		}
	})
}
