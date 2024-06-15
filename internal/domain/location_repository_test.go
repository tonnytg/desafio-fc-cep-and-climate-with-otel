package domain_test

import (
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/domain"
	"testing"
)

func TestLocationRepositoryGet(t *testing.T) {

	repo := domain.NewLocationRepository()

	data := repo.Get("12345678")
	if data == nil {
		t.Errorf("error, data cannot be nil")
	}

}

func TestLocationRepositorySave(t *testing.T) {

	repo := domain.NewLocationRepository()

	l, _ := domain.NewLocation("12345678")

	err := repo.Save(l)
	if err != nil {
		t.Errorf("error, method save in repository return %v\n", err)
	}

}
