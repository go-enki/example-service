package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Config defines all configuration fields for the gRPC server
type Config struct {
	Port                string        `mapstructure:"port"`
	PortMetrics         int           `mapstructure:"port-metrics"`
	Hostname            string        `mapstructure:"hostname"`
	GrpcShutdownTimeout time.Duration `mapstructure:"grpc-server-shutdown-timeout"`
	HttpShutdownTimeout time.Duration `mapstructure:"http-metrics-shutdown-timeout"`
}

// Server defines the default behaviour of gRPC servers
type Server struct {
	GoogleGrpc *grpc.Server
	logger     *zap.Logger
	config     *Config
	listener   net.Listener
}

// NewServer returns a new, pre-initialized, Server instance
// The application will terminate if the server cannot bind to the configured port.
// If the application does not terminate, the port is open and a raw gRPC server has been created after
// the call of NewServer()
func NewServer(logger *zap.Logger, config *Config) *Server {
	srv := &Server{
		logger:     logger.Named("grpc"),
		config:     config,
	}

	srv.setupGrpc()

	return srv
}

// setupGrpc will create a new, raw google gRPC server as well as the listener
// If the listener cannot bind to the port, it's considered a fatal error on which
// the application will be terminated.
func (srv *Server) setupGrpc() {
	var err error

	srv.GoogleGrpc = grpc.NewServer()
	srv.listener, err = net.Listen("tcp", fmt.Sprintf(":%v", srv.config.Port))
	if err != nil {
		srv.logger.Fatal("failed to listen on port", zap.Error(err))
	}
}

// ListenAndServe ties everything together and runs the gRPC server in a separate goroutine.
// The method then blocks until the passed context is cancelled, so this method should also be started
// as goroutine if more work is needed after starting the gRPC server.
func (srv *Server) ListenAndServe(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// TODO serve in goroutine
	go func() {
		srv.logger.Info("gRPC server running", zap.String("port", viper.GetString("grpc-port")))
		if err := srv.GoogleGrpc.Serve(srv.listener); err != nil {
			srv.logger.Fatal("gRPC server crashed", zap.Error(err))
		}
	}()


	<-ctx.Done()
	srv.logger.Info("gRPC server shutdown requested")
	srv.shutdownGrpc()
}

// shutdownGrpc gracefully shuts down the gRPC server
func (srv *Server) shutdownGrpc() {
	stopped := make(chan struct{})
	go func() {
		srv.GoogleGrpc.GracefulStop()
		close(stopped)
	}()
	t := time.NewTicker(srv.config.GrpcShutdownTimeout)
	select {
	case <-t.C:
		srv.logger.Warn("gRPC server graceful shutdown timed-out")
	case <-stopped:
		srv.logger.Info("gRPC server stopped")
		t.Stop()
	}
}