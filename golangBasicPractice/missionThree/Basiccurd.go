package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Student struct {
	ID    int
	Name  string
	Age   int
	Grade string
}

func main() {
	var stu = Student{
		ID:    1,
		Name:  "Zhang San",
		Age:   20,
		Grade: "Three",
	}

	db, err := gorm.Open(mysql.Open("root:root1234@tcp(127.0.0.1:3306)/gorm?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&stu)
	db.Create(&stu)
	res := db.Find(&Student{}, "age > ?", 18)
	if res.Error != nil {
		panic("failed to find student")
	}
	db.Model(&stu).UpdateColumn("grade", "Four")
	db.Delete(&Student{}, "age < ?", 15)

}
