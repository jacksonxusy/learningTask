package router

import (
	"golangBasicPractice/missionFour/internal/config"
	"golangBasicPractice/missionFour/internal/errors"
	"golangBasicPractice/missionFour/source"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

//实现文章的创建功能，只有已认证的用户才能创建文章，创建文章时需要提供文章的标题和内容。
//实现文章的读取功能，支持获取所有文章列表和单个文章的详细信息。
//实现文章的更新功能，只有文章的作者才能更新自己的文章。
//实现文章的删除功能，只有文章的作者才能删除自己的文章。

// JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			handleError(c, errors.NewAppError(http.StatusUnauthorized, "Authorization header is required", nil))
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			handleError(c, errors.NewAppError(http.StatusUnauthorized, "Invalid authorization header format", nil))
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 解析JWT token
		jwtSecret := getJWTSecretFromConfig()
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			handleError(c, errors.ErrInvalidToken)
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 检查token是否过期
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					handleError(c, errors.ErrTokenExpired)
					c.Abort()
					return
				}
			}

			// 获取用户ID - 支持两种字段名
			var userID uint
			if uid, exists := claims["user_id"]; exists {
				userID = uint(uid.(float64))
			} else if uid, exists := claims["id"]; exists {
				userID = uint(uid.(float64))
			} else {
				handleError(c, errors.ErrInvalidToken)
				c.Abort()
				return
			}

			username, ok := claims["username"].(string)
			if !ok {
				handleError(c, errors.ErrInvalidToken)
				c.Abort()
				return
			}

			// 将用户信息存储到上下文中
			c.Set("userID", userID)
			c.Set("username", username)
			c.Next()
		} else {
			handleError(c, errors.ErrInvalidToken)
			c.Abort()
			return
		}
	}
}

func InitPostRoutes(r *gin.Engine) {
	post := r.Group("/api/post")
	post.Use(AuthMiddleware()) // 对所有post路由应用JWT认证
	{
		post.POST("/create", CreatePost)
		post.GET("/list", GetPostList)
		post.GET("/detail", GetPostDetail)
		post.PUT("/update", UpdatePost)
		post.DELETE("/delete", DeletePost)
	}
}

func CreatePost(r *gin.Context) {
	// 从 JWT 中获取用户 ID
	userID := r.MustGet("userID").(uint)
	var post source.Post
	if err := r.ShouldBindJSON(&post); err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if post.UserID != userID {
		r.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized to create post for this user"})
		return
	}
	// 保存文章到数据库
	if err := source.CreatePost(&post); err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create post"})
		return
	}
	r.JSON(http.StatusOK, gin.H{"message": "post created successfully"})
}

func GetPostList(r *gin.Context) {
	userID := r.MustGet("userID").(uint)
	var posts []source.Post
	// 从数据库查询用户的所有文章
	if err := source.GetPostByUserID(userID, &posts); err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get post list"})
		return
	}
	r.JSON(http.StatusOK, gin.H{"posts": posts})
}

func GetPostDetail(r *gin.Context) {
	postID := r.Query("postID")
	if postID == "" {
		r.JSON(http.StatusBadRequest, gin.H{"error": "postID is required"})
		return
	}
	var post source.Post
	if err := source.GetPostByID(postID, &post); err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get post detail"})
		return
	}
	r.JSON(http.StatusOK, gin.H{"post": post})
}

func UpdatePost(r *gin.Context) {
	userID := r.MustGet("userID").(uint)
	var post source.Post
	if err := r.ShouldBindJSON(&post); err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if post.UserID != userID {
		r.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized to update post for this user"})
		return
	}
	// 更新文章到数据库
	if err := source.UpdatePost(&post); err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update post"})
		return
	}
	r.JSON(http.StatusOK, gin.H{"message": "post updated successfully"})
}

func DeletePost(r *gin.Context) {
	userID := r.MustGet("userID").(uint)
	postID := r.Query("postID")
	if postID == "" {
		r.JSON(http.StatusBadRequest, gin.H{"error": "postID is required"})
		return
	}
	var post source.Post
	if err := source.GetPostByID(postID, &post); err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get post detail"})
		return
	}
	if post.UserID != userID {
		r.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized to delete post for this user"})
		return
	}

	// 删除文章从数据库
	if err := source.DeletePost(&post); err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete post"})
		return
	}
	r.JSON(http.StatusOK, gin.H{"message": "post deleted successfully"})
}

// getJWTSecretFromConfig loads JWT secret from YAML config for middleware
func getJWTSecretFromConfig() string {
	cfg, err := config.LoadSimple("configs/config.yaml")
	if err != nil {
		return "jackson" // fallback default
	}
	return cfg.JWT.Secret
}

// handleError handles application errors and returns appropriate JSON responses
func handleError(c *gin.Context, err error) {
	if appErr, ok := errors.IsAppError(err); ok {
		c.JSON(appErr.Code, gin.H{
			"error": appErr.Message,
		})
		return
	}

	// Handle unexpected errors
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "Internal server error",
	})
}
