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
	accountsMin map[int]models.AccountMin
	fnames      map[int]string
	snames      map[int]string
	sex         map[byte]string
	countries   map[int]string
	cities      map[int]string
	status      map[byte]string
	interests   map[int]string
}

type DB struct {
	minData
	mu                sync.RWMutex
	accounts          map[int]models.Account
	ids               []int
	sexIdx            map[string]map[int]bool
	statusIdx         map[string]map[int]bool
	statusNeqIdx      map[string]map[int]bool
	fnameIdx          map[string]map[int]bool
	fnameNotNullIdx   map[int]bool
	snameIdx          map[string]map[int]bool
	snameNotNullIdx   map[int]bool
	phoneIdx          map[string]map[int]bool
	countryIdx        map[string]map[int]bool
	countryNotNullIdx map[int]bool
	cityIdx           map[string]map[int]bool
	cityNotNullIdx    map[int]bool
	emailIdx          []emailIdxEntry
	emailDomainIdx    map[string]map[int]bool
	snamePrefixIdx    trieNode
	birthIdx          []birthIdxEntry
	birthYearIdx      map[int]map[int]bool
	interestsIdx      map[string]map[int]bool
	likesIdx          map[int]map[int]bool
	premiumIdx        map[string]map[int]bool
}

type M map[string]interface{}

type emailIdxEntry struct {
	email string
	id    int
}

type birthIdxEntry struct {
	birth int
	id    int
}

type trieNode struct {
	next map[int32]trieNode
	ids  map[int]bool
}

func NewDB() *DB {
	return &DB{
		accounts:          make(map[int]models.Account),
		ids:               make([]int, 0),
		sexIdx:            make(map[string]map[int]bool),
		statusIdx:         make(map[string]map[int]bool),
		statusNeqIdx:      map[string]map[int]bool{"заняты": {}, "свободны": {}, "всё сложно": {}},
		fnameIdx:          make(map[string]map[int]bool),
		fnameNotNullIdx:   make(map[int]bool),
		snameIdx:          make(map[string]map[int]bool),
		snameNotNullIdx:   make(map[int]bool),
		phoneIdx:          make(map[string]map[int]bool),
		countryIdx:        make(map[string]map[int]bool),
		countryNotNullIdx: make(map[int]bool),
		cityIdx:           make(map[string]map[int]bool),
		cityNotNullIdx:    make(map[int]bool),
		emailIdx:          make([]emailIdxEntry, 0),
		emailDomainIdx:    make(map[string]map[int]bool),
		snamePrefixIdx:    trieNode{next: make(map[int32]trieNode), ids: make(map[int]bool)},
		birthIdx:          make([]birthIdxEntry, 0),
		birthYearIdx:      make(map[int]map[int]bool),
		interestsIdx:      make(map[string]map[int]bool),
		likesIdx:          make(map[int]map[int]bool),
		premiumIdx:        make(map[string]map[int]bool),

		minData: minData{
			accountsMin: map[int]models.AccountMin{},
			fnames:      map[int]string{},
			snames:      map[int]string{},
			sex:         map[byte]string{0: "m", 1: "f"},
			countries:   map[int]string{},
			cities:      map[int]string{},
			status:      map[byte]string{0: "свободны", 1: "заняты", 2: "всё сложно"},
			interests:   map[int]string{},
		},
	}
}

func (db *DB) getEmailLtIdxEntries(prefix string) []emailIdxEntry {
	low := 0
	high := len(db.emailIdx) - 1

	for low <= high {
		mid := (low + high) / 2
		guess := db.emailIdx[mid]
		if guess.email <= prefix && mid+1 < len(db.emailIdx) && db.emailIdx[mid+1].email > prefix {
			return db.emailIdx[:mid+1]
		}

		if guess.email > prefix {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}

	return []emailIdxEntry{}
}

func (db *DB) getEmailGtIdxEntries(prefix string) []emailIdxEntry {
	low := 0
	high := len(db.emailIdx) - 1

	for low <= high {
		mid := (low + high) / 2
		guess := db.emailIdx[mid]
		if guess.email >= prefix && mid-1 >= 0 && db.emailIdx[mid-1].email < prefix {
			return db.emailIdx[mid:]
		}

		if guess.email > prefix {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}

	return []emailIdxEntry{}
}

func (db *DB) getBirthLtIdxEntries(birth int) []birthIdxEntry {
	low := 0
	high := len(db.birthIdx) - 1

	for low <= high {
		mid := (low + high) / 2
		guess := db.birthIdx[mid]
		if guess.birth <= birth && mid+1 < len(db.birthIdx) && db.birthIdx[mid+1].birth > birth {
			return db.birthIdx[:mid+1]
		}

		if guess.birth > birth {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}

	return []birthIdxEntry{}
}

func (db *DB) getBirthGtIdxEntries(birth int) []birthIdxEntry {
	low := 0
	high := len(db.birthIdx) - 1

	for low <= high {
		mid := (low + high) / 2
		guess := db.birthIdx[mid]
		if guess.birth >= birth && mid-1 >= 0 && db.birthIdx[mid-1].birth < birth {
			return db.birthIdx[mid:]
		}

		if guess.birth > birth {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}

	return []birthIdxEntry{}
}

func (db *DB) getSnamePrefixIds(prefix string) map[int]bool {
	start := true
	var currentNode trieNode
	for _, char := range prefix {
		if start {
			if _, ok := db.snamePrefixIdx.next[char]; !ok {
				return make(map[int]bool)
			}
			currentNode = db.snamePrefixIdx.next[char]
			start = false
		} else {
			if _, ok := currentNode.next[char]; !ok {
				return make(map[int]bool)
			}
			currentNode = currentNode.next[char]
		}
	}
	return currentNode.ids
}

func (db *DB) LoadData(accounts []models.Account) {
	db.mu.RLock()
	for _, account := range accounts {
		db.accounts[account.ID] = account
	}
	db.mu.RUnlock()
	runtime.GC()
}

func (db *DB) LoadMinData(accounts []models.Account) {
	db.mu.RLock()
	for _, account := range accounts {
		fnameId := len(db.fnames)
		for k, v := range db.fnames {
			if v == account.FName {
				fnameId = k
				break
			}
		}

		if fnameId == len(db.fnames) {
			db.fnames[fnameId] = account.FName
		}

		snameId := len(db.snames)
		for k, v := range db.snames {
			if v == account.SName {
				snameId = k
				break
			}
		}

		if snameId == len(db.snames) {
			db.snames[snameId] = account.SName
		}

		var sex byte = 0
		if account.Sex == "f" {
			sex = 1
		}

		countryId := len(db.countries)
		for k, v := range db.countries {
			if v == account.Country {
				countryId = k
				break
			}
		}

		if countryId == len(db.countries) {
			db.countries[countryId] = account.Country
		}

		cityId := len(db.cities)
		for k, v := range db.cities {
			if v == account.City {
				cityId = k
				break
			}
		}

		if cityId == len(db.cities) {
			db.cities[cityId] = account.City
		}

		var status byte = 0
		switch account.Status {
		case "заняты":
			status = 1
		case "всё сложно":
			status = 2
		}

		var interests []int
		for _, interest := range account.Interests {
			interestId := len(db.interests)
			for k, v := range db.interests {
				if v == interest {
					interestId = k
					break
				}
			}

			if interestId == len(db.interests) {
				db.interests[interestId] = interest
			}
			interests = append(interests, interestId)
		}

		accountMin := models.AccountMin{
			Email:     account.Email,
			FName:     fnameId,
			SName:     snameId,
			Phone:     account.Phone,
			Sex:       sex,
			Birth:     account.Birth,
			Country:   countryId,
			City:      cityId,
			Joined:    account.Joined,
			Status:    status,
			Interests: interests,
			Premium:   account.Premium,
			Likes:     account.Likes,
		}
		db.accountsMin[account.ID] = accountMin
	}
	db.mu.RUnlock()
	runtime.GC()
}

func (db *DB) CreateIndexes(now int) bool {
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

			switch v.Status {
			case "заняты":
				db.statusNeqIdx["свободны"][k] = true
				db.statusNeqIdx["всё сложно"][k] = true
			case "свободны":
				db.statusNeqIdx["заняты"][k] = true
				db.statusNeqIdx["всё сложно"][k] = true
			case "всё сложно":
				db.statusNeqIdx["заняты"][k] = true
				db.statusNeqIdx["свободны"][k] = true
			}
		}

		if _, ok := db.fnameIdx[v.FName]; !ok {
			db.fnameIdx[v.FName] = make(map[int]bool)
		}
		db.fnameIdx[v.FName][k] = true

		if v.FName != "" {
			db.fnameNotNullIdx[k] = true
		}

		if _, ok := db.snameIdx[v.SName]; !ok {
			db.snameIdx[v.SName] = make(map[int]bool)
		}
		db.snameIdx[v.SName][k] = true

		if v.SName != "" {
			db.snameNotNullIdx[k] = true
		}

		phoneCode := ""
		if v.Phone != "" {
			s := strings.Split(v.Phone, "(")
			s = strings.Split(s[1], ")")
			phoneCode = s[0]
		}
		if _, ok := db.phoneIdx[phoneCode]; !ok {
			db.phoneIdx[phoneCode] = make(map[int]bool)
		}
		db.phoneIdx[phoneCode][k] = true

		if _, ok := db.countryIdx[v.Country]; !ok {
			db.countryIdx[v.Country] = make(map[int]bool)
		}
		db.countryIdx[v.Country][k] = true

		if v.Country != "" {
			db.countryNotNullIdx[k] = true
		}

		if _, ok := db.cityIdx[v.City]; !ok {
			db.cityIdx[v.City] = make(map[int]bool)
		}
		db.cityIdx[v.City][k] = true

		if v.City != "" {
			db.cityNotNullIdx[k] = true
		}

		db.emailIdx = append(db.emailIdx, emailIdxEntry{v.Email, k})

		domain := strings.Split(v.Email, "@")[1]
		if _, ok := db.emailDomainIdx[domain]; !ok {
			db.emailDomainIdx[domain] = make(map[int]bool)
		}
		db.emailDomainIdx[domain][k] = true

		if v.SName != "" {
			start := true
			var currentNode trieNode
			for _, char := range v.SName {
				if start {
					if _, ok := db.snamePrefixIdx.next[char]; !ok {
						db.snamePrefixIdx.next[char] = trieNode{next: make(map[int32]trieNode), ids: make(map[int]bool)}
					}
					currentNode = db.snamePrefixIdx.next[char]
					currentNode.ids[k] = true
					start = false
				} else {
					if _, ok := currentNode.next[char]; !ok {
						currentNode.next[char] = trieNode{next: make(map[int32]trieNode), ids: make(map[int]bool)}
					}
					currentNode = currentNode.next[char]
					currentNode.ids[k] = true
				}
			}
		}

		db.birthIdx = append(db.birthIdx, birthIdxEntry{v.Birth, k})

		year := time.Unix(int64(v.Birth), 0).Year()
		if _, ok := db.birthYearIdx[year]; !ok {
			db.birthYearIdx[year] = make(map[int]bool)
		}
		db.birthYearIdx[year][k] = true

		for _, interest := range v.Interests {
			if _, ok := db.interestsIdx[interest]; !ok {
				db.interestsIdx[interest] = make(map[int]bool)
			}
			db.interestsIdx[interest][k] = true
		}

		for _, like := range v.Likes {
			if _, ok := db.likesIdx[like.ID]; !ok {
				db.likesIdx[like.ID] = make(map[int]bool)
			}
			db.likesIdx[like.ID][k] = true
		}

		if v.Premium == nil {
			if _, ok := db.premiumIdx["null"]; !ok {
				db.premiumIdx["null"] = make(map[int]bool)
			}
			db.premiumIdx["null"][k] = true
		} else {
			if _, ok := db.premiumIdx["not_null"]; !ok {
				db.premiumIdx["not_null"] = make(map[int]bool)
			}
			db.premiumIdx["not_null"][k] = true
			if v.PremiumNow(now) {
				if _, ok := db.premiumIdx["now"]; !ok {
					db.premiumIdx["now"] = make(map[int]bool)
				}
				db.premiumIdx["now"][k] = true
			}
		}

		db.ids = append(db.ids, k)
	}
	sort.Slice(db.emailIdx, func(i, j int) bool {
		return db.emailIdx[i].email < db.emailIdx[j].email
	})
	sort.Slice(db.birthIdx, func(i, j int) bool {
		return db.birthIdx[i].birth < db.birthIdx[j].birth
	})
	sort.Slice(db.ids, func(i, j int) bool {
		return db.ids[i] > db.ids[j]
	})
	db.mu.RUnlock()
	runtime.GC()

	log.Println("indexes size", utils.Sizeof(
		db.sexIdx,
		db.statusIdx,
		db.fnameIdx,
		db.snameIdx,
		db.phoneIdx,
		db.countryIdx,
		db.cityIdx,
		db.emailIdx,
		db.emailDomainIdx,
		db.snamePrefixIdx,
		db.birthIdx,
		db.birthYearIdx,
		db.interestsIdx,
		db.likesIdx,
		db.premiumIdx))

	log.Println("db size", utils.Sizeof(db.accounts))
	return true
}

func (db *DB) Find(query M) models.Accounts {
	//log.Println("[DEBUG] query", query)
	res := make([]map[int]bool, 0)
	projection := make(map[string]bool)
	for k, v := range query {
		switch k {
		case "sex_eq":
			res = append(res, db.sexIdx[v.(string)])
			projection["sex"] = true
		case "status_eq":
			res = append(res, db.statusIdx[v.(string)])
			projection["status"] = true
		case "status_neq":
			res = append(res, db.statusNeqIdx[v.(string)])
			projection["status"] = true
		case "fname_eq":
			res = append(res, db.fnameIdx[v.(string)])
			projection["fname"] = true
		case "fname_any":
			r := make(map[int]bool)
			for _, fname := range v.([]string) {
				for kr, vr := range db.fnameIdx[fname] {
					r[kr] = vr
				}
			}
			res = append(res, r)
			projection["fname"] = true
		case "fname_null":
			switch v.(string) {
			case "0":
				res = append(res, db.fnameNotNullIdx)
			case "1":
				res = append(res, db.fnameIdx[""])
			}
			projection["fname"] = true
		case "sname_eq":
			res = append(res, db.snameIdx[v.(string)])
			projection["sname"] = true
		case "sname_null":
			switch v.(string) {
			case "0":
				res = append(res, db.snameNotNullIdx)
			case "1":
				res = append(res, db.snameIdx[""])
			}
			projection["sname"] = true
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
		case "country_eq":
			res = append(res, db.countryIdx[v.(string)])
			projection["country"] = true
		case "country_null":
			switch v.(string) {
			case "0":
				res = append(res, db.countryNotNullIdx)
			case "1":
				res = append(res, db.countryIdx[""])
			}
			projection["country"] = true
		case "city_eq":
			res = append(res, db.cityIdx[v.(string)])
			projection["city"] = true
		case "city_any":
			r := make(map[int]bool)
			for _, city := range v.([]string) {
				for kr, vr := range db.cityIdx[city] {
					r[kr] = vr
				}
			}
			res = append(res, r)
			projection["city"] = true
		case "city_null":
			switch v.(string) {
			case "0":
				res = append(res, db.cityNotNullIdx)
			case "1":
				res = append(res, db.cityIdx[""])
			}
			projection["city"] = true
		case "email_domain":
			res = append(res, db.emailDomainIdx[v.(string)])
		case "email_lt":
			x := db.getEmailLtIdxEntries(v.(string))
			ids := make(map[int]bool)
			for _, e := range x {
				ids[e.id] = true
			}
			res = append(res, ids)
		case "email_gt":
			x := db.getEmailGtIdxEntries(v.(string))
			ids := make(map[int]bool)
			for _, e := range x {
				ids[e.id] = true
			}
			res = append(res, ids)
		case "sname_starts":
			res = append(res, db.getSnamePrefixIds(v.(string)))
			projection["sname"] = true
		case "phone_code":
			res = append(res, db.phoneIdx[v.(string)])
			projection["phone"] = true
		case "birth_lt":
			x := db.getBirthLtIdxEntries(v.(int))
			ids := make(map[int]bool)
			for _, e := range x {
				ids[e.id] = true
			}
			res = append(res, ids)
			projection["birth"] = true
		case "birth_gt":
			x := db.getBirthGtIdxEntries(v.(int))
			ids := make(map[int]bool)
			for _, e := range x {
				ids[e.id] = true
			}
			res = append(res, ids)
			projection["birth"] = true
		case "birth_year":
			b, ok := db.birthYearIdx[v.(int)]
			if !ok {
				b = make(map[int]bool)
			}
			res = append(res, b)
			projection["birth"] = true
		case "interests_contains":
			r := make([]map[int]bool, 0)
			for _, interest := range v.([]string) {
				r = append(r, db.interestsIdx[interest])
			}

			if len(r) == 0 {
				res = append(res, make(map[int]bool))
				break
			}

			if len(r) == 1 {
				res = append(res, r[0])
				break
			}

			sort.Slice(r, func(i, j int) bool {
				return len(r[i]) < len(r[j])
			})

			ids := make(map[int]bool)
		InterestsContainsLoop:
			for id := range r[0] {
				for i := 1; i < len(r); i++ {
					if _, ok := r[i][id]; !ok {
						continue InterestsContainsLoop
					}
				}
				ids[id] = true
			}
			res = append(res, ids)
		case "interests_any":
			ids := make(map[int]bool)
			for _, interest := range v.([]string) {
				for ki, vi := range db.interestsIdx[interest] {
					ids[ki] = vi
				}
			}
			res = append(res, ids)
		case "likes_contains":
			r := make([]map[int]bool, 0)
			for _, like := range v.([]int) {
				r = append(r, db.likesIdx[like])
			}

			if len(r) == 0 {
				res = append(res, make(map[int]bool))
				break
			}

			if len(r) == 1 {
				res = append(res, r[0])
				break
			}

			sort.Slice(r, func(i, j int) bool {
				return len(r[i]) < len(r[j])
			})

			ids := make(map[int]bool)
		LikesContainsLoop:
			for id := range r[0] {
				for i := 1; i < len(r); i++ {
					if _, ok := r[i][id]; !ok {
						continue LikesContainsLoop
					}
				}
				ids[id] = true
			}
			res = append(res, ids)
		case "premium_now":
			res = append(res, db.premiumIdx["now"])
			projection["premium"] = true
		case "premium_null":
			switch v.(string) {
			case "0":
				res = append(res, db.premiumIdx["not_null"])
			case "1":
				res = append(res, db.premiumIdx["null"])
			}
			projection["premium"] = true
		}
	}

	limit := query["limit"].(int)
	ids := make([]int, 0)
	accounts := models.Accounts{}
	accounts.Accounts = make([]models.Account, 0)

	if len(res) == 0 {
		for i := 0; i < limit; i++ {
			accounts.Accounts = append(accounts.Accounts, db.accounts[db.ids[i]])
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
			accounts.Accounts = append(accounts.Accounts, db.accounts[id])
		}
	} else {
		idsMap := make(map[int]bool)
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
			idsMap[k] = true
		}

		for id := range idsMap {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool {
			return ids[i] > ids[j]
		})

		if len(ids) > limit {
			ids = ids[:limit]
		}
		for _, id := range ids {
			accounts.Accounts = append(accounts.Accounts, db.accounts[id])
		}
	}

	for i := range accounts.Accounts {

		accounts.Accounts[i].Interests = []string{}
		accounts.Accounts[i].Likes = []models.Like{}
		accounts.Accounts[i].Joined = 0

		if _, ok := projection["fname"]; !ok {
			accounts.Accounts[i].FName = ""
		}

		if _, ok := projection["sname"]; !ok {
			accounts.Accounts[i].SName = ""
		}

		if _, ok := projection["phone"]; !ok {
			accounts.Accounts[i].Phone = ""
		}

		if _, ok := projection["sex"]; !ok {
			accounts.Accounts[i].Sex = ""
		}

		if _, ok := projection["birth"]; !ok {
			accounts.Accounts[i].Birth = 0
		}

		if _, ok := projection["country"]; !ok {
			accounts.Accounts[i].Country = ""
		}

		if _, ok := projection["city"]; !ok {
			accounts.Accounts[i].City = ""
		}

		if _, ok := projection["status"]; !ok {
			accounts.Accounts[i].Status = ""
		}

		if _, ok := projection["premium"]; !ok {
			accounts.Accounts[i].Premium = nil
		}
	}

	return accounts
}
