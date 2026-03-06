package testutil

import (
	"golite"
	"testing"
)

func FuzzSQL(f *testing.F) {
	seeds := []string{
		"CREATE TABLE t1(a, b);",
		"INSERT INTO t1 VALUES(1, 'hello');",
		"SELECT a, b FROM t1 WHERE a > 10;",
		"CREATE TABLE t2(id INTEGER PRIMARY KEY, data BLOB);",
		"UPDATE t1 SET b = 'world' WHERE a = 1;",
		"DELETE FROM t1;",
		"PRAGMA journal_mode=WAL;",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, sqlStr string) {
		dbPath, cleanup := CreateTempDBFile()
		defer cleanup()

		db, err := golite.Open(dbPath)
		if err != nil {
			return 
		}
		defer db.Close()

		_ = db.Exec(sqlStr)
	})
}
