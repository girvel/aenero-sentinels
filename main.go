package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func handleError(c *gin.Context) {
	c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte("500"))
}

func mdToHTML(md string) template.HTML {
	p := parser.NewWithExtensions(
		parser.CommonExtensions |
		parser.AutoHeadingIDs |
		parser.NoEmptyLineBeforeBlock,
	)
	doc := p.Parse([]byte(md))
	renderer := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags})
	return template.HTML(markdown.Render(doc, renderer))
}

func setup() error {
	router := gin.Default()
	router.Use(gzip.Gzip(gzip.BestCompression))
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		data, err := os.ReadFile("./content/index.md")
		if err != nil {
			handleError(c)
			return
		}

		pattern := regexp.MustCompile(`(?m)^# (.+)\n\n(.+)\n\n([\s\S]+)$`)
		matches := pattern.FindStringSubmatch(string(data))
		if len(matches) < 4 {
			handleError(c)
			return
		}

		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title": matches[1],
			"Description": mdToHTML(matches[2]),
			"Content": mdToHTML(matches[3]),
		})
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
