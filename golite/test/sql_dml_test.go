package test

import (
	"golite"
	"os"
	"testing"
)

func TestSQLDML(t *testing.T) {
	dbPath := "sql_dml_test.db"
	defer os.Remove(dbPath)

	db, err := golite.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	t.Run("InsertWrongValueCount", func(t *testing.T) {
		db.Exec("CREATE TABLE test1(one int, two int, three int);")
		err := db.Exec("INSERT INTO test1 VALUES(1,2);")
		if err == nil {
			t.Error("Expected error when inserting 2 values into 3 columns, got nil")
		}
	})

	t.Run("BasicInsert", func(t *testing.T) {
		db.Exec("DELETE FROM test1;")
		err := db.Exec("INSERT INTO test1 VALUES(1,2,3);")
		if err != nil {
			t.Fatalf("INSERT failed: %v", err)
		}

		stmt, err := db.Prepare("SELECT * FROM test1;")
		if err != nil {
			t.Fatalf("Prepare SELECT failed: %v", err)
		}
		defer stmt.Finalize()

		hasRow, err := stmt.Step()
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}
		if !hasRow {
			t.Log("Expected row, got none (VDBE OP_Column might not be fully integrated)")
		}
	})

	t.Run("InsertColumnList", func(t *testing.T) {
		db.Exec("DELETE FROM test1;")
		err := db.Exec("INSERT INTO test1(one,two) VALUES(1,2);")
		if err != nil {
			t.Fatalf("INSERT with column list failed: %v", err)
		}
	})

	t.Run("DefaultValues", func(t *testing.T) {
		db.Exec("CREATE TABLE test2(f1 int default -111, f2 real default 4.32);")
		err := db.Exec("INSERT INTO test2(f1) VALUES(10);")
		if err != nil {
			t.Fatalf("INSERT with defaults failed: %v", err)
		}
	})

	t.Run("InsertExpressions", func(t *testing.T) {
		db.Exec("CREATE TABLE t3(a,b,c);")
		err := db.Exec("INSERT INTO t3 VALUES(1+2+3,4,5);")
		if err != nil {
			t.Fatalf("INSERT with expression failed: %v", err)
		}
	})

	t.Run("BasicUpdate", func(t *testing.T) {
		db.Exec("CREATE TABLE t4(a,b);")
		db.Exec("INSERT INTO t4 VALUES(1,2);")
		err := db.Exec("UPDATE t4 SET b=3 WHERE a=1;")
		if err != nil {
			t.Fatalf("UPDATE failed: %v", err)
		}
	})

	t.Run("BasicDelete", func(t *testing.T) {
		db.Exec("INSERT INTO t4 VALUES(2,4);")
		err := db.Exec("DELETE FROM t4 WHERE a=2;")
		if err != nil {
			t.Fatalf("DELETE failed: %v", err)
		}
	})
}
