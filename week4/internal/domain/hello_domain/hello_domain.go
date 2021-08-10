package hello_domain

import "geek-time/week4/internal/infrastructure/hello_infrastructure"

type DO struct {
	Name string
}

type Domain struct {
	hello_infrastructure.DB
}

func (domain Domain) Handle(do DO) DO {
	return DO{"domain_" + do.Name}
}

func NewDomain(db hello_infrastructure.DB) Domain {
	return Domain{DB: db}
}
