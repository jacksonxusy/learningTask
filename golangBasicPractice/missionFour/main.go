package main

import (
	"golangBasicPractice/missionFour/router"
	"golangBasicPractice/missionFour/source"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	source.InitDB()

	// Create gin router
	r := gin.Default()

	// Initialize routes
	router.InitAuthRoutes(r)
	router.InitPostRoutes(r)
	router.InitCommentRoutes(r)

	// Start server
	r.Run(":8080")
}
