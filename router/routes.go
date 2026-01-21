package router

import (
	"time"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/AbaraEmmanuel/jaromind-backend/controllers"
	"github.com/AbaraEmmanuel/jaromind-backend/middleware"
)

func RegisterRoutes(router *gin.Engine) {
    // CORS configuration - ADD YOUR LIVE WEB APP URL
    router.Use(cors.New(cors.Config{
        AllowOrigins: []string{
            "http://localhost:8003",          // Your admin panel
            "http://127.0.0.1:8000",         // Alternative local
            "http://localhost:3000",         // Common dev port
            "https://edu-tech-v1-mu.vercel.app", // Your live web app (REMOVE trailing /)
            "https://jaromind.com",   // Add if you have this
            "http://localhost:5500",         // VS Code Live Server
            "http://127.0.0.1:5500",         // Alternative
        },
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
        ExposeHeaders:    []string{"Content-Length", "Authorization"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

	// Public routes
	router.POST("/register", controllers.RegisterUser)
	router.POST("/login", controllers.LoginUser)
	
	// Public course routes
	router.GET("/courses", controllers.GetAllCourses)
	router.GET("/courses/:id", controllers.GetCourseByID)
	router.GET("/courses/:id/stats", controllers.GetCourseStats)

	// Protected user routes
	protected := router.Group("/user")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		protected.GET("/profile", controllers.GetProfile)
		protected.POST("/enroll/:courseId", controllers.EnrollInCourse)
		protected.GET("/enrollments", controllers.GetUserEnrollments)
		protected.PUT("/courses/:courseId/progress", controllers.UpdateProgress)
		protected.POST("/courses/:courseId/review", controllers.AddReview)
	}

	// Admin routes (add admin middleware later)
	admin := router.Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware()) // Add admin check middleware
	{
		admin.POST("/courses", controllers.CreateCourse)
		admin.PUT("/courses/:id", controllers.UpdateCourse)
		admin.DELETE("/courses/:id", controllers.DeleteCourse)
	}
}