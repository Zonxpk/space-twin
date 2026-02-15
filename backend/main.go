// @title FloorPlan Whiteboard API
// @version 1.0
// @description Backend API for floorplan processing and detection
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http

package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	_ "floorplan-whiteboard/docs" // Load generated swagger docs
	"floorplan-whiteboard/handler"
	"floorplan-whiteboard/realtime"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//go:embed docs/swagger.json docs/swagger.yaml
var swaggerFS embed.FS

func main() {
	r := gin.Default()

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/swagger.yaml", func(c *gin.Context) {
		c.Header("Content-Type", "application/yaml")
		data, _ := fs.ReadFile(swaggerFS, "docs/swagger.yaml")
		c.Data(http.StatusOK, "application/yaml", data)
	})
	r.GET("/swagger.json", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		data, _ := fs.ReadFile(swaggerFS, "docs/swagger.json")
		c.Data(http.StatusOK, "application/json", data)
	})

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
		api.POST("/upload", handler.UploadFloorplan)
		api.POST("/process/edges", handler.ProcessFloorplanEdges)
		api.POST("/process/edges-json", handler.ProcessFloorplanWithJSON)
	}

	r.GET("/ws", func(c *gin.Context) {
		hub.HandleWebSocket(c)
	})

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
