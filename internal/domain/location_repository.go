package domain

import "log"

type LocationRepository struct{}

type LocationRepositoryInterface interface {
	Get(cep string) *Location
	Save(*Location) error
}

func NewLocationRepository() *LocationRepository {
	return &LocationRepository{}
}

func (lr *LocationRepository) Get(cep string) *Location {

	log.Println("repository get:", cep)

	location, err := NewLocation(cep)
	if err != nil {
		return nil
	}

	return location

}

func (lr *LocationRepository) Save(location *Location) error {

	log.Println("repository save:", location)

	return nil
}
