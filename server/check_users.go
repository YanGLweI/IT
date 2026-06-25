package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "it:a*999999@tcp(10.60.254.127:3306)/it_platform")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("=== 用户表 ===")
	rows, _ := db.Query("SELECT * FROM users LIMIT 5")
	defer rows.Close()
	
	for rows.Next() {
		var id int
		var username, password string
		rows.Scan(&id, &username, &password)
		fmt.Printf("ID=%d, Username=%s\n", id, username)
	}
}
