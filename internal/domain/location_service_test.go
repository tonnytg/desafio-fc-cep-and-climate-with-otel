package domain_test

import (
	"context"
	_ "embed"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/domain"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/infra/otel"
	"os"
	"strings"
	"testing"
)

func TestExecute(t *testing.T) {

	cleanup := otel.InitTracerTest()
	defer cleanup()

	//os.Setenv("WEATHER_API_KEY", "011d847082bc437cbcc192904241206")

	b, err := os.ReadFile("../../.env")
	if err != nil {
		t.Errorf("file .env not found")
	}

	env := string(b)
	list := strings.Split(env, "=")
	if list[0] == "WEATHER_API_KEY" {
		_ = os.Setenv(list[0], list[1])
	}

	r := domain.NewLocationRepository()
	s := domain.NewLocationService(r)

	l, _ := domain.NewLocation("05541000")

	var ctx context.Context

	s.Execute(ctx, l)
}
