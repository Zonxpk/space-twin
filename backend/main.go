package main

import (
	"log"
	"net/http"

	"floorplan-whiteboard/handler"
	"floorplan-whiteboard/realtime"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	hub := realtime.NewHub()
	go hub.Run()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "SmartFloor Backend Running"})
	})

	api := r.Group("/api/v1")
	{
		api.GET("/openapi.yaml", handler.GetOpenAPISpec)
		api.GET("/docs", handler.GetSwaggerUI)
		api.POST("/upload", handler.UploadFloorplan)
		api.POST("/debug/crop", handler.DebugCrop)
	}

	r.GET("/ws", func(c *gin.Context) {
		hub.HandleWebSocket(c)
	})

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
