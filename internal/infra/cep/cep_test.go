package cep_test

import (
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/infra/cep"
	"testing"
)

func TestCepGET(t *testing.T) {

	c, err := cep.GetCity("01308080")
	if err != nil {
		t.Error("error to get info by cep")
	}

	if c != "São Paulo" {
		t.Error("error to get city")
	}
}
