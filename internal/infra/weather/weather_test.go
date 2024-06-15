package weather_test

import (
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/infra/weather"
	"log"
	"os"
	"strings"
	"testing"
)

func TestWeatherGet(t *testing.T) {

	// Lê o conteúdo do arquivo .env
	b, err := os.ReadFile("../../../.env")
	if err != nil {
		t.Errorf("file .env not found")
	}

	env := string(b)
	log.Println("env:", env)

	// Remove o caractere % do final da linha, se existir
	env = strings.TrimRight(env, "%\n")

	// Divide a string em chave e valor
	list := strings.Split(env, "=")
	if list[0] == "WEATHER_API_KEY" {

		secret := strings.Trim(list[1], "\"")
		err = os.Setenv(list[0], secret)
		if err != nil {
			log.Println("error to set env:", err)
		}
	}

	checkApiToken := os.Getenv("WEATHER_API_KEY")
	log.Println("checkApiToken:", checkApiToken)
	if checkApiToken == "" {
		t.Error("error to load weather api key, check end of line in .env with cat if has %")
	}

	wc, err := weather.GetWeather("São Paulo")
	if err != nil {
		t.Error("error to get weather")
	}

	if wc < 1 {
		t.Error("sorry but something wrong with São Paulo")
	}
}
