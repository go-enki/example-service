package greeting

import (
	"context"

	"github.com/lukasjarosch/enki/logging"
	"go.uber.org/zap"

	"github.com/go-enki/enki-example/internal/models"
)

type ServiceImplementation struct {
	logger *zap.Logger
	repository Repository
}

func NewServiceImplementation(logger *zap.Logger, repository Repository) Service {
	return &ServiceImplementation{logger:logger, repository:repository}
}

func (svc *ServiceImplementation) Hello(ctx context.Context, name string) (greeting *models.Greeting, err error)  {
	logger := logging.WithContext(ctx, svc.logger)

	greeting = &models.Greeting{
		Name:     name,
	}
	if err := greeting.Validate(); err != nil {
		return nil, models.ErrGreetingEmptyName
	}

	template, err := svc.repository.GetHelloTemplate(greeting.Name)
	if err != nil {
	    return nil, models.ErrGetHelloTemplate
	}
	greeting.Template = template
	greeting.Rendered = greeting.Render()
	logger.Info("rendered greeting", zap.Any("greeting", greeting))

	return greeting, nil
}
