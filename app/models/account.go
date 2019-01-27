package models

type Account struct {
	ID        int      `json:"id,omitempty"`        //unique
	Email     string   `json:"email,omitempty"`     //up to 100 symbols, unique
	FName     string   `json:"fname,omitempty"`     //up to 50 symbols, optional
	SName     string   `json:"sname,omitempty"`     //up to 50 symbols, optional
	Phone     string   `json:"phone,omitempty"`     //up to 16 symbols, unique, optional
	Sex       string   `json:"sex,omitempty"`       //m|f
	Birth     int      `json:"birth,omitempty"`     //timestamp from 01.01.1950 to 01.01.2005
	Country   string   `json:"country,omitempty"`   //up to 50 symbols, optional
	City      string   `json:"city,omitempty"`      //up to 50 symbols, optional, every city belongs to defined country
	Joined    int      `json:"joined,omitempty"`    //timestamp from 01.01.2011 to 01.01.2018
	Status    string   `json:"status,omitempty"`    //"свободны", "заняты", "всё сложно"
	Interests []string `json:"interests,omitempty"` //every string is up to 100 symbols, optional
	Premium   *Premium `json:"premium,omitempty"`
	Likes     []Like   `json:"likes,omitempty"`
}

type Premium struct {
	Start  int `json:"start,omitempty"`  //timestamp from 01.01.2018
	Finish int `json:"finish,omitempty"` //timestamp from 01.01.2018
}

type Like struct {
	ID int `json:"id,omitempty"` //id of the liked account
	TS int `json:"ts,omitempty"` //timestamp when like has been set
}

type Accounts struct {
	Accounts []Account `json:"accounts"`
}

func (a *Account) PremiumNow(now int) int {
	if a.Premium == nil {
		return 0
	}
	if a.Premium.Start <= now && a.Premium.Finish >= now {
		return 2
	}
	return 1
}
