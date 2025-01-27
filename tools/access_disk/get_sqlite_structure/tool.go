package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Opens main sqlite database and returns JSON with list of tables(name, description) and columns(name, type).
type get_sqlite_structure struct {
}

type ColumnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
type TableSchema struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Columns     []ColumnInfo `json:"columns"`
}
type DatabaseSchema struct {
	Tables []TableSchema `json:"tables"`
}

func (st *get_sqlite_structure) run() (schema DatabaseSchema) {
	// Open the SQLite database
	db, err := sql.Open("sqlite3", "main.sqlite")
	if err != nil {
		log.Fatal(fmt.Errorf("error opening database: %v", err))
	}
	defer db.Close()

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

	// Query to get table names
	tableQuery := `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`
	tableRows, err := db.Query(tableQuery)
	if err != nil {
		log.Fatal(fmt.Errorf("error querying tables: %v", err))
	}
	defer tableRows.Close()

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

		var tableSchema TableSchema
		tableSchema.Name = tableName
		tableSchema.Description = descriptions[tableName]

		// Iterate through columns
		for columnRows.Next() {
			var colName, colType string
			var colIndex, nullable int
			var defaultVal interface{}

			if err := columnRows.Scan(&colIndex, &colName, &colType, &nullable, &defaultVal, &colIndex); err != nil {
				log.Fatal(fmt.Errorf("error scanning column info: %v", err))
			}

			tableSchema.Columns = append(tableSchema.Columns, ColumnInfo{
				Name: colName,
				Type: colType,
			})
		}

		schema.Tables = append(schema.Tables, tableSchema)
	}

	return
}
