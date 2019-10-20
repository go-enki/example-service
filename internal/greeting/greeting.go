package greeting

import (
	"context"

	"github.com/go-enki/enki-example/internal/models"
)


type Service interface {
	Hello(ctx context.Context, name string) (greeting *models.Greeting, err error)
}

type Middleware func(svc Service) Service


type Repository interface {
	GetHelloTemplate(name string) (format string, err error)
}