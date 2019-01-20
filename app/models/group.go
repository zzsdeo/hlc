package models

var Keys = map[string]struct{}{"sex": {}, "status": {}, "interests": {}, "country": {}, "city": {}}

type Group struct {
	Sex       string `json:"sex,omitempty"`
	Status    string `json:"status,omitempty"`
	Interests string `json:"interests,omitempty"`
	Country   string `json:"country,omitempty"`
	City      string `json:"city,omitempty"`
	Count     int    `json:"count"`
}

type Groups struct {
	Groups []Group `json:"groups"`
}
