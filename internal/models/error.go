package models

import "errors"

var (
	ErrGreetingEmptyName = errors.New("greeting.name cannot be empty")
	ErrGetHelloTemplate  = errors.New("cannot get hello template for name")
)
