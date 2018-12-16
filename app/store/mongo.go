package store

import (
	"ProjectDB/app/models"
	"errors"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	dbName              = "ProjectDB"
	specsCollectionName = "specs"
)

type MongoStore struct {
	mgoSession *mgo.Session
}

func (ms *MongoStore) Initialize(url string) error {
	session, err := mgo.Dial(url)
	if err != nil {
		return err
	}

	ms.mgoSession = session

	return nil
}

func (ms *MongoStore) getSessionAndSpecsCollection() (*mgo.Session, *mgo.Collection, error) {
	if ms.mgoSession == nil {
		return nil, nil, errors.New("no session found")
	}

	session := ms.mgoSession.Copy()
	specsCollection := session.DB(dbName).C(specsCollectionName)

	return session, specsCollection, nil
}

func (ms *MongoStore) GetSpecs() ([]models.Spec, error) {
	session, collection, err := ms.getSessionAndSpecsCollection()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var specs []models.Spec

	err = collection.Find(bson.M{}).All(&specs)
	if err != nil {
		return nil, err
	}

	return specs, nil
}

func (ms *MongoStore) GetSpec(ID string) (models.Spec, error) {
	session, collection, err := ms.getSessionAndSpecsCollection()
	if err != nil {
		return models.Spec{}, err
	}
	defer session.Close()

	spec := models.Spec{}

	err = collection.Find(bson.M{"_id": ID}).One(&spec)
	if err != nil {
		return spec, err
	}

	return spec, nil
}

func (ms *MongoStore) CreateSpec(spec models.Spec) (models.Spec, error) {
	session, collection, err := ms.getSessionAndSpecsCollection()
	if err != nil {
		return spec, err
	}
	defer session.Close()

	spec.ID = bson.NewObjectId().Hex()

	err = collection.Insert(&spec)
	if err != nil {
		return spec, err
	}

	return spec, nil
}

func (ms *MongoStore) UpdateSpec(updatedSpec models.Spec) (models.Spec, error) {
	session, collection, err := ms.getSessionAndSpecsCollection()
	if err != nil {
		return updatedSpec, err
	}
	defer session.Close()

	err = collection.Insert(&updatedSpec)
	if err != nil {
		return updatedSpec, err
	}

	return updatedSpec, nil
}

func (ms *MongoStore) DeleteSpec(ID string) error {
	session, collection, err := ms.getSessionAndSpecsCollection()
	if err != nil {
		return err
	}
	defer session.Close()

	err = collection.RemoveId(ID)
	if err != nil {
		return err
	}

	return nil
}

func (ms *MongoStore) CreateItem(specID string, item models.Item) (models.Item, error) {
	session, collection, err := ms.getSessionAndSpecsCollection()
	if err != nil {
		return item, err
	}
	defer session.Close()

	spec := models.Spec{}

	err = collection.Find(bson.M{"_id": specID}).One(&spec)
	if err != nil {
		return item, err
	}

	item.ID = bson.NewObjectId().Hex()

	spec.Items = append(spec.Items, item)

	err = collection.Insert(&spec)
	if err != nil {
		return item, err
	}

	return item, nil
}
