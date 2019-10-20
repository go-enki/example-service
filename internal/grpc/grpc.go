package grpc

import (
	"go.uber.org/zap"

	"github.com/go-enki/enki-example/internal/greeting"
)

type ExampleService struct {
	logger   *zap.Logger
	greeting greeting.Service
}

func NewExampleService(logger *zap.Logger, greetingService greeting.Service) *ExampleService {
	return &ExampleService{
		logger:   logger.Named("example"),
		greeting: greetingService,
	}
}

// TODO:
// + shared <-> internal mappings
// + error mapping
