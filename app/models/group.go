package models

var Keys = map[string]bool{"sex": true, "status": true, "interests": true, "country": true, "city": true}

type Group struct {
	Sex       string `json:"sex,omitempty" bson:"sex,omitempty"`
	Status    string `json:"status,omitempty" bson:"status,omitempty"`
	Interests string `json:"interests,omitempty" bson:"interests,omitempty"`
	Country   string `json:"country,omitempty" bson:"country,omitempty"`
	City      string `json:"city,omitempty" bson:"city,omitempty"`
	Count     int    `json:"count" bson:"count"`
}

type Groups struct {
	Groups []Group `json:"groups"`
}
