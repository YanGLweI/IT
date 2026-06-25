package main

import (
	"encoding/json"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type UserPermission struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	Name            string `json:"name"`
	DepartmentID    uint   `json:"department_id"`
	PositionName    string `json:"position_name"`
	SystemRolesJSON string `json:"system_roles_json"`
}

func main() {
	dsn := "it:a*999999@tcp(10.60.254.127:3306)/it_platform?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	var users []UserPermission
	result := db.Find(&users)
	
	fmt.Printf("查询结果: %+v\n", result)
	fmt.Printf("用户数量: %d\n", len(users))
	
	for i, u := range users {
		fmt.Printf("\n用户[%d]:\n", i)
		fmt.Printf("  ID=%v (type: %T, value: %d)\n", u.ID, u.ID, u.ID)
		fmt.Printf("  Name=%s\n", u.Name)
		fmt.Printf("  DepartmentID=%d\n", u.DepartmentID)
		
		jsonData, _ := json.Marshal(u)
		fmt.Printf("  JSON: %s\n", string(jsonData))
	}
}
