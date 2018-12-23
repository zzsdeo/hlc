package models

import "github.com/globalsign/mgo/bson"

type Account struct {
	MongoID   bson.ObjectId `json:"-" bson:"_id"`
	ID        uint32        `json:"id,omitempty" bson:"id,omitempty"`               //unique
	Email     string        `json:"email,omitempty" bson:"email,omitempty"`         //up to 100 symbols, unique
	FName     string        `json:"fname,omitempty" bson:"fname,omitempty"`         //up to 50 symbols, optional
	SName     string        `json:"sname,omitempty" bson:"sname,omitempty"`         //up to 50 symbols, optional
	Phone     string        `json:"phone,omitempty" bson:"phone,omitempty"`         //up to 16 symbols, unique, optional
	Sex       string        `json:"sex,omitempty" bson:"sex,omitempty"`             //m|f
	Birth     int           `json:"birth,omitempty" bson:"birth,omitempty"`         //timestamp from 01.01.1950 to 01.01.2005
	Country   string        `json:"country,omitempty" bson:"country,omitempty"`     //up to 50 symbols, optional
	City      string        `json:"city,omitempty" bson:"city,omitempty"`           //up to 50 symbols, optional, every city belongs to defined country
	Joined    uint          `json:"joined,omitempty" bson:"joined,omitempty"`       //timestamp from 01.01.2011 to 01.01.2018
	Status    string        `json:"status,omitempty" bson:"status,omitempty"`       //"свободны", "заняты", "всё сложно"
	Interests []string      `json:"interests,omitempty" bson:"interests,omitempty"` //every string is up to 100 symbols, optional
	Premium   *Premium      `json:"premium,omitempty" bson:"premium,omitempty"`
	Likes     []Like        `json:"likes,omitempty" bson:"likes,omitempty"`
}

type Premium struct {
	Start  uint `json:"start,omitempty" bson:"start,omitempty"`   //timestamp from 01.01.2018
	Finish uint `json:"finish,omitempty" bson:"finish,omitempty"` //timestamp from 01.01.2018
}

type Like struct {
	ID uint32 `json:"id,omitempty" bson:"id,omitempty"` //id of the liked account
	TS int    `json:"ts,omitempty" bson:"ts,omitempty"` //timestamp when like has been set
}

type Accounts struct {
	Accounts []Account `json:"accounts"`
}
