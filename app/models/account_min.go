package models

import (
	"math"
	"sort"
	"time"
)

type AccountMin struct {
	ID        int
	Email     string  //up to 100 symbols, unique
	FName     uint8   //up to 50 symbols, optional
	SName     uint16  //up to 50 symbols, optional
	Phone     string  //up to 16 symbols, unique, optional
	Sex       byte    //0-m 1-f
	Birth     int     //timestamp from 01.01.1950 to 01.01.2005
	Country   uint8   //up to 50 symbols, optional
	City      uint16  //up to 50 symbols, optional, every city belongs to defined country
	Joined    int     //timestamp from 01.01.2011 to 01.01.2018
	Status    byte    //0-"свободны", 1-"заняты", 2-"всё сложно"
	Interests []uint8 //every string is up to 100 symbols, optional
	Premium   *Premium
	//Likes     []Like
	//likesMap  map[int][]int
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

func (a *AccountMin) CheckBirth(year int) bool {
	if time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC).Unix() <= int64(a.Birth) &&
		int64(a.Birth) < time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC).Unix() {
		return true
	}
	return false
}

func (a *AccountMin) CheckJoined(year int) bool {
	if time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC).Unix() <= int64(a.Joined) &&
		int64(a.Joined) < time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC).Unix() {
		return true
	}
	return false
}

func (a *AccountMin) CheckCompatibility(account AccountMin, now int) int {
	var compatibility int
	if account.Birth > a.Birth {
		compatibility = 1000000000 - account.Birth + a.Birth
	} else {
		compatibility = 1000000000 - a.Birth + account.Birth
	}

	for i := range account.Interests {
		for ii := range a.Interests {
			if account.Interests[i] == a.Interests[ii] {
				compatibility += 10000000000
				break
			}
		}
	}

	switch account.Status {
	case 0:
		compatibility += 300000000000
	case 2:
		compatibility += 200000000000
	case 1:
		compatibility += 100000000000
	}

	if account.PremiumNow(now) {
		compatibility += 1000000000000
	}

	return compatibility
}

func (a *AccountMin) CheckSimilarity(account AccountMin) float64 {
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

func (a *AccountMin) GetNewIds(account AccountMin) []int {
	var ids []int
	account.PrepareLikesMap()
	for k, _ := range account.likesMap {
		if _, ok := a.likesMap[k]; !ok {
			ids = append(ids, k)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] > ids[j]
	})
	return ids
}
