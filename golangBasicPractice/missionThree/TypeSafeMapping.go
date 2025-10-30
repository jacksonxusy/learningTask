package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

//假设有一个 books 表，包含字段 id 、 title 、 author 、 price 。
//要求 ：
//定义一个 Book 结构体，包含与 books 表对应的字段。
//编写Go代码，使用Sqlx执行一个复杂的查询，例如查询价格大于 50 元的书籍，并将结果映射到 Book 结构体切片中，确保类型安全。

type Book struct {
	ID     int    `source:"id"`
	Title  string `source:"title"`
	Author string `source:"author"`
	Price  int    `source:"price"`
}

func main() {
	db, err := gorm.Open(mysql.Open("root:root1234@tcp(127.0.0.1:3306)/gorm?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Book{})
	db.Create(&Book{
		ID:     1,
		Title:  "Go语言程序设计",
		Author: "张三",
		Price:  60,
	})
	db.Create(&Book{
		ID:     2,
		Title:  "Go语言程序设计2",
		Author: "李四",
		Price:  70,
	})
	var books []Book
	db.Where("price > ?", 50).Find(&books)
	fmt.Println(books)
}
