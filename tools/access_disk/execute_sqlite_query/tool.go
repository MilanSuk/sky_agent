package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Executes query in main sqlite database. Returns JSON with result.
type execute_sqlite_query struct {
	Query string //SQLite query to execute.
}

func (st *execute_sqlite_query) run() (results []map[string]interface{}) {
	// Open the SQLite database
	db, err := sql.Open("sqlite3", "data.sqlite")
	if err != nil {
		log.Fatal(fmt.Errorf("error opening database: %v", err))
	}
	defer db.Close()

	if strings.HasPrefix(strings.ToUpper(st.Query), "SELECT") {
		// Execute the query
		rows, err := db.Query(st.Query)
		if err != nil {
			log.Fatal(fmt.Errorf("error executing query: %v", err))
		}
		defer rows.Close()

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			log.Fatal(fmt.Errorf("error getting columns: %v", err))
		}

		// Scan rows into maps
		for rows.Next() {
			// Create a slice of interfaces to hold column values
			columnValues := make([]interface{}, len(columns))
			columnPointers := make([]interface{}, len(columns))

			for i := range columns {
				columnPointers[i] = &columnValues[i]
			}

			// Scan row values
			if err := rows.Scan(columnPointers...); err != nil {
				log.Fatal(fmt.Errorf("error scanning row: %v", err))
			}

			// Create a map for the current row
			rowMap := make(map[string]interface{})
			for i, colName := range columns {
				val := columnValues[i]
				// Handle potential nil values
				if val == nil {
					rowMap[colName] = nil
				} else {
					rowMap[colName] = val
				}
			}

			results = append(results, rowMap)
		}

		// Check for any errors encountered during iteration
		if err = rows.Err(); err != nil {
			log.Fatal(fmt.Errorf("error during row iteration: %v", err))
		}
	} else {
		ret, err := db.Exec(st.Query)
		if err != nil {
			log.Fatal(fmt.Errorf("error executing query: %v", err))
		}
		rows, _ := ret.RowsAffected()
		lid, _ := ret.LastInsertId()
		if rows > 0 {
			rowMap := make(map[string]interface{})
			rowMap["RowsAffected"] = rows
			rowMap["LastInsertId"] = lid
			results = append(results, rowMap)
		}
	}

	return
}
