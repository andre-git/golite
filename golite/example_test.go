package golite_test

import (
	"fmt"
	"golite"
	"log"
)

func Example() {
	// 1. Open a database connection
	db, err := golite.Open("test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 2. Execute a simple SQL statement
	err = db.Exec("BEGIN; SELECT ALL; COMMIT;")
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
	} else {
		fmt.Println("SQL executed successfully!")
	}

	// Output: SQL executed successfully!
}
