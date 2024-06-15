package domain_test

import (
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/domain"
	"log"
	"testing"
)

func TestLocationConstructor(t *testing.T) {

	l, err := domain.NewLocation("12345678")
	if err != nil {
		t.Errorf("location constructor cannot return error")
	}

	err = l.SetCEP("12345678")
	if err != nil {
		t.Errorf("SetCEP cannot return error")
	}
}

func TestLocationTemperatureF(t *testing.T) {

	l, err := domain.NewLocation("12345678")
	if err != nil {
		t.Errorf("location constructor cannot return error")
	}

	fahrenheit := l.GetTempF()
	log.Println(fahrenheit)
	if fahrenheit != 32.0 {
		t.Errorf("location constructor cannot return error")
	}
}

func TestLocationTemperatureK(t *testing.T) {

	l, err := domain.NewLocation("12345678")
	if err != nil {
		t.Errorf("location constructor cannot return error")
	}

	kelvin := l.GetTempK()
	if kelvin != 273.0 {
		t.Errorf("location constructor cannot return error")
	}
}
