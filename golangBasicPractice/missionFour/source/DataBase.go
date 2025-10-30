package source

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"column:username;unique;not null"`
	Email    string `gorm:"column:email;unique;not null"`
	Password string `gorm:"column:password;not null"`
}

type Post struct {
	gorm.Model
	Title   string `gorm:"column:title;not null"`
	Content string `gorm:"column:content;not null"`
	UserID  uint   `gorm:"column:user_id;not null"`
	User    User   `gorm:"foreignKey:UserID"`
}

type Comment struct {
	gorm.Model
	PostID  uint   `gorm:"column:post_id;not null"`
	Post    Post   `gorm:"foreignKey:PostID"`
	Content string `gorm:"column:content;not null"`
	UserID  uint   `gorm:"column:user_id;not null"`
	User    User   `gorm:"foreignKey:UserID"`
}

var db *gorm.DB

func InitDB() {
	// Use environment variables for database configuration
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "four")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "root1234")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	fmt.Printf("Connecting to database with DSN: %s:***@tcp(%s:%s)/%s\n", dbUser, dbHost, dbPort, dbName)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		panic("failed to connect database")
	}

	fmt.Println("Successfully connected to database!")

	// Auto migrate tables
	err = db.AutoMigrate(&User{}, &Post{}, &Comment{})
	if err != nil {
		fmt.Printf("Failed to migrate database: %v\n", err)
		panic("failed to migrate database")
	}

	fmt.Println("Database migration completed!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetDB() *gorm.DB {
	return db
}

func CreateUser(user *User) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.Create(user).Error
}

func GetUserByUsername(username string, user *User) error {
	return db.Where("username = ?", username).First(user).Error
}

func CreatePost(post *Post) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.Create(post).Error
}

func CreateComment(comment *Comment) error {
	return db.Create(comment).Error
}

func GetPostByUserID(userID uint, posts *[]Post) error {
	return db.Where("user_id = ?", userID).Find(posts).Error
}

func GetPostByID(postID string, post *Post) error {
	return db.Where("id = ?", postID).First(post).Error
}

func UpdatePost(post *Post) error {
	// Only update the specified fields, preserve created_at and other fields
	return db.Model(&Post{}).Where("id = ?", post.ID).Updates(map[string]interface{}{
		"title":   post.Title,
		"content": post.Content,
	}).Error
}

func DeletePost(post *Post) error {
	return db.Delete(post).Error
}

func InsertComment(comment *Comment) error {
	return db.Create(comment).Error
}

func GetCommentsByPostID(postID string, comments *[]Comment) error {
	return db.Where("post_id = ?", postID).Find(comments).Error
}
