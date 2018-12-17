package models

type Account struct {
	ID        uint32   `json:"id"`                   //unique
	Email     string   `json:"email"`                //up to 100 symbols, unique
	FName     string   `json:"fname, omitempty"`     //up to 50 symbols, optional
	SName     string   `json:"sname, omitempty"`     //up to 50 symbols, optional
	Phone     string   `json:"phone, omitempty"`     //up to 16 symbols, unique, optional
	Sex       string   `json:"sex"`                  //m|f
	Birth     int      `json:"birth"`                //timestamp from 01.01.1950 to 01.01.2005
	Country   string   `json:"country, omitempty"`   //up to 50 symbols, optional
	City      string   `json:"city, omitempty"`      //up to 50 symbols, optional, every city belongs to defined country
	Joined    uint     `json:"joined"`               //timestamp from 01.01.2011 to 01.01.2018
	Status    string   `json:"status"`               //"свободны", "заняты", "всё сложно"
	Interests []string `json:"interests, omitempty"` //every string is up to 100 symbols, optional
	Premium   Premium  `json:"premium, omitempty"`
	Likes     []Like   `json:"likes, omitempty"`
}

type Premium struct {
	Start  uint `json:"start"`  //timestamp from 01.01.2018
	Finish uint `json:"finish"` //timestamp from 01.01.2018
}

type Like struct {
	ID uint32 `json:"id"` //id of the liked account
	TS int    `json:"ts"` //timestamp when like has been set
}
