package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/okakafavour/jaromind-backend/controllers"
	"github.com/okakafavour/jaromind-backend/middleware"
)

func RegisterRoutes(router *gin.Engine) {

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // change * to frontend URL in production
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Public routes
	router.POST("/register", controllers.RegisterUser)
	router.POST("/login", controllers.LoginUser)

	// Protected routes
	protected := router.Group("/user")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		protected.GET("/profile", controllers.GetProfile)
	}
}
