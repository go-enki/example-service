package main

import (
	"context"
	"fmt"
	stdhttp "net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/go-enki/enki-example/pkg/grpc"
	"github.com/go-enki/enki-example/pkg/http"
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

	// HTTP server general flags
	fs.Duration("http-grace-period", 5*time.Second, "gRPC server grace period when shutting down")

	// HTTP debug server flags
	fs.String("http-debug-port", "3000", "the port on which the debug HTTP server is bound")

	// gRPC general flags
	fs.String("grpc-port", "50051", "the port on which the gRPC server is exposed")
	fs.Duration("grpc-grace-period", 5*time.Second, "gRPC server grace period when shutting down")

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
	var grpcConfig grpc.Config
	if err := viper.Unmarshal(&grpcConfig); err != nil {
		logger.Fatal("failed to unmarshal gRPC configuration", zap.Error(err))
	}
	grpc := grpc.NewServer(logger, &grpcConfig)

	// setup debug HTTP server
	httpDebug := http.NewServer(logger, &http.Config{
		Port:        viper.GetString("http-debug-port"),
		GracePeriod: viper.GetDuration("http-grace-period"),
	})
	debugMux := &stdhttp.ServeMux{}
	debugMux.Handle("/grpc/health", grpc.Health())
	debugMux.Handle("/debug/health", httpDebug.Health())

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
