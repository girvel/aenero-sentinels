package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func setup() error {
	router := gin.Default()
	router.Use(gzip.Gzip(gzip.BestCompression))
	router.LoadHTMLGlob("html/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	router.GET("/static/*filepath", func(c *gin.Context) {
		c.File(filepath.Join("./static", c.Param("filepath")))
	})

	port := os.Getenv("PORT")
	if port == "" {
	  port = "8080"
	}

	return router.Run(":" + port)
}

func main() {
	err := setup()
	if err != nil {
		log.Fatalf("Setup error: %s", err.Error())
	}
}
