package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var serverDSN = "root:Awwad@411@tcp(127.0.0.1:3306)/"
var dbDSN = "root:Awwad@411@tcp(127.0.0.1:3306)/Task"
func connectServer() {

	db, err := sql.Open("mysql", serverDSN)
	if err != nil {
		panic(err)
	}
	
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to MySQL server")

}
func createDatabase() {

	db, err := sql.Open("mysql", serverDSN)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	query := "CREATE DATABASE IF NOT EXISTS Task"

	_, err = db.Exec(query)
	if err != nil {
		panic(err)
	}

	fmt.Println("Database created")

}
func connectDatabase() *sql.DB {

	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to Task DB")

	return db
}
func createTable(db *sql.DB) {

	query := `CREATE TABLE IF NOT EXISTS employee (
		Id INT AUTO_INCREMENT PRIMARY KEY,
		Name VARCHAR(30),
		Age INT
	)`

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}

	fmt.Println("Table created")

}
func insertEmployee(db *sql.DB, name string, age int) {

	query := `INSERT INTO employee (Name, Age) VALUES (?, ?)`

	_, err := db.Exec(query, name, age)
	if err != nil {
		panic(err)
	}

	fmt.Println("Inserted")

}
func updateEmployee(db *sql.DB, name string, age int) {

	query := `UPDATE employee SET Age=? WHERE Name=?`

	_, err := db.Exec(query, age, name)
	if err != nil {
		panic(err)
	}

	fmt.Println("Updated")

}
func getEmployees(db *sql.DB) {

	rows, err := db.Query("SELECT * FROM employee")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	fmt.Println("Employees:")

	for rows.Next() {

		var id int
		var name string
		var age int

		err := rows.Scan(&id, &name, &age)
		if err != nil {
			panic(err)
		}

		fmt.Println(id, name, age)
	}

}
func deleteEmployee(db *sql.DB, name string) {

	query := `DELETE FROM employee WHERE Name=?`

	_, err := db.Exec(query, name)
	if err != nil {
		panic(err)
	}

	fmt.Println("Deleted")

}
func main() {

	// mydb:= *sql.DB

	connectServer()

	createDatabase()

	db := connectDatabase()
	defer db.Close()

	createTable(db)

	insertEmployee(db, "Awwad", 22)

	updateEmployee(db, "Awwad", 23)

	getEmployees(db)

	deleteEmployee(db, "Awwad")

}