package router

import (
	"golangBasicPractice/missionFour/internal/config"
	"golangBasicPractice/missionFour/internal/errors"
	"golangBasicPractice/missionFour/source"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

//实现用户注册和登录功能，用户注册时需要对密码进行加密存储，登录时验证用户输入的用户名和密码。
//使用 JWT（JSON Web Token）实现用户认证和授权，用户登录成功后返回一个 JWT，后续的需要认证的接口需要验证该 JWT 的有效性。

// InitAuthRoutes 初始化认证相关的路由
func InitAuthRoutes(r *gin.Engine) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", Register)
		auth.POST("/login", Login)
	}
}

func Register(c *gin.Context) {
	var user source.User
	if err := c.ShouldBindJSON(&user); err != nil {
		handleError(c, errors.NewAppError(http.StatusBadRequest, "Invalid request format", err))
		return
	}

	// Validate input
	if user.Username == "" || user.Email == "" || user.Password == "" {
		handleError(c, errors.NewAppError(http.StatusBadRequest, "Username, email, and password are required", nil))
		return
	}

	// Check if user already exists
	var existingUser source.User
	if err := source.GetUserByUsername(user.Username, &existingUser); err == nil {
		handleError(c, errors.ErrUserExists)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		handleError(c, errors.NewAppError(http.StatusInternalServerError, "Failed to hash password", err))
		return
	}
	user.Password = string(hashedPassword)

	// Create user
	if err := source.CreateUser(&user); err != nil {
		handleError(c, errors.NewAppError(http.StatusInternalServerError, "Failed to create user", err))
		return
	}

	// Clear password before returning
	user.Password = ""
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

func Login(c *gin.Context) {
	var loginReq source.User
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		handleError(c, errors.NewAppError(http.StatusBadRequest, "Invalid request format", err))
		return
	}

	// Validate input
	if loginReq.Username == "" || loginReq.Password == "" {
		handleError(c, errors.NewAppError(http.StatusBadRequest, "Username and password are required", nil))
		return
	}

	// Get user from database
	var user source.User
	if err := source.GetUserByUsername(loginReq.Username, &user); err != nil {
		handleError(c, errors.ErrInvalidCredentials)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		handleError(c, errors.ErrInvalidPassword)
		return
	}

	// Generate JWT
	jwtSecret := getJWTSecret()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iat":      time.Now().Unix(),
	})

	// Sign JWT
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		handleError(c, errors.NewAppError(http.StatusInternalServerError, "Failed to sign token", err))
		return
	}

	// Clear password before returning
	user.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user,
	})
}

// getJWTSecret loads JWT secret from YAML config
func getJWTSecret() string {
	cfg, err := config.LoadSimple("configs/config.yaml")
	if err != nil {
		return "jackson" // fallback default
	}
	return cfg.JWT.Secret
}
