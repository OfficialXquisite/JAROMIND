// main.go - CLEANED UP VERSION
package main

import (
    "os"
    "github.com/gin-gonic/gin"
    "github.com/AbaraEmmanuel/jaromind-backend/database"
    "github.com/AbaraEmmanuel/jaromind-backend/router"
)

func main() {
    // Initialize MongoDB
    database.InitDatabase()

    // Create router
    r := gin.Default()

    // Register routes (CORS is already in router.RegisterRoutes)
    router.RegisterRoutes(r)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    r.Run("0.0.0.0:" + port)
}