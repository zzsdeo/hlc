package models

type Spec struct {
	ID         string `json:"id" bson:"_id"`
	SystemName string `json:"system_name" bson:"system_name"`
	Items      []Item `json:"items" bson:"items"`
}

type Item struct {
	ID         string  `json:"id" bson:"_id"`
	Name       string  `json:"name" bson:"name"`
	PartNumber string  `json:"part_number" bson:"part_number"`
	Vendor     string  `json:"vendor" bson:"vendor"`
	Measure    string  `json:"measure" bson:"measure"`
	Quantity   float64 `json:"quantity" bson:"quantity"`
}
