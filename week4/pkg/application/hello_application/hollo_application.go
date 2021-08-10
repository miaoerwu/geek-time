package hello_application

import (
	"geek-time/week4/internal/domain/hello_domain"
)

type DTO struct {
	Content string
}

type Application struct {
	hello_domain.Domain
}

func NewApplication(domain hello_domain.Domain) Application {
	return Application{Domain: domain}
}
