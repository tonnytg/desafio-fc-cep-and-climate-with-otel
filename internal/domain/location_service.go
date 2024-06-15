package domain

import (
	"fmt"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/infra/cep"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/infra/weather"
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

func (s *LocationService) Execute(l *Location) error {

	//data := s.repo.Get(l.CEP)
	//log.Println("service received:", data)

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

	wc, err := weather.GetWeather(l.GetCity())
	if err != nil {
		log.Println("error to execute and get weather for city:", city)
		return fmt.Errorf("500")
	}
	err = l.SetTemperatures(wc)
	if err != nil {
		log.Println("error to set temperatures")
		return fmt.Errorf("500")
	}

	log.Println("execute finish with success:", l)

	//err = s.repo.Save(l)
	//if err != nil {
	//	log.Printf("error to save location: %v\n", l)
	//}

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
