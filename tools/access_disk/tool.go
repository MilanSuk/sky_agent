package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Search for or change user's and device data.
type access_disk struct {
	Description string //Describe action(read, update, insert, delete). Place or hints where the data could be stored. If writing, the value of data.
}

func (st *access_disk) run() string {
	// Open the SQLite database
	db, err := sql.Open("sqlite3", "main.sqlite")
	if err != nil {
		log.Fatal(fmt.Errorf("error opening database: %v", err))
	}
	defer db.Close()

	SystemPrompt := "You are an AI programming assistant, who enjoys precision and carefully follows the user's requirements. You write SQL queries. You use tools all the time."

	UserPrompt := ""
	UserPrompt += "This is the structure(tables, columns) of database:\n"
	UserPrompt += getStructure(db)
	UserPrompt += "\n\n"

	UserPrompt += "This is the prompt from user:\n"
	UserPrompt += st.Description
	UserPrompt += "\n\n"
	UserPrompt += "Write an SQLite query based on user's prompt and run that query with tool 'execute_sqlite_query'.\n"
	UserPrompt += "Look at the data returned and If you are not happy with the result, you can try different queries. Specifically, be careful and precise with the WHERE clause. If no result is returned, remove the WHERE clause.\n"
	UserPrompt += "Create answer from the final query result. If value was read, you can describe place from which value(s) was selected."

	fmt.Println("UserPrompt:", UserPrompt)

	answer := SDK_RunAgent("agent", 20, 20000, SystemPrompt, UserPrompt)

	return answer
}

func getStructure(db *sql.DB) string {
	// Query to get table names
	tableQuery := `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`
	tableRows, err := db.Query(tableQuery)
	if err != nil {
		log.Fatal(fmt.Errorf("error querying tables: %v", err))
	}
	defer tableRows.Close()

	descriptions := make(map[string]string)
	{
		rows, err := db.Query("SELECT table_name, description FROM tables_descriptions")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var nm, desc string
				err := rows.Scan(&nm, &desc)
				if err == nil {
					descriptions[nm] = desc
				}
			}
		}

	}

	str := ""

	// Iterate through tables
	for tableRows.Next() {
		var tableName string
		if err := tableRows.Scan(&tableName); err != nil {
			log.Fatal(fmt.Errorf("error scanning table name: %v", err))
		}

		// Query columns for each table
		columnQuery := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
		columnRows, err := db.Query(columnQuery)
		if err != nil {
			log.Fatal(fmt.Errorf("error querying columns for table %s: %v", tableName, err))
		}
		defer columnRows.Close()

		str += tableName
		desc, found := descriptions[tableName]
		if found {
			str += "\t//" + desc
		}
		str += "\n"

		// Iterate through columns
		for columnRows.Next() {
			var colName, colType string
			var colIndex, nullable int
			var defaultVal interface{}

			if err := columnRows.Scan(&colIndex, &colName, &colType, &nullable, &defaultVal, &colIndex); err != nil {
				log.Fatal(fmt.Errorf("error scanning column info: %v", err))
			}

			str += fmt.Sprintf("\t- %s(%s)\n", colName, colType)
		}
	}

	return str
}
