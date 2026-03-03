package main

import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v6/osfs"
	"github.com/go-git/go-git/v6/plumbing/transport"
	"github.com/repocraft-project/server-go/internal/graceful"
	"github.com/repocraft-project/server-go/pkg/repolift"
)

func main() {
	log.SetOutput(gin.DefaultErrorWriter)

	fs := osfs.New(".repos")
	transferer := repolift.NewDefaultTransferer(fs)

	r := gin.Default()

	r.GET("/:user/:repo/info/refs", func(c *gin.Context) {
		service := c.Query("service")
		if service == "" {
			c.String(400, "service parameter required")
			return
		}

		repo := cleanRepo(c.Param("user"), c.Param("repo"))
		if err := handleInfoRefs(c, repo, service, transferer); err != nil {
			log.Println("info refs error:", err)
		}
	})

	r.POST("/:user/:repo/git-upload-pack", func(c *gin.Context) {
		repo := cleanRepo(c.Param("user"), c.Param("repo"))
		if err := handleUploadPack(c, repo, transferer); err != nil {
			log.Println("upload pack error:", err)
		}
	})

	r.POST("/:user/:repo/git-receive-pack", func(c *gin.Context) {
		repo := cleanRepo(c.Param("user"), c.Param("repo"))
		if err := handleReceivePack(c, repo, transferer); err != nil {
			log.Println("receive pack error:", err)
		}
	})

	if err := graceful.Run(r, ":8080"); err != nil {
		log.Fatalln(err)
	}
	log.Println("Done")
}

func cleanRepo(user, repo string) string {
	repo = strings.TrimSuffix(repo, ".git")
	return filepath.Join(user, repo)
}

func handleInfoRefs(c *gin.Context, repo, service string, transferer *repolift.DefaultTransferer) error {
	sto, err := transferer.OpenRepo(repo)
	if err != nil {
		c.Header("Content-Type", "text/plain")
		c.String(404, "Repository not found")
		return err
	}

	var svc transport.Service
	switch service {
	case "git-upload-pack":
		svc = transport.UploadPackService
	case "git-receive-pack":
		svc = transport.ReceivePackService
	default:
		c.String(400, "Unsupported service: %s", service)
		return nil
	}

	c.Header("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", svc.Name()))
	return transport.AdvertiseReferences(c.Request.Context(), sto, c.Writer, svc, true)
}

func handleUploadPack(c *gin.Context, repo string, transferer *repolift.DefaultTransferer) error {
	return transferer.UploadPack(c.Request.Context(), repo, c.Request.Body, &ginWriter{c.Writer}, &repolift.UploadPackOptions{
		StatelessRPC: true,
	})
}

func handleReceivePack(c *gin.Context, repo string, transferer *repolift.DefaultTransferer) error {
	return transferer.ReceivePack(c.Request.Context(), repo, c.Request.Body, &ginWriter{c.Writer}, &repolift.ReceivePackOptions{
		StatelessRPC: true,
	})
}

type ginWriter struct {
	gin.ResponseWriter
}

func (w *ginWriter) Write(p []byte) (int, error) {
	return w.ResponseWriter.Write(p)
}

func (w *ginWriter) Close() error {
	return nil
}

var _ io.WriteCloser = (*ginWriter)(nil)
