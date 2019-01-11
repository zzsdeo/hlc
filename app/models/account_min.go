package models

type AccountMin struct {
	Email     string //up to 100 symbols, unique
	FName     int    //up to 50 symbols, optional
	SName     int    //up to 50 symbols, optional
	Phone     string //up to 16 symbols, unique, optional
	Sex       string //0-m 1-f
	Birth     int    //timestamp from 01.01.1950 to 01.01.2005
	Country   int    //up to 50 symbols, optional
	City      int    //up to 50 symbols, optional, every city belongs to defined country
	Joined    int    //timestamp from 01.01.2011 to 01.01.2018
	Status    byte   //0-"свободны", 1-"заняты", 2-"всё сложно"
	Interests []int  //every string is up to 100 symbols, optional
	Premium   *Premium
	Likes     []Like
}

func (a *AccountMin) PremiumNow(now int) bool {
	if a.Premium == nil {
		return false
	}
	if a.Premium.Start < now && a.Premium.Finish > now {
		return true
	}
	return false
}
