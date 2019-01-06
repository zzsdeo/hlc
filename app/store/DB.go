package store

import (
	"hlc/app/models"
	"sort"
	"strings"
	"sync"
	"time"
)

type DB struct {
	mu         sync.RWMutex
	accounts   map[int]models.Account
	sexIdx     map[string]map[int]bool
	statusIdx  map[string]map[int]bool
	fnameIdx   map[string]map[int]bool
	snameIdx   map[string]map[int]bool
	phoneIdx   map[string]map[int]bool
	countryIdx map[string]map[int]bool
	cityIdx    map[string]map[int]bool
}

type M map[string]interface{}

func NewDB() *DB {
	return &DB{
		accounts:   make(map[int]models.Account),
		sexIdx:     make(map[string]map[int]bool),
		statusIdx:  make(map[string]map[int]bool),
		fnameIdx:   make(map[string]map[int]bool),
		snameIdx:   make(map[string]map[int]bool),
		phoneIdx:   make(map[string]map[int]bool),
		countryIdx: make(map[string]map[int]bool),
		cityIdx:    make(map[string]map[int]bool),
	}
}

func (db *DB) LoadData(accounts []models.Account) bool {
	db.mu.RLock()
	for _, account := range accounts {
		db.accounts[account.ID] = account
	}
	db.mu.RUnlock()
	return true
}

func (db *DB) Count() int {
	return len(db.accounts)
}

func (db *DB) CreateIndexes() bool {
	db.mu.RLock()
	for k, v := range db.accounts {
		if v.Sex != "" {
			if _, ok := db.sexIdx[v.Sex]; !ok {
				db.sexIdx[v.Sex] = make(map[int]bool)
			}
			db.sexIdx[v.Sex][k] = true
		}

		if v.Status != "" {
			if _, ok := db.statusIdx[v.Status]; !ok {
				db.statusIdx[v.Status] = make(map[int]bool)
			}
			db.statusIdx[v.Status][k] = true
		}

		if _, ok := db.fnameIdx[v.FName]; !ok {
			db.fnameIdx[v.FName] = make(map[int]bool)
		}
		db.fnameIdx[v.FName][k] = true

		if _, ok := db.snameIdx[v.SName]; !ok {
			db.snameIdx[v.SName] = make(map[int]bool)
		}
		db.snameIdx[v.SName][k] = true

		if _, ok := db.phoneIdx[v.Phone]; !ok {
			db.phoneIdx[v.Phone] = make(map[int]bool)
		}
		db.phoneIdx[v.Phone][k] = true

		if _, ok := db.countryIdx[v.Country]; !ok {
			db.countryIdx[v.Country] = make(map[int]bool)
		}
		db.countryIdx[v.Country][k] = true

		if _, ok := db.cityIdx[v.City]; !ok {
			db.cityIdx[v.City] = make(map[int]bool)
		}
		db.cityIdx[v.City][k] = true
	}
	db.mu.RUnlock()
	return true
}

func (db *DB) Find(query M) models.Accounts {
	//log.Println("[DEBUG] query", query)
	accounts := models.Accounts{}
	res := make([]map[int]bool, 0)
	projection := make(map[string]bool)
	for k, v := range query {
		switch k {
		case "sex_eq":
			res = append(res, db.sexIdx[v.(string)])
			projection["sex"] = true
			delete(query, k)
		case "status_eq":
			res = append(res, db.statusIdx[v.(string)])
			projection["status"] = true
			delete(query, k)
		case "status_neq":
			r := make(map[int]bool)
			for ks, vs := range db.statusIdx {
				if ks != v.(string) {
					for kr, vr := range vs {
						r[kr] = vr
					}
				}
			}
			res = append(res, r)
			projection["status"] = true
			delete(query, k)
		case "fname_eq":
			res = append(res, db.fnameIdx[v.(string)])
			projection["fname"] = true
			delete(query, k)
		case "fname_any":
			r := make(map[int]bool)
			for _, fname := range v.([]string) {
				for kr, vr := range db.fnameIdx[fname] {
					r[kr] = vr
				}
			}
			res = append(res, r)
			projection["fname"] = true
			delete(query, k)
		case "fname_null":
			switch v.(string) {
			case "0":
				r := make(map[int]bool)
				for kf, vf := range db.fnameIdx {
					if kf != "" {
						for kr, vr := range vf {
							r[kr] = vr
						}
					}
				}
				res = append(res, r)
			case "1":
				res = append(res, db.fnameIdx[""])
			}
			projection["fname"] = true
			delete(query, k)
		case "sname_eq":
			res = append(res, db.snameIdx[v.(string)])
			projection["sname"] = true
			delete(query, k)
		case "sname_null":
			switch v.(string) {
			case "0":
				r := make(map[int]bool)
				for ks, vs := range db.snameIdx {
					if ks != "" {
						for kr, vr := range vs {
							r[kr] = vr
						}
					}
				}
				res = append(res, r)
			case "1":
				res = append(res, db.snameIdx[""])
			}
			projection["sname"] = true
			delete(query, k)
		case "phone_null":
			switch v.(string) {
			case "0":
				r := make(map[int]bool)
				for kp, vp := range db.phoneIdx {
					if kp != "" {
						for kr, vr := range vp {
							r[kr] = vr
						}
					}
				}
				res = append(res, r)
			case "1":
				res = append(res, db.phoneIdx[""])
			}
			projection["phone"] = true
			delete(query, k)
		case "country_eq":
			res = append(res, db.countryIdx[v.(string)])
			projection["country"] = true
			delete(query, k)
		case "country_null":
			switch v.(string) {
			case "0":
				r := make(map[int]bool)
				for kc, vc := range db.countryIdx {
					if kc != "" {
						for kr, vr := range vc {
							r[kr] = vr
						}
					}
				}
				res = append(res, r)
			case "1":
				res = append(res, db.countryIdx[""])
			}
			projection["country"] = true
			delete(query, k)
		case "city_eq":
			res = append(res, db.cityIdx[v.(string)])
			projection["city"] = true
			delete(query, k)
		case "city_any":
			r := make(map[int]bool)
			for _, city := range v.([]string) {
				for kr, vr := range db.cityIdx[city] {
					r[kr] = vr
				}
			}
			res = append(res, r)
			projection["city"] = true
			delete(query, k)
		case "city_null":
			switch v.(string) {
			case "0":
				r := make(map[int]bool)
				for kc, vc := range db.cityIdx {
					if kc != "" {
						for kr, vr := range vc {
							r[kr] = vr
						}
					}
				}
				res = append(res, r)
			case "1":
				res = append(res, db.cityIdx[""])
			}
			projection["city"] = true
			delete(query, k)
		}
	}

	if len(res) > 0 {
		if len(res) > 1 {
			ids := make(map[int]bool)
			sort.Slice(res, func(i, j int) bool {
				return len(res[i]) < len(res[j])
			})

		MinResLoop:
			for k, _ := range res[0] {
				for i := 1; i < len(res); i++ {
					if _, ok := res[i][k]; !ok {
						continue MinResLoop
					}
				}
				ids[k] = true
			}

			for id, _ := range ids {
				accounts.Accounts = append(accounts.Accounts, db.accounts[id])
			}
		} else {
			for id, _ := range res[0] {
				accounts.Accounts = append(accounts.Accounts, db.accounts[id])
			}
		}
	} else {
		for _, account := range db.accounts {
			accounts.Accounts = append(accounts.Accounts, account)
		}
	}

	result := models.Accounts{}
	result.Accounts = make([]models.Account, 0)
MainLoop:
	for _, account := range accounts.Accounts {
		for k, v := range query {
			switch k {
			case "email_domain":
				if !strings.Contains(account.Email, "@"+v.(string)) {
					continue MainLoop
				}
			case "email_lt":
				if account.Email > v.(string) {
					continue MainLoop
				}
			case "email_gt":
				if account.Email < v.(string) {
					continue MainLoop
				}
			case "sname_starts":
				projection["sname"] = true
				if !strings.HasPrefix(account.SName, v.(string)) {
					continue MainLoop
				}
			case "phone_code":
				projection["phone"] = true
				if !strings.Contains(account.Phone, "("+v.(string)+")") {
					continue MainLoop
				}
			case "birth_lt":
				projection["birth"] = true
				if account.Birth >= v.(int) {
					continue MainLoop
				}
			case "birth_gt":
				projection["birth"] = true
				if account.Birth <= v.(int) {
					continue MainLoop
				}
			case "birth_year":
				projection["birth"] = true
				s := time.Date(v.(int), time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
				f := time.Date(v.(int)+1, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
				if !(account.Birth > int(s) && account.Birth < int(f)) {
					continue MainLoop
				}
			case "interests_contains":
				account.PrepareInterestsMap()
				if !account.InterestsContains(v.([]string)) {
					continue MainLoop
				}
			case "interests_any":
				account.PrepareInterestsMap()
				if !account.InterestsAny(v.([]string)) {
					continue MainLoop
				}
			case "likes_contains":
				account.PrepareLikesMap()
				if !account.LikesContains(v.([]int)) {
					continue MainLoop
				}
			case "premium_now":
				projection["premium"] = true
				if !account.PremiumNow(v.(int)) {
					continue MainLoop
				}
			case "premium_null":
				projection["premium"] = true
				switch v.(string) {
				case "0":
					if account.Premium == nil {
						continue MainLoop
					}
				case "1":
					if account.Premium != nil {
						continue MainLoop
					}
				}
			}
		}

		account.Interests = []string{}
		account.Likes = []models.Like{}
		account.Joined = 0

		if ok, _ := projection["fname"]; !ok {
			account.FName = ""
		}

		if ok, _ := projection["sname"]; !ok {
			account.SName = ""
		}

		if ok, _ := projection["phone"]; !ok {
			account.Phone = ""
		}

		if ok, _ := projection["sex"]; !ok {
			account.Sex = ""
		}

		if ok, _ := projection["birth"]; !ok {
			account.Birth = 0
		}

		if ok, _ := projection["country"]; !ok {
			account.Country = ""
		}

		if ok, _ := projection["city"]; !ok {
			account.City = ""
		}

		if ok, _ := projection["status"]; !ok {
			account.Status = ""
		}

		if ok, _ := projection["premium"]; !ok {
			account.Premium = nil
		}

		result.Accounts = append(result.Accounts, account)
	}

	sort.Slice(result.Accounts, func(i, j int) bool {
		return result.Accounts[i].ID > result.Accounts[j].ID
	})

	limit := query["limit"].(int)
	if limit < len(result.Accounts) {
		result.Accounts = result.Accounts[:limit]
	}

	return result
}
