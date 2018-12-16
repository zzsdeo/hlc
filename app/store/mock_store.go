package store

import (
	"ProjectDB/app/models"
	"errors"
	"github.com/google/uuid"
)

type MockStore struct {
	specs []models.Spec
}

func (ms *MockStore) Initialize() {
	var dumbSpec1 = models.Spec{
		"1",
		"SKUD",
		[]models.Item{models.Item{"2", "UPS", "220", "APC", "pcs", 20}, models.Item{"3", "Controller", "Bastion", "ES-Prom", "pcs", 30}},
	}

	var dumbSpec2 = models.Spec{
		"5",
		"OOS",
		[]models.Item{models.Item{"7", "Sensor", "C2000-IK", "Bolid", "pcs", 8}, models.Item{"8", "PPKOP", "C2000M", "Bolid", "pcs", 1}},
	}

	ms.specs = []models.Spec{dumbSpec1, dumbSpec2}
}

func (ms *MockStore) GetSpecs() ([]models.Spec, error) {
	return ms.specs, nil
}

func (ms *MockStore) GetSpec(ID string) (models.Spec, error) {
	for _, spec := range ms.specs {
		if spec.ID == ID {
			return spec, nil
		}
	}
	return models.Spec{}, errors.New("spec not found")
}

func (ms *MockStore) CreateSpec(spec models.Spec) (models.Spec, error) {
	specUuid, err := uuid.NewRandom()
	if err != nil {
		return spec, err
	}

	spec.ID = specUuid.String()
	ms.specs = append(ms.specs, spec)

	return spec, nil
}

func (ms *MockStore) UpdateSpec(updatedSpec models.Spec) (models.Spec, error) {
	for i, spec := range ms.specs {
		if spec.ID == updatedSpec.ID {
			ms.specs[i] = updatedSpec
			return updatedSpec, nil
		}
	}

	return updatedSpec, errors.New("spec not found")
}

func (ms *MockStore) DeleteSpec(ID string) error {
	for i, spec := range ms.specs {
		if spec.ID == ID {
			ms.specs = append(ms.specs[:i], ms.specs[i+1:]...)
			return nil
		}
	}

	return errors.New("spec not found")
}

func (ms *MockStore) CreateItem(specID string, item models.Item) (models.Item, error) {
	itemUuid, err := uuid.NewRandom()
	if err != nil {
		return item, err
	}

	item.ID = itemUuid.String()

	spec, err := ms.findSpecByID(specID)
	if err != nil {
		return item, err
	}

	spec.Items = append(spec.Items, item)

	return item, nil
}

func (ms *MockStore) findSpecByID(id string) (*models.Spec, error) {
	for i, spec := range ms.specs {
		if spec.ID == id {
			return &ms.specs[i], nil
		}
	}

	return nil, errors.New("spec not found")
}
