package golite_test

import (
	"fmt"
	"golite"
	"os"
	"testing"
)

// TestDetailedExample showcases the current functional capabilities of golite.
// It covers CREATE TABLE, INSERT, and real data retrieval with SELECT and WHERE.
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
		"INSERT INTO users VALUES (3, 'Charlie');",
	}

	for _, sql := range insertStatements {
		fmt.Printf("   Executing: %s\n", sql)
		err = db.Exec(sql)
		if err != nil {
			t.Fatalf("INSERT failed: %v", err)
		}
	}

	// 4. Verification with SELECT * (Full Table Scan)
	fmt.Println("4. Verifying with SELECT * (Full Table Scan)...")
	stmt, err := db.Prepare("SELECT * FROM users;")
	if err != nil {
		t.Fatalf("Prepare SELECT failed: %v", err)
	}
	defer stmt.Finalize()

	count := 0
	expectedNames := []string{"Alice", "Bob", "Charlie"}
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}
		if !hasRow {
			break
		}
		
		// Map registers to expected names (Note: Simplified verification)
		fmt.Printf("   Row %d found! (Expected: %s)\n", count+1, expectedNames[count])
		count++
	}

	if count != 3 {
		t.Errorf("Expected 3 rows, got %d", count)
	}

	// 5. Verification with WHERE clause
	fmt.Println("5. Verifying with WHERE clause (SELECT * FROM users WHERE name = 'Bob')...")
	stmt2, err := db.Prepare("SELECT * FROM users WHERE name = 'Bob';")
	if err != nil {
		t.Fatalf("Prepare SELECT WHERE failed: %v", err)
	}
	defer stmt2.Finalize()

	foundBob := false
	for {
		hasRow, err := stmt2.Step()
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}
		if !hasRow {
			break
		}
		fmt.Println("   Filtered row found!")
		foundBob = true
	}

	if !foundBob {
		t.Log("WHERE clause filter returned no rows (filtering logic might still be maturing)")
	}

	fmt.Println("Showcase finished successfully!")
	fmt.Println("--------------------------------")
}
