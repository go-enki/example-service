package http

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func ListenAndServe(ctx context.Context, wg *sync.WaitGroup, logger *zap.Logger, handler http.Handler) {
	defer wg.Done()

	httpServer := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%s", viper.GetString("http-port")), Handler: handler}

	// serve
	go func() {
		logger.Info("http server started", zap.String("address", httpServer.Addr))
		logger.Fatal("http server crashed", zap.Error(httpServer.ListenAndServe()))
	}()

	<-ctx.Done()
	logger.Info("http server shutdown requested")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	httpServer.Shutdown(shutdownCtx)

	logger.Info("http server stopped gracefully")
}
