package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

//假设你要开发一个博客系统，有以下几个实体： User （用户）、 Post （文章）、 Comment （评论）。
//要求 ：
//使用Gorm定义 User 、 Post 和 Comment 模型，其中 User 与 Post 是一对多关系（一个用户可以发布多篇文章）， Post 与 Comment 也是一对多关系（一篇文章可以有多个评论）。
//编写Go代码，使用Gorm创建这些模型对应的数据库表。

type User struct {
	ID        int    `gorm:"primaryKey"`
	Name      string `gorm:"column:name"`
	Email     string `gorm:"column:email"`
	Password  string `gorm:"column:password"`
	PostCount int    `gorm:"column:post_count;default:0"` // Add this field
}

type Post struct {
	ID            int    `gorm:"primaryKey"`
	Title         string `gorm:"column:title"`
	Content       string `gorm:"column:content"`
	AuthorID      int    `gorm:"column:author_id"`
	Author        User   `gorm:"foreignKey:AuthorID;references:ID"`
	Comments      []Comment
	CommentCount  int    `gorm:"column:comment_count;default:0"`      // Add this field
	CommentStatus string `gorm:"column:comment_status;default:'有评论'"` // Add this field
}

type Comment struct {
	ID      int    `gorm:"primaryKey"`
	Content string `gorm:"column:content"`
	PostID  int    `gorm:"column:post_id"`
	Post    Post   `gorm:"foreignKey:PostID;references:ID"`
}

// 为 Post 模型添加一个钩子函数，在文章创建时自动更新用户的文章数量统计字段。
func (p *Post) BeforeCreate(tx *gorm.DB) (err error) {
	tx.Model(&User{}).Where("id = ?", p.AuthorID).Update("post_count", gorm.Expr("post_count + ?", 1))
	return
}

func (c *Comment) BeforeDelete(tx *gorm.DB) (err error) {
	// 在删除评论前，更新对应文章的评论数量统计字段,如果评论数量为 0，则更新文章的评论状态为 "无评论"
	tx.Model(&Post{}).Where("id = ?", c.PostID).Update("comment_status", gorm.Expr("case when comment_count = 0 then '无评论' end"))
	return
}

func main() {
	db, err := gorm.Open(mysql.Open("root:root1234@tcp(127.0.0.1:3306)/gorm?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&User{}, &Post{}, &Comment{})

	var comment3 = Comment{
		ID:      3,
		PostID:  2,
		Content: "这是第二篇文章的第一个评论",
	}
	db.Delete(&comment3)

	//var post3 = Post{
	//	Title:    "第三篇文章",
	//	Content:  "这是第三篇文章的内容",
	//	AuthorID: 1,
	//}
	//source.Create(&post3)

	//var user = User{
	//	Name:     "张三",
	//	Email:    "zhangsan@example.com",
	//	Password: "123456",
	//}
	//var post = Post{
	//	Title:    "第一篇文章",
	//	Content:  "这是第一篇文章的内容",
	//	AuthorID: 1,
	//}
	//var post2 = Post{
	//	Title:    "第二篇文章",
	//	Content:  "这是第二篇文章的内容",
	//	AuthorID: 1,
	//}
	//var comment = Comment{
	//	ID:      1,
	//	PostID:  1,
	//	Content: "这是第一篇文章的第一个评论",
	//}
	//var comment2 = Comment{
	//	ID:      2,
	//	PostID:  1,
	//	Content: "这是第一篇文章的第二个评论",
	//}

	//source.Create(&user)
	//var posts = []Post{post, post2}
	//var comments = []Comment{comment, comment2, comment3}
	//source.Create(&posts)
	//source.Create(&comments)
	// 2.1 编写Go代码，使用Gorm查询某个用户发布的所有文章及其对应的评论信息。
	var posts []Post
	db.Where("author_id = ?", 1).Preload("Comments").Find(&posts)
	//for _, post := range posts {
	//	fmt.Println(post.Title)
	//	for _, comment := range post.Comments {
	//		fmt.Println(comment.Content)
	//	}
	//}
	// select post_id, count(*) from comments group by post_id order by count(*) limit 1

	// 编写Go代码，使用Gorm查询评论数量最多的文章信息。
	var postWithMaxComment Post
	db.Model(&Post{}).Where("author_id = ?", 1).Order("(select count(*) from comments where comments.post_id = posts.id) desc").First(&postWithMaxComment)
	fmt.Println(postWithMaxComment.Title)

	// 3.1 为 Post 模型添加一个钩子函数，在文章创建时自动更新用户的文章数量统计字段。

}
