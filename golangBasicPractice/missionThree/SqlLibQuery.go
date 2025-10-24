package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

//假设你已经使用Sqlx连接到一个数据库，并且有一个 employees 表，包含字段 id 、 name 、 department 、 salary 。
//要求 ：
//编写Go代码，使用Sqlx查询 employees 表中所有部门为 "技术部" 的员工信息，并将结果映射到一个自定义的 Employee 结构体切片中。
//编写Go代码，使用Sqlx查询 employees 表中工资最高的员工信息，并将结果映射到一个 Employee 结构体中。

type Employee struct {
	ID         int    `db:"id"`
	Name       string `db:"name"`
	Department string `db:"department"`
	Salary     int    `db:"salary"`
}

func main() {
	db, err := gorm.Open(mysql.Open("root:root1234@tcp(127.0.0.1:3306)/gorm?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	//db.AutoMigrate(&Employee{})
	//db.Create(&Employee{
	//	Name:       "张三",
	//	Department: "技术部",
	//	Salary:     5000,
	//})
	//db.Create(&Employee{
	//	Name:       "李四",
	//	Department: "技术部",
	//	Salary:     6000,
	//})
	//var res = []Employee{}
	//db.Where("department = ?", "技术部").Find(&res)
	//fmt.Println(res)

	var maxSalaryEmployee Employee
	db.Where("department = ?", "技术部").Order("salary desc").First(&maxSalaryEmployee)
	fmt.Println(maxSalaryEmployee)

}
