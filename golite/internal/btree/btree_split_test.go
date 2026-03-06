package btree

import (
	"fmt"
	"golite/internal/pager"
	"golite/internal/testutil"
	"golite/internal/util"
	"testing"
)

func VerifyTreeIntegrity(t *testing.T, bt Btree, root util.Pgno) {
	cur, err := bt.Cursor(root, false)
	if err != nil {
		t.Fatalf("Failed to create cursor for integrity check: %v", err)
	}
	defer cur.Close()

	var lastKey []byte
	count := 0
	err = cur.First()
	if err != nil {
		if err.Error() == "EOF" {
			return 
		}
		t.Fatalf("First() failed: %v", err)
	}

	for {
		key, err := cur.Key()
		if err != nil {
			t.Errorf("Failed to get key at index %d: %v", count, err)
			break
		}

		if lastKey != nil {
			if string(key) <= string(lastKey) {
				t.Errorf("Integrity check failed: key %s is not greater than last key %s", string(key), string(lastKey))
			}
		}

		lastKey = make([]byte, len(key))
		copy(lastKey, key)
		count++

		err = cur.Next()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Errorf("Next() failed at index %d: %v", count, err)
			break
		}
	}
}

func TestBtreeLargeScaleInsertions(t *testing.T) {
	path, cleanup := testutil.CreateTempDBFile()
	defer cleanup()

	vfs := testutil.NewTestVFS()
	file, err := vfs.Open(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	pageSize := 512
	p := pager.New(file, uint32(pageSize))
	bt := New(p)

	err = bt.BeginTrans(true)
	if err != nil {
		t.Fatalf("BeginTrans failed: %v", err)
	}

	root, err := bt.CreateTable(BTREE_INTKEY)
	if err != nil {
		t.Fatalf("CreateTable failed: %v", err)
	}

	cur, err := bt.Cursor(root, true)
	if err != nil {
		t.Fatalf("Cursor creation failed: %v", err)
	}
	defer cur.Close()

	// Position the cursor at the first record (even if it's EOF)
	// This ensures the cursor stack is populated.
	_ = cur.First()

	numRecords := 5
	for i := 0; i < numRecords; i++ {
		key := []byte(fmt.Sprintf("key-%05d", i))
		payload := []byte(fmt.Sprintf("data-payload-%05d", i))
		
		err = cur.Insert(key, payload)
		if err != nil {
			t.Fatalf("Insert failed at iteration %d: %v", i, err)
		}
	}

	if err := bt.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	if err := bt.BeginTrans(false); err != nil {
		t.Fatal(err)
	}
	VerifyTreeIntegrity(t, bt, root)
	bt.Rollback()
}
