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

	fmt.Println("=== permission_rules 表数据 ===")
	rows, _ := db.Query("SELECT id, position_name, rules_json FROM permission_rules LIMIT 3")
	defer rows.Close()
	
	for rows.Next() {
		var id int
		var positionName string
		var rulesJSON string
		rows.Scan(&id, &positionName, &rulesJSON)
		
		fmt.Printf("\nID=%d, PositionName=%s\n", id, positionName)
		fmt.Printf("RulesJSON:\n%s\n", rulesJSON)
		fmt.Println("---")
	}
}
