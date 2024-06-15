package domain

import (
	"fmt"
	"regexp"
)

type Location struct {
	CEP   string  `json:"cep"`
	City  string  `json:"city"`
	TempC float64 `json:"temp_c"`
	TempF float64 `json:"temp_f"`
	TempK float64 `json:"temp_k"`
}

func NewLocation(cep string) (*Location, error) {

	var l Location

	err := l.SetCEP(cep)
	if err != nil {
		return nil, fmt.Errorf("error to create object Location")
	}

	err = l.SetTemperatures(0)
	if err != nil {
		return nil, err
	}

	return &l, nil
}

func (l *Location) Validate() error {

	if match, _ := regexp.MatchString(`^\d{8}$`, l.CEP); !match {
		return fmt.Errorf("invalid format for cep - exaple: 12345678")
	}

	return nil
}

func (l *Location) GetCEP() string {
	return l.CEP
}

func (l *Location) GetCity() string {
	return l.City
}

func (l *Location) GetTempC() float64 {
	return l.TempC
}

func (l *Location) GetTempF() float64 {
	return l.TempF
}

func (l *Location) GetTempK() float64 {
	return l.TempK
}

func (l *Location) SetCEP(cep string) error {

	if match, _ := regexp.MatchString(`^\d{8}$`, cep); !match {
		return fmt.Errorf("invalid format for cep - example: 12345678")
	}

	l.CEP = cep
	return nil
}

func (l *Location) SetCity(city string) error {

	if len(city) < 1 {
		return fmt.Errorf(" invalid city")
	}

	l.City = city
	return nil
}

func (l *Location) SetTempC(celsius float64) error {

	l.TempC = celsius

	return nil
}

func (l *Location) setTempF() error {
	l.TempF = (l.TempC * 1.8) + 32
	return nil
}

func (l *Location) setTempK() error {
	l.TempK = l.TempC + 273
	return nil
}

func (l *Location) SetTemperatures(celsius float64) error {

	err := l.SetTempC(celsius)
	if err != nil {
		return err
	}
	err = l.setTempF()
	if err != nil {
		return err
	}
	err = l.setTempK()
	if err != nil {
		return err
	}

	return nil
}
