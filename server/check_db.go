package main

import (
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

type Department struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
}

func main() {
	dsn := "it:a*999999@tcp(10.60.254.127:3306)/it_platform?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	fmt.Println("=== 部门表 ===")
	var depts []Department
	db.Find(&depts)
	for _, d := range depts {
		fmt.Printf("ID=%d, Name=%s\n", d.ID, d.Name)
	}

	fmt.Println("\n=== 用户权限表 ===")
	var users []UserPermission
	db.Find(&users)
	for _, u := range users {
		fmt.Printf("ID=%v (type: %T), Name=%s, DepartmentID=%d, PositionName=%s\n", 
			u.ID, u.ID, u.Name, u.DepartmentID, u.PositionName)
	}
}
