package main

import (
	"context"
	stdhttp "net/http"
	"sync"
	"time"

	"github.com/lukasjarosch/enki/config"
	"github.com/lukasjarosch/enki/logging"
	"github.com/lukasjarosch/enki/server"
	"github.com/lukasjarosch/enki/signals"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/go-enki/enki-example/internal/database"
	"github.com/go-enki/enki-example/internal/greeting"
	grpc2 "github.com/go-enki/enki-example/internal/grpc"
	example "github.com/go-enki/enki-example/proto"
)

var logger *zap.Logger

// setupFlags defines the FlagSet of the server, parses and binds them to viper and env vars.
// Flags are bound to viper and thus are available throughout the application
// Flags can also be set through environment variables: 'log-level' => 'LOG_LEVEL
func setupFlags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("server", pflag.ContinueOnError)

	fs.String("log-level", "info", "log level debug, info, warn, error, flat or panic")

	// HTTP server general flags
	fs.Duration("http-grace-period", 5*time.Second, "gRPC server grace period when shutting down")

	// HTTP debug server flags
	fs.String("http-debug-port", "3000", "the port on which the debug HTTP server is bound")

	// gRPC general flags
	fs.String("grpc-port", "50051", "the port on which the gRPC server is exposed")
	fs.Duration("grpc-grace-period", 5*time.Second, "gRPC server grace period when shutting down")

	return fs
}

func main() {
	config.ParseFlagSet(setupFlags())

	// setup logger
	logger, _ = logging.NewZapLogger(viper.GetString("log-level"))
	logger = logger.Named("server")

	// setup concurrency sync stuff
	stopChan := signals.SetupSignalHandler()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}

	// todo: setup datastore
	// todo: setup service layer
	// setup domain-level services and their middleware
	greetingService := greeting.NewServiceImplementation(logger.Named("greeting"), database.NewInMemDB())
	greetingService = greeting.NewLoggingMiddleware(logger.Named("middleware"))(greetingService)

	// setup gRPC server
	var grpcConfig server.GrpcConfig
	if err := viper.Unmarshal(&grpcConfig); err != nil {
		logger.Fatal("failed to unmarshal gRPC configuration", zap.Error(err))
	}
	grpc := server.NewGrpcServer(logger, &grpcConfig)
	example.RegisterExampleServiceServer(grpc.GoogleGrpc, grpc2.NewExampleService(logger.Named("grpc"), greetingService))

	// setup debug HTTP server
	httpDebug := server.NewHttpServer(logger, &server.HttpConfig{
		Port:        viper.GetString("http-debug-port"),
		GracePeriod: viper.GetDuration("http-grace-period"),
	})
	debugMux := &stdhttp.ServeMux{}
	debugMux.Handle("/grpc/health", grpc.Health())
	debugMux.Handle("/debug/health", httpDebug.Health())
	debugMux.Handle("/metrics", promhttp.Handler())

	// start working
	wg.Add(1)
	go grpc.ListenAndServe(ctx, &wg)
	wg.Add(1)
	go httpDebug.ListenAndServe(ctx, &wg, debugMux)

	// wait for signal handler to fire and shutdown
	<-stopChan
	cancel()
	wg.Wait()
	logger.Info("all goroutines finished")
}
