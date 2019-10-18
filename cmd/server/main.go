package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	enkigrpc "github.com/go-enki/enki-example/pkg/grpc"
	"github.com/go-enki/enki-example/pkg/logging"
	"github.com/go-enki/enki-example/pkg/signals"
)

var logger *zap.Logger

// flags defines the FlagSet of the server, parses and binds them to viper and env vars.
// Flags are bound to viper and thus are available throughout the application
// Flags can also be set through environment variables: 'log-level' => 'LOG_LEVEL
func setupFlags() {
	fs := pflag.NewFlagSet("server", pflag.ContinueOnError)

	fs.String("log-level", "info", "log level debug, info, warn, error, flat or panic")
	fs.String("grpc-port", "50051", "the port on which the gRPC server is exposed")

	// gRPC flags
	fs.Int("port-metrics", 3000, "HTTP metrics server port")
	fs.Duration("grpc-server-shutdown-timeout", 5*time.Second, "gRPC server graceful shutdown duration")
	fs.Duration("http-metrics-shutdown-timeout", 3*time.Second, "HTTP metrics erver graceful shutdown duration")

	parseFlags(fs)
	bindFlags(fs)
}

func main() {
	setupFlags()

	// setup logger
	logger, _ = logging.SetupZapLogger(viper.GetString("log-level"))
	logger = logger.Named("server")

	// setup concurrency sync stuff
	stopChan := signals.SetupSignalHandler()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}

	// todo: setup datastore
	// todo: setup service layer

	// setup gRPC server
	var grpcConfig enkigrpc.Config
	if err := viper.Unmarshal(&grpcConfig); err != nil {
		logger.Fatal("failed to unmarshal gRPC configuration", zap.Error(err))
	}
	grpc := enkigrpc.NewServer(logger, &grpcConfig)

	// start working
	wg.Add(1)
	go grpc.ListenAndServe(ctx, &wg)

	// wait for signal handler to fire and shutdown
	<-stopChan
	cancel()
	wg.Wait()
	logger.Info("all goroutines finished")
}

// parseFlags parses the flags passed to the binary.
// If an error occurs, the error is logged and the help displayed.
func parseFlags(fs *pflag.FlagSet) {
	err := fs.Parse(os.Args[1:])
	switch {
	case err == pflag.ErrHelp:
		os.Exit(0)
	case err != nil:
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
		fs.PrintDefaults()
		os.Exit(2)
	}
}

// bindFlags will bind the configured flags to environment variables, prefixed with 'EnvPrefix'.
// It will also set the 'hostname' field in viper in case it's needed.
func bindFlags(fs *pflag.FlagSet) {
	_ = viper.BindPFlags(fs)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}
