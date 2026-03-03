package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/repocraft-project/server-go/internal/graceful"
)

func main() {
	log.SetOutput(gin.DefaultErrorWriter)

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World")
	})
	if err := graceful.Run(r, ":8080"); err != nil {
		log.Fatalln(err)
	}
	log.Println("Done")
}
