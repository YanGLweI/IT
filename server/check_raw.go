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

	fmt.Println("=== user_permissions 表结构 ===")
	rows, _ := db.Query("DESCRIBE user_permissions")
	for rows.Next() {
		var field, typ, null, key, defaultVal, extra string
		rows.Scan(&field, &typ, &null, &key, &defaultVal, &extra)
		fmt.Printf("%-20s %-20s %-4s %-4s %-10s %s\n", field, typ, null, key, defaultVal, extra)
	}
	rows.Close()

	fmt.Println("\n=== user_permissions 表数据 ===")
	rows, _ = db.Query("SELECT * FROM user_permissions")
	columns, _ := rows.Columns()
	fmt.Printf("列名: %v\n\n", columns)
	
	count := 0
	for rows.Next() {
		count++
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		
		fmt.Printf("行[%d]:\n", count)
		for i, col := range columns {
			fmt.Printf("  %s = %v (type: %T)\n", col, values[i], values[i])
		}
		fmt.Println()
	}
	rows.Close()
	
	if count == 0 {
		fmt.Println("表中没有数据")
	}
}
