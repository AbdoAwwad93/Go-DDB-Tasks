package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var serverDSN = "root:Awwad@411@tcp(127.0.0.1:3306)/"
var db *sql.DB

func connectServer() {

	var err error
	db, err = sql.Open("mysql", serverDSN)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to MySQL server")

}
func createDatabase(Dbname string) {
	conn, err := sql.Open("mysql", serverDSN)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", Dbname)

	_, err = conn.Exec(query)
	if err != nil {
		panic(err)
	}

	fmt.Println("Database created")
}
func connectDatabase(Dbname string) {
	var dbDSN = fmt.Sprintf("root:Awwad@411@tcp(127.0.0.1:3306)/%s", Dbname)
	var err error
	db, err = sql.Open("mysql", dbDSN)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to: ", Dbname)
}
func createTable(tableName string, columns map[string]string) {

	var cols []string
	for colName, colType := range columns {
		cols = append(cols, fmt.Sprintf("%s %s", colName, colType))
	}
	query := fmt.Sprintf("Create table if not exists %s (%s)", tableName, strings.Join(cols, ","))

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}

	fmt.Println("Table created")

}
func insert(Tablename string, data map[string]interface{}) {

	var cols []string
	var placholders []string
	var values []interface{}

	for col, val := range data {
		cols = append(cols, col)
		placholders = append(placholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf("insert into %s (%s) values(%s)", Tablename,
		strings.Join(cols, ","),
		strings.Join(placholders, ","),
	)

	_, err := db.Exec(query, values...)
	if err != nil {
		panic(err)
	}

	fmt.Println("Inserted into table:", Tablename)

}

func update(tablename string, data map[string]interface{}, condition string) {
	var set []string
	var values []interface{}
	for col, val := range data {
		set = append(set, fmt.Sprintf("%s = ?", col))
		values = append(values, val)
	}

	query := fmt.Sprintf("Update %s set %s where %s", tablename,
		strings.Join(set, ","),
		condition,
	)
	_, err := db.Exec(query, values...)
	if err != nil {
		panic(err)
	}

	fmt.Println("Updated tabel : ",tablename)

}
func Read(tablename string) {

	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", tablename))
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	values := make([]interface{}, len(cols))
	valuesptrs := make([]interface{}, len(cols))

	for i := range values {
		valuesptrs[i] = &values[i]
	}
	for rows.Next() {
		err := rows.Scan(valuesptrs...)
		if err != nil {
			panic(err)
		}

		for i, col := range cols {
			fmt.Printf("%s: %v | ", col, values[i])
		}
		fmt.Println()
	}

}

func delete(tablename string, condition string) {

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", tablename, condition)

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}

	fmt.Println("Deleted from", tablename)

}
func main() {
	var dbName string

	fmt.Println("=== GO DB System ===")
	fmt.Print("Enter Database Name: ")
	fmt.Scan(&dbName)

	connectServer()
	createDatabase(dbName)
	connectDatabase(dbName)

	for {
		fmt.Println("\nChoose Operation:")
		fmt.Println("1) Create Table")
		fmt.Println("2) Insert")
		fmt.Println("3) Read")
		fmt.Println("4) Update")
		fmt.Println("5) Delete")
		fmt.Println("6) Exit")

		var choice int
		fmt.Print("Enter choice: ")
		fmt.Scan(&choice)

		switch choice {

		case 1:
			var tableName string
			fmt.Print("Enter table name: ")
			fmt.Scan(&tableName)

			var n int
			fmt.Print("Number of columns: ")
			fmt.Scan(&n)

			columns := make(map[string]string)

			for i := 0; i < n; i++ {
				var name, typ string
				fmt.Printf("Column %d name: ", i+1)
				fmt.Scan(&name)

				fmt.Printf("Column %d type: ", i+1)
				fmt.Scan(&typ)

				columns[name] = typ
			}

			createTable(tableName, columns)

		case 2:
			var tableName string
			fmt.Print("Enter table name: ")
			fmt.Scan(&tableName)

			var n int
			fmt.Print("Number of fields: ")
			fmt.Scan(&n)

			data := make(map[string]interface{})

			for i := 0; i < n; i++ {
				var key, val string
				fmt.Print("Column name: ")
				fmt.Scan(&key)

				fmt.Print("Value: ")
				fmt.Scan(&val)

				data[key] = val
			}

			insert(tableName, data)

		case 3:
			var tableName string
			fmt.Print("Enter table name: ")
			fmt.Scan(&tableName)

			Read(tableName)

		case 4:
			var tableName string
			fmt.Print("Enter table name: ")
			fmt.Scan(&tableName)

			var n int
			fmt.Print("Number of fields to update: ")
			fmt.Scan(&n)

			data := make(map[string]interface{})

			for i := 0; i < n; i++ {
				var key, val string
				fmt.Print("Column: ")
				fmt.Scan(&key)

				fmt.Print("New Value: ")
				fmt.Scan(&val)

				data[key] = val
			}

			var condition string
			fmt.Print("Condition (e.g. id=1): ")
			fmt.Scanln(&condition)

			update(tableName, data, condition)

		case 5:
			var tableName, condition string

			fmt.Print("Enter table name: ")
			fmt.Scan(&tableName)

			fmt.Print("Condition: ")
			fmt.Scanln(&condition)

			delete(tableName, condition)

		case 6:
			fmt.Println("Exiting...")
			return

		default:
			fmt.Println("Invalid choice")
		}
	}
}