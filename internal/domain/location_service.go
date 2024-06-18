package domain

import (
	"context"
	"fmt"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/infra/cep"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/infra/otel"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/infra/weather"
	"go.opentelemetry.io/otel/attribute"
	"log"
)

type LocationService struct {
	repo LocationRepositoryInterface
}

type LocationServiceInterface interface{}

func NewLocationService(repo LocationRepositoryInterface) *LocationService {
	return &LocationService{
		repo: repo,
	}
}

func (s *LocationService) Execute(ctx context.Context, l *Location) error {

	tr := otel.GetTracer()
	ctxSpanFull, spanFull := tr.Start(ctx, "Execute CEP and Weather: getting information")
	defer spanFull.End()

	//data := s.repo.Get(l.CEP)
	//log.Println("service received:", data)

	ctxSpanCep, spanCep := tr.Start(ctxSpanFull, "Execute CEP: getting information about cep")
	defer spanCep.End()

	city, err := cep.GetCity(l.GetCEP())
	if err != nil {
		log.Println("error to get cep:", l.GetCEP())
		otel.Logger(ctxSpanCep, "status: error to get cep")
		spanCep.SetAttributes(attribute.String("action", "get cep"), attribute.String("status", "failed"))
		return fmt.Errorf("404")
	}

	if city == "" {
		log.Println("error to get cep:", l.GetCEP())
		otel.Logger(ctxSpanCep, "status: error to get city")
		spanCep.SetAttributes(attribute.String("action", "get city"), attribute.String("status", "failed"))
		return fmt.Errorf("404")
	}
	spanCep.SetAttributes(attribute.String("action", "get city"), attribute.String("status", "success"))

	err = l.SetCity(city)
	if err != nil {
		return fmt.Errorf("500")
	}

	ctxSpanWeather, spanWeather := tr.Start(ctxSpanFull, "Execute WEATHER: getting information about city")
	defer spanWeather.End()

	wc, err := weather.GetWeather(l.GetCity())
	if err != nil {
		log.Println("error to execute and get weather for city:", city)
		otel.Logger(ctxSpanWeather, "status: error to get weather")
		spanWeather.SetAttributes(attribute.String("action", "get weather"), attribute.String("status", "failed"))
		return fmt.Errorf("500")
	}
	err = l.SetTemperatures(wc)
	if err != nil {
		log.Println("error to set temperatures")
		otel.Logger(ctxSpanWeather, "status: error to set temperatures")
		spanWeather.SetAttributes(attribute.String("action", "set temperature"), attribute.String("status", "failed"))
		return fmt.Errorf("500")
	}
	spanWeather.SetAttributes(attribute.String("action", "get weather"), attribute.String("status", "success"))

	log.Println("execute finish with success:", l)
	spanFull.SetAttributes(attribute.String("status", "success"))
	return nil
}

func (s *LocationService) GetCEP(l *Location) error {

	city, err := cep.GetCity(l.GetCEP())
	if err != nil {
		log.Println("error to get cep:", l.GetCEP())
		return fmt.Errorf("404")
	}

	if city == "" {
		log.Println("error to get cep:", l.GetCEP())
		return fmt.Errorf("404")
	}

	err = l.SetCity(city)
	if err != nil {
		return fmt.Errorf("500")
	}

	return nil
}

func (s *LocationService) GetWeather(l *Location) error {

	if l.GetCity() == "" {
		log.Println("error to get city from location:", l)
		return fmt.Errorf("404")
	}

	city := l.GetCity()

	wc, err := weather.GetWeather(l.GetCity())
	if err != nil {
		log.Println("error to execute and get weather for city:", city)
		return fmt.Errorf("500")
	}
	_ = l.SetTemperatures(wc)

	err = s.repo.Save(l)
	if err != nil {
		log.Printf("error to save location: %v\n", l)
	}

	return nil
}
