package models

import (
	"errors"
	"math"
	"sort"
)

type Account struct {
	ID           int             `json:"id,omitempty" bson:"id,omitempty"`               //unique
	Email        string          `json:"email,omitempty" bson:"email,omitempty"`         //up to 100 symbols, unique
	FName        string          `json:"fname,omitempty" bson:"fname,omitempty"`         //up to 50 symbols, optional
	SName        string          `json:"sname,omitempty" bson:"sname,omitempty"`         //up to 50 symbols, optional
	Phone        string          `json:"phone,omitempty" bson:"phone,omitempty"`         //up to 16 symbols, unique, optional
	Sex          string          `json:"sex,omitempty" bson:"sex,omitempty"`             //m|f
	Birth        int             `json:"birth,omitempty" bson:"birth,omitempty"`         //timestamp from 01.01.1950 to 01.01.2005
	Country      string          `json:"country,omitempty" bson:"country,omitempty"`     //up to 50 symbols, optional
	City         string          `json:"city,omitempty" bson:"city,omitempty"`           //up to 50 symbols, optional, every city belongs to defined country
	Joined       int             `json:"joined,omitempty" bson:"joined,omitempty"`       //timestamp from 01.01.2011 to 01.01.2018
	Status       string          `json:"status,omitempty" bson:"status,omitempty"`       //"свободны", "заняты", "всё сложно"
	Interests    []string        `json:"interests,omitempty" bson:"interests,omitempty"` //every string is up to 100 symbols, optional
	interestsMap map[string]bool `json:"-"`
	Premium      *Premium        `json:"premium,omitempty" bson:"premium,omitempty"`
	Likes        []Like          `json:"likes,omitempty" bson:"likes,omitempty"`
	likesMap     map[int][]int   `json:"-"`
}

type Premium struct {
	Start  int `json:"start,omitempty" bson:"start,omitempty"`   //timestamp from 01.01.2018
	Finish int `json:"finish,omitempty" bson:"finish,omitempty"` //timestamp from 01.01.2018
}

type Like struct {
	ID int `json:"id,omitempty" bson:"id,omitempty"` //id of the liked account
	TS int `json:"ts,omitempty" bson:"ts,omitempty"` //timestamp when like has been set
}

type Accounts struct {
	Accounts []Account `json:"accounts"`
}

func (a *Accounts) ExtractAccountByID(id int) (Account, error) {
	for i, account := range a.Accounts {
		if account.ID == id {
			a.Accounts = append(a.Accounts[:i], a.Accounts[i+1:]...)
			return account, nil
		}
	}
	return Account{}, errors.New("account not found")
}

func (a *Account) PrepareInterestsMap() {
	a.interestsMap = make(map[string]bool)
	for _, i := range a.Interests {
		a.interestsMap[i] = true
	}
}

func (a *Account) CheckCompatibility(account Account, now int) int {
	var compatibility int
	if account.Birth > a.Birth {
		compatibility = 1000000000 - account.Birth + a.Birth
	} else {
		compatibility = 1000000000 - a.Birth + account.Birth
	}

	for _, interest := range account.Interests {
		if _, ok := a.interestsMap[interest]; ok {
			compatibility += 10000000000
		}
	}

	switch account.Status {
	case "свободны":
		compatibility += 300000000000
	case "всё сложно":
		compatibility += 200000000000
	case "заняты":
		compatibility += 100000000000
	}

	if account.PremiumNow(now) {
		compatibility += 1000000000000
	}

	return compatibility
}

//func (a *Account) CheckCompatibility(account Account, now int) string {
//	compatibility := "0"
//
//	if account.PremiumNow(now) {
//		compatibility = "1"
//	}
//
//	switch account.Status {
//	case "свободны":
//		compatibility += "3"
//	case "всё сложно":
//		compatibility += "2"
//	case "заняты":
//		compatibility += "1"
//	}
//
//	interestsCount := 0
//	for _, interest := range account.Interests {
//		if _, ok := a.interestsMap[interest]; ok {
//			interestsCount++
//		}
//	}
//	if interestsCount < 10 {
//		compatibility += "0"
//	}
//	compatibility += strconv.Itoa(interestsCount)
//
//	diff := account.Birth - a.Birth
//	if diff < 0 {
//		compatibility += strconv.Itoa(1000000000 + diff)
//	} else {
//		compatibility += strconv.Itoa(1000000000 - diff)
//	}
//
//	return compatibility
//}

func (a *Account) PrepareLikesMap() {
	a.likesMap = make(map[int][]int)
	for _, like := range a.Likes {
		a.likesMap[like.ID] = append(a.likesMap[like.ID], like.TS)
	}
}

func (a *Account) CheckSimilarity(account Account) float64 {
	var similarity float64
	account.PrepareLikesMap()
	var avrLikes, avrMyLikes float64
	for k, likes := range account.likesMap {
		avrLikes, avrMyLikes = 0, 0
		if myLikes, ok := a.likesMap[k]; ok {
			for _, myLike := range myLikes {
				avrMyLikes += float64(myLike)
			}
			avrMyLikes /= float64(len(myLikes))

			for _, like := range likes {
				avrLikes += float64(like)
			}
			avrLikes /= float64(len(likes))

			if avrMyLikes == avrLikes {
				similarity += 1
				continue
			}

			similarity += 1 / math.Abs(avrMyLikes-avrLikes)
		}
	}
	return similarity
}

func (a *Account) GetNewIds(account Account) []int {
	var ids []int
	for _, like := range account.Likes {
		if _, ok := a.likesMap[like.ID]; !ok {
			ids = append(ids, like.ID)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] > ids[j]
	})
	return ids
}

func (a *Account) InterestsContains(interests []string) bool {
	if len(interests) == 0 {
		return false
	}
	for _, interest := range interests {
		if _, ok := a.interestsMap[interest]; !ok {
			return false
		}
	}
	return true
}

func (a *Account) InterestsAny(interests []string) bool {
	if len(interests) == 0 {
		return false
	}
	for _, interest := range interests {
		if _, ok := a.interestsMap[interest]; ok {
			return true
		}
	}
	return false
}

func (a *Account) LikesContains(likes []int) bool {
	if len(likes) == 0 {
		return false
	}
	for _, like := range likes {
		if _, ok := a.likesMap[like]; !ok {
			return false
		}
	}
	return true
}

func (a *Account) PremiumNow(now int) bool {
	if a.Premium == nil {
		return false
	}
	if a.Premium.Start < now && a.Premium.Finish > now {
		return true
	}
	return false
}
