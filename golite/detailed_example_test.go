package golite_test

import (
	"fmt"
	"golite"
	"os"
	"testing"
)

// TestDetailedExample showcases the current functional capabilities of golite.
// It covers CREATE TABLE, INSERT, and basic SQL execution with transactions.
func TestDetailedExample(t *testing.T) {
	dbPath := "detailed_example.db"
	defer os.Remove(dbPath)

	fmt.Println("--- Golite Detailed Showcase ---")

	// 1. Open a new database file
	fmt.Printf("1. Opening database: %s\n", dbPath)
	db, err := golite.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 2. Create a table
	fmt.Println("2. Creating table 'users'...")
	err = db.Exec("CREATE TABLE users (id INT, name TEXT);")
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}

	// 3. Insert data into the table
	fmt.Println("3. Inserting records...")
	insertStatements := []string{
		"INSERT INTO users VALUES (1, 'Alice');",
		"INSERT INTO users VALUES (2, 'Bob');",
	}

	for _, sql := range insertStatements {
		fmt.Printf("   Executing: %s\n", sql)
		err = db.Exec(sql)
		if err != nil {
			t.Fatalf("INSERT failed: %v", err)
		}
	}

	// 4. Use a Transaction
	fmt.Println("4. Running a transaction...")
	err = db.Exec("BEGIN;")
	if err != nil {
		t.Fatalf("BEGIN failed: %v", err)
	}

	fmt.Println("   Inserting 'Charlie' inside transaction...")
	err = db.Exec("INSERT INTO users VALUES (3, 'Charlie');")
	if err != nil {
		db.Exec("ROLLBACK;")
		t.Fatalf("INSERT in transaction failed: %v", err)
	}

	fmt.Println("   Committing transaction...")
	err = db.Exec("COMMIT;")
	if err != nil {
		t.Fatalf("COMMIT failed: %v", err)
	}

	// 5. Basic SELECT (Verification)
	fmt.Println("5. Verifying with SELECT (Parser/Lexer check)...")
	stmt, err := db.Prepare("SELECT * FROM users;")
	if err != nil {
		t.Fatalf("Prepare SELECT failed: %v", err)
	}
	defer stmt.Finalize()

	// Currently, Step() for SELECT * returns false/done as full B-Tree scan
	// bytecode generation for SELECT is still in progress.
	_, err = stmt.Step()
	if err != nil {
		t.Fatalf("Step SELECT failed: %v", err)
	}

	fmt.Println("Showcase finished successfully!")
	fmt.Println("--------------------------------")
}
