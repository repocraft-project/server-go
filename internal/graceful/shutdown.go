package graceful

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	DefaultShutdownTimeout = 10 * time.Second
)

func Run(engine *gin.Engine, addr string) error {
	return RunWithTimeout(DefaultShutdownTimeout, engine, addr)
}

func RunWithTimeout(timeout time.Duration, engine *gin.Engine, addr string) error {
	return run(engine, addr, func(srv *http.Server) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return nil
	})
}

func RunWithContext(ctx context.Context, engine *gin.Engine, addr string) error {
	return run(engine, addr, func(srv *http.Server) error {
		return srv.Shutdown(ctx)
	})
}

func run(engine *gin.Engine, addr string, doShutdown func(*http.Server) error) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	errCh := make(chan error, 1)
	quit := make(chan os.Signal, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-quit:
		if err := doShutdown(srv); err != nil {
			return err
		}
		return nil
	}
}
