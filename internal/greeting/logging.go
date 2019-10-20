package greeting

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/lukasjarosch/enki/logging"

	"github.com/go-enki/enki-example/internal/models"
)

type LoggingMiddleware struct {
	logger *zap.Logger
	next   Service
}

func NewLoggingMiddleware(logger *zap.Logger) Middleware {
	return func(svc Service) Service {
		return &LoggingMiddleware{logger: logger.Named("logging"), next: svc}
	}
}

func (mw *LoggingMiddleware) Hello(ctx context.Context, name string) (greeting *models.Greeting, err error) {
	logger := logging.WithContext(ctx, mw.logger)
	logger.Info("call to Hello", zap.String("name", name))

	defer func(started time.Time) {

		logger := logger.With(
			zap.String("rpc", "Hello"),
			zap.Any("greeting", greeting),
			zap.Duration("took", time.Since(started)),
		)
		if err != nil {
			logger.Info("endpoint failed", zap.Error(err))
			return
		}
		logger.Info("endpoint succeeded")
	}(time.Now())

	return mw.next.Hello(ctx, name)
}
