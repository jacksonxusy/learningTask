package router

import (
	"golangBasicPractice/missionFour/source"
	"net/http"

	"github.com/gin-gonic/gin"
)

//实现评论的创建功能，已认证的用户可以对文章发表评论。
//实现评论的读取功能，支持获取某篇文章的所有评论列表。

func InitCommentRoutes(r *gin.Engine) {
	comment := r.Group("/api/comment")
	comment.Use(AuthMiddleware()) // 对所有comment路由应用JWT认证
	{
		comment.POST("/insert", InsertComment)
		comment.GET("/get", GetComment)
	}
}

func InsertComment(c *gin.Context) {
	// 从上下文获取用户ID
	userID := c.MustGet("userID").(uint)
	var comment source.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	comment.UserID = userID
	// 插入评论到数据库
	if err := source.InsertComment(&comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert comment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "comment inserted successfully"})
}

func GetComment(c *gin.Context) {
	// 从查询参数中获取文章ID
	postID := c.Query("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "postID is required"})
		return
	}
	var comments []source.Comment
	// 从数据库查询评论
	if err := source.GetCommentsByPostID(postID, &comments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get comments"})
		return
	}
	c.JSON(http.StatusOK, comments)
}
