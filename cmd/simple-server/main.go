package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/repocraft-project/server-go/internal/graceful"
	"github.com/repocraft-project/server-go/pkg/repolift"
)

func main() {
	log.SetOutput(gin.DefaultErrorWriter)

	rootFS := repolift.NewLocalFS(".repos")
	transferer := repolift.NewTransferer(rootFS)

	r := gin.Default()

	r.GET("/:user/:repo/info/refs", func(c *gin.Context) {
		path := c.Param("user") + "/" + c.Param("repo")
		service := c.Query("service")

		log.Printf("[HTTP GET info/refs] path=%s service=%s\n", path, service)

		var svc repolift.Service
		switch service {
		case "git-upload-pack":
			svc = repolift.UploadPackService
			c.Header("Content-Type", "application/x-git-upload-pack-advertisement")
		case "git-receive-pack":
			svc = repolift.ReceivePackService
			c.Header("Content-Type", "application/x-git-receive-pack-advertisement")
		default:
			log.Printf("[HTTP GET info/refs] unknown service: %s\n", service)
			c.Status(400)
			return
		}

		err := transferer.AdvertiseReferences(c.Request.Context(), path, c.Writer, svc)
		if err != nil {
			log.Printf("[HTTP GET info/refs] ERROR: %v\n", err)
		} else {
			log.Printf("[HTTP GET info/refs] SUCCESS\n")
		}
	})

	r.POST("/:user/:repo/git-upload-pack", func(c *gin.Context) {
		c.Header("Content-Type", "application/x-git-upload-pack-result")
		path := c.Param("user") + "/" + c.Param("repo")
		err := transferer.UploadPack(c.Request.Context(), path, c.Request.Body, c.Writer)
		c.Writer.Flush()
		if err != nil {
			log.Printf("[HTTP POST git-upload-pack] ERROR: %v\n", err)
		} else {
			log.Printf("[HTTP POST git-upload-pack] SUCCESS\n")
		}
	})

	r.POST("/:user/:repo/git-receive-pack", func(c *gin.Context) {
		c.Header("Content-Type", "application/x-git-receive-pack-result")
		path := c.Param("user") + "/" + c.Param("repo")
		err := transferer.ReceivePack(c.Request.Context(), path, c.Request.Body, c.Writer)
		c.Writer.Flush()
		if err != nil {
			log.Printf("[HTTP POST git-receive-pack] ERROR: %v\n", err)
		} else {
			log.Printf("[HTTP POST git-receive-pack] SUCCESS\n")
		}
	})

	graceful.Run(r, ":8080")
}
