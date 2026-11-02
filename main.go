package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func handleError(c *gin.Context, err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
	c.HTML(http.StatusInternalServerError, "base.html", gin.H{
		"Title": "500",
		"Description": "Внутренняя ошибка сервера",
		"Content": "",
	})
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
	router.GET("/static/*filepath", func(c *gin.Context) {
		c.File(filepath.Join("./static", c.Param("filepath")))
	})
	router.GET("/favicon.png", func(c *gin.Context) {
	    c.Redirect(http.StatusFound, "/static/favicon.png")
	})

	md_with_description := regexp.MustCompile(`(?m)^# (.+)\n+(.+)\n+([\s\S]+)$`)
	md_without_description := regexp.MustCompile(`(?m)^# (.+)\n+([\s\S]*)$`)
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/" {
			path = "index"
		}
		path = strings.TrimSuffix(path, ".html")
		path = filepath.Join("./content/", path + ".md")

		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				c.HTML(http.StatusNotFound, "base.html", gin.H{
					"Title": "404",
					"Description": "Такой страницы не существует",
					"Content": "",
				})
			} else {
				handleError(c, err)
			}
			return
		}

		var args gin.H
		md_string := string(data)
		matches := md_with_description.FindStringSubmatch(md_string)
		if len(matches) < 4 {
			matches := md_without_description.FindStringSubmatch(md_string)
			if len(matches) < 3 {
				args = gin.H{
					"Title": "",
					"Description": "",
					"Content": mdToHTML(md_string),
				}
			} else {
				args = gin.H{
					"Title": matches[1],
					"Description": "",
					"Content": mdToHTML(matches[2]),
				}
			}
		} else {
			args = gin.H{
				"Title": matches[1],
				"Description": mdToHTML(matches[2]),
				"Content": mdToHTML(matches[3]),
			}
		}

		c.HTML(http.StatusOK, "base.html", args)
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
