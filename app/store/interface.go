package store

import "ProjectDB/app/models"

type SpecsRepository interface {
	GetSpecs() ([]models.Spec, error)
	GetSpec(ID string) (models.Spec, error)
	CreateSpec(spec models.Spec) (models.Spec, error)
	UpdateSpec(updatedSpec models.Spec) (models.Spec, error)
	DeleteSpec(ID string) error
	CreateItem(specID string, item models.Item) (models.Item, error)
}
