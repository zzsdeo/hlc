package store

import (
	"hlc/app/models"
	"hlc/app/utils"
	"log"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type minData struct {
	accountsMin []models.AccountMin
	fnames      map[uint8]string
	snames      map[uint16]string
	sex         map[byte]string
	countries   map[uint8]string
	cities      map[uint16]string
	status      map[byte]string
	interests   map[uint8]string
}

type DB struct {
	minData
	mu *sync.Mutex
}

type void struct{}

type M map[string]interface{}

func NewDB() *DB {
	return &DB{
		mu: &sync.Mutex{},

		minData: minData{
			accountsMin: []models.AccountMin{},
			fnames:      map[uint8]string{},
			snames:      map[uint16]string{},
			sex:         map[byte]string{0: "m", 1: "f"},
			countries:   map[uint8]string{},
			cities:      map[uint16]string{},
			status:      map[byte]string{0: "свободны", 1: "заняты", 2: "всё сложно"},
			interests:   map[uint8]string{},
		},
	}
}

func (db *DB) LoadMinData2(accounts []models.Account, now int) {
	jobs := make(chan models.Account, len(accounts))
	numOfWorkers := 100
	for numOfWorkers >= 0 {
		go db.accountWorker(jobs)
		numOfWorkers--
	}
	for i := range accounts {
		jobs <- accounts[i]
	}
	close(jobs)
}

func (db *DB) SortDB() {
	sort.Slice(db.accountsMin, func(i, j int) bool {
		return db.accountsMin[i].ID > db.accountsMin[j].ID
	})
}

func (db *DB) accountWorker(jobs <-chan models.Account) {
	for j := range jobs {
		db.AddAccount(j)
	}
}

func (db *DB) AddAccount(account models.Account) {
	db.mu.Lock()
	accountMin := models.AccountMin{
		ID:      account.ID,
		Email:   account.Email,
		Phone:   account.Phone,
		Birth:   account.Birth,
		Joined:  account.Joined,
		Premium: account.Premium,
		Likes:   account.Likes,
	}

	accountMin.FName = uint8(len(db.fnames))
	for k, v := range db.fnames {
		if v == account.FName {
			accountMin.FName = k
			break
		}
	}
	if accountMin.FName == uint8(len(db.fnames)) {
		db.fnames[accountMin.FName] = account.FName
	}

	accountMin.SName = uint16(len(db.snames))
	for k, v := range db.snames {
		if v == account.SName {
			accountMin.SName = k
			break
		}
	}
	if accountMin.SName == uint16(len(db.snames)) {
		db.snames[accountMin.SName] = account.SName
	}

	if account.Sex == "f" {
		accountMin.Sex = 1
	}

	accountMin.Country = uint8(len(db.countries))
	for k, v := range db.countries {
		if v == account.Country {
			accountMin.Country = k
			break
		}
	}
	if accountMin.Country == uint8(len(db.countries)) {
		db.countries[accountMin.Country] = account.Country
	}

	accountMin.City = uint16(len(db.cities))
	for k, v := range db.cities {
		if v == account.City {
			accountMin.City = k
			break
		}
	}
	if accountMin.City == uint16(len(db.cities)) {
		db.cities[accountMin.City] = account.City
	}

	switch account.Status {
	case "заняты":
		accountMin.Status = 1
	case "всё сложно":
		accountMin.Status = 2
	}

	var interests []uint8
	for _, interest := range account.Interests {
		interestId := uint8(len(db.interests))
		for k, v := range db.interests {
			if v == interest {
				interestId = k
				break
			}
		}
		if interestId == uint8(len(db.interests)) {
			db.interests[interestId] = interest
		}
		interests = append(interests, interestId)
	}
	accountMin.Interests = interests

	db.accountsMin = append(db.accountsMin, accountMin)

	db.mu.Unlock()
}

func (db *DB) Find(query M) models.Accounts {
	var res models.Accounts
	projection := make(map[string]void)
MainLoop:
	for i := range db.accountsMin {
		for k, v := range query {
			switch k {
			case "sex_eq":
				if db.sex[db.accountsMin[i].Sex] != v.(string) {
					continue MainLoop
				}
				projection["sex"] = void{}
			case "status_eq":
				if db.status[db.accountsMin[i].Status] != v.(string) {
					continue MainLoop
				}
				projection["status"] = void{}
			case "status_neq": //todo
				res = append(res, db.statusNeqIdx[v.(string)])
				projection["status"] = void{}
			case "fname_eq":
				res = append(res, db.fnameIdx[v.(string)])
				projection["fname"] = void{}
			case "fname_any":
				r := make(map[int]void)
				for _, fname := range v.([]string) {
					for kr, vr := range db.fnameIdx[fname] {
						r[kr] = vr
					}
				}
				res = append(res, r)
				projection["fname"] = void{}
			case "fname_null":
				switch v.(string) {
				case "0":
					res = append(res, db.fnameNotNullIdx)
				case "1":
					res = append(res, db.fnameIdx[""])
				}
				projection["fname"] = void{}
			case "sname_eq":
				res = append(res, db.snameIdx[v.(string)])
				projection["sname"] = void{}
			case "sname_null":
				switch v.(string) {
				case "0":
					res = append(res, db.snameNotNullIdx)
				case "1":
					res = append(res, db.snameIdx[""])
				}
				projection["sname"] = void{}
			case "phone_null":
				switch v.(string) {
				case "0":
					r := make(map[int]void) //todo make phone null idx
					for kp, vp := range db.phoneCodeIdx {
						if kp != "" {
							for kr, vr := range vp {
								r[kr] = vr
							}
						}
					}
					res = append(res, r)
				case "1":
					res = append(res, db.phoneCodeIdx[""])
				}
				projection["phone"] = void{}
			case "country_eq":
				res = append(res, db.countryIdx[v.(string)])
				projection["country"] = void{}
			case "country_null":
				switch v.(string) {
				case "0":
					res = append(res, db.countryNotNullIdx)
				case "1":
					res = append(res, db.countryIdx[""])
				}
				projection["country"] = void{}
			case "city_eq":
				res = append(res, db.cityIdx[v.(string)])
				projection["city"] = void{}
			case "city_any":
				r := make(map[int]void)
				for _, city := range v.([]string) {
					for kr, vr := range db.cityIdx[city] {
						r[kr] = vr
					}
				}
				res = append(res, r)
				projection["city"] = void{}
			case "city_null":
				switch v.(string) {
				case "0":
					res = append(res, db.cityNotNullIdx)
				case "1":
					res = append(res, db.cityIdx[""])
				}
				projection["city"] = void{}
			case "email_domain":
				res = append(res, db.emailDomainIdx[v.(string)])
			case "email_lt":
				x := db.getEmailLtIdxEntries(v.(string))
				ids := make(map[int]void, len(x))
				for _, e := range x {
					ids[e.id] = void{}
				}
				res = append(res, ids)
			case "email_gt":
				x := db.getEmailGtIdxEntries(v.(string))
				ids := make(map[int]void, len(x))
				for _, e := range x {
					ids[e.id] = void{}
				}
				res = append(res, ids)
			case "sname_starts":
				res = append(res, db.getSnamePrefixIds(v.(string)))
				projection["sname"] = void{}
			case "phone_code":
				res = append(res, db.phoneCodeIdx[v.(string)])
				projection["phone"] = void{}
			case "birth_lt":
				x := db.getBirthLtIdxEntries(v.(int))
				ids := make(map[int]void, len(x))
				for _, e := range x {
					ids[e.id] = void{}
				}
				res = append(res, ids)
				projection["birth"] = void{}
			case "birth_gt":
				x := db.getBirthGtIdxEntries(v.(int))
				ids := make(map[int]void, len(x))
				for _, e := range x {
					ids[e.id] = void{}
				}
				res = append(res, ids)
				projection["birth"] = void{}
			case "birth_year":
				b, ok := db.birthYearIdx[v.(int)]
				if !ok {
					b = make(map[int]void)
				}
				res = append(res, b)
				projection["birth"] = void{}
			case "interests_contains":
				r := make([]map[int]void, 0)
				for _, interest := range v.([]string) {
					r = append(r, db.interestsIdx[interest])
				}

				if len(r) == 0 {
					res = append(res, make(map[int]void))
					break
				}

				if len(r) == 1 {
					res = append(res, r[0])
					break
				}

				sort.Slice(r, func(i, j int) bool {
					return len(r[i]) < len(r[j])
				})

				ids := make(map[int]void)
			InterestsContainsLoop:
				for id := range r[0] {
					for i := 1; i < len(r); i++ {
						if _, ok := r[i][id]; !ok {
							continue InterestsContainsLoop
						}
					}
					ids[id] = void{}
				}
				res = append(res, ids)
			case "interests_any":
				ids := make(map[int]void)
				for _, interest := range v.([]string) {
					for ki, vi := range db.interestsIdx[interest] {
						ids[ki] = vi
					}
				}
				res = append(res, ids)
			case "likes_contains":
				r := make([]map[int]void, 0)
				for _, like := range v.([]int) {
					r = append(r, db.likesIdx[like])
				}

				if len(r) == 0 {
					res = append(res, make(map[int]void))
					break
				}

				if len(r) == 1 {
					res = append(res, r[0])
					break
				}

				sort.Slice(r, func(i, j int) bool {
					return len(r[i]) < len(r[j])
				})

				ids := make(map[int]void)
			LikesContainsLoop:
				for id := range r[0] {
					for i := 1; i < len(r); i++ {
						if _, ok := r[i][id]; !ok {
							continue LikesContainsLoop
						}
					}
					ids[id] = void{}
				}
				res = append(res, ids)
			case "premium_now":
				res = append(res, db.premiumIdx[2])
				projection["premium"] = void{}
			case "premium_null":
				switch v.(string) {
				case "0":
					res = append(res, db.premiumIdx[1])
				case "1":
					res = append(res, db.premiumIdx[0])
				}
				projection["premium"] = void{}
			}
		}
	}

	limit := query["limit"].(int)
	ids := make([]int, 0)
	accountsMin := make([]models.AccountMin, 0)

	if len(res) == 0 {
		for i := 0; i < limit; i++ {
			ids = append(ids, db.ids[i])
			accountsMin = append(accountsMin, db.accountsMin[db.ids[i]])
		}
	} else if len(res) == 1 {
		for id := range res[0] {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool {
			return ids[i] > ids[j]
		})

		if len(ids) > limit {
			ids = ids[:limit]
		}
		for _, id := range ids {
			accountsMin = append(accountsMin, db.accountsMin[id])
		}
	} else {
		sort.Slice(res, func(i, j int) bool {
			return len(res[i]) < len(res[j])
		})

	MinResLoop:
		for k := range res[0] {
			for i := 1; i < len(res); i++ {
				if _, ok := res[i][k]; !ok {
					continue MinResLoop
				}
			}

			ids = append(ids, k)
		}

		sort.Slice(ids, func(i, j int) bool {
			return ids[i] > ids[j]
		})

		if len(ids) > limit {
			ids = ids[:limit]
		}
		for _, id := range ids {
			accountsMin = append(accountsMin, db.accountsMin[id])
		}
	}

	accounts := models.Accounts{}
	accounts.Accounts = make([]models.Account, 0)
	for i := range accountsMin {
		account := models.Account{ID: ids[i], Email: accountsMin[i].Email}

		if _, ok := projection["fname"]; ok {
			account.FName = db.fnames[accountsMin[i].FName]
		}

		if _, ok := projection["sname"]; ok {
			account.SName = db.snames[accountsMin[i].SName]
		}

		if _, ok := projection["phone"]; ok {
			account.Phone = accountsMin[i].Phone
		}

		if _, ok := projection["sex"]; ok {
			account.Sex = db.sex[accountsMin[i].Sex]
		}

		if _, ok := projection["birth"]; ok {
			account.Birth = accountsMin[i].Birth
		}

		if _, ok := projection["country"]; ok {
			account.Country = db.countries[accountsMin[i].Country]
		}

		if _, ok := projection["city"]; ok {
			account.City = db.cities[accountsMin[i].City]
		}

		if _, ok := projection["status"]; ok {
			account.Status = db.status[accountsMin[i].Status]
		}

		if _, ok := projection["premium"]; ok {
			account.Premium = accountsMin[i].Premium
		}

		accounts.Accounts = append(accounts.Accounts, account)
	}

	return accounts
}
