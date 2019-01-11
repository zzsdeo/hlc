package store

import (
	"hlc/app/models"
	"hlc/app/utils"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

type minData struct {
	accountsMin map[int]models.AccountMin
	fnames      map[int]string
	snames      map[int]string
	countries   map[int]string
	cities      map[int]string
	status      []string
	interests   map[int]string
}

type DB struct {
	minData
	mu                *sync.RWMutex
	ids               []int
	sexIdx            map[string][]int
	statusIdx         map[string][]int
	statusNeqIdx      map[string][]int
	fnameIdx          map[string][]int
	fnameNotNullIdx   []int
	snameIdx          map[string][]int
	snameNotNullIdx   []int
	phoneCodeIdx      map[string][]int
	countryIdx        map[string][]int
	countryNotNullIdx []int
	cityIdx           map[string][]int
	cityNotNullIdx    []int
	emailIdx          []emailIdxEntry
	emailDomainIdx    map[string][]int
	snamePrefixIdx    trieNode
	birthIdx          []birthIdxEntry
	birthYearIdx      map[int][]int
	interestsIdx      map[string][]int
	likesIdx          map[int][]int
	premiumIdx        map[byte][]int // 0-null, 1-not_null, 2-now
}

type void struct{}

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
	ids  *[]int
}

func NewDB() *DB {
	return &DB{
		mu: &sync.RWMutex{},

		//indexes
		ids:               make([]int, 0),
		sexIdx:            map[string][]int{},
		statusIdx:         map[string][]int{},
		statusNeqIdx:      map[string][]int{"свободны": {}, "заняты": {}, "всё сложно": {}},
		fnameIdx:          map[string][]int{},
		fnameNotNullIdx:   []int{},
		snameIdx:          map[string][]int{},
		snameNotNullIdx:   []int{},
		phoneCodeIdx:      map[string][]int{},
		countryIdx:        map[string][]int{},
		countryNotNullIdx: []int{},
		cityIdx:           map[string][]int{},
		cityNotNullIdx:    []int{},
		emailIdx:          []emailIdxEntry{},
		emailDomainIdx:    map[string][]int{},
		snamePrefixIdx:    trieNode{next: map[int32]trieNode{}, ids: &[]int{}},
		birthIdx:          []birthIdxEntry{},
		birthYearIdx:      map[int][]int{},
		interestsIdx:      map[string][]int{},
		likesIdx:          map[int][]int{},
		premiumIdx:        map[byte][]int{},

		minData: minData{
			accountsMin: map[int]models.AccountMin{},
			fnames:      map[int]string{},
			snames:      map[int]string{},
			countries:   map[int]string{},
			cities:      map[int]string{},
			status:      []string{"свободны", "заняты", "всё сложно"},
			interests:   map[int]string{},
		},
	}
}

func (db *DB) getEmailLtIdxEntries(prefix string) []int {
	var result []int
	low := 0
	high := len(db.emailIdx) - 1

	for low <= high {
		mid := (low + high) / 2
		guess := db.emailIdx[mid]
		if guess.email <= prefix && mid+1 < len(db.emailIdx) && db.emailIdx[mid+1].email > prefix {
			for _, e := range db.emailIdx[:mid+1] {
				result = append(result, e.id)
			}
			sort.Ints(result)
			return result
		}

		if guess.email > prefix {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}

	return result
}

func (db *DB) getEmailGtIdxEntries(prefix string) []int {
	var result []int
	low := 0
	high := len(db.emailIdx) - 1

	for low <= high {
		mid := (low + high) / 2
		guess := db.emailIdx[mid]
		if guess.email >= prefix && mid-1 >= 0 && db.emailIdx[mid-1].email < prefix {
			for _, e := range db.emailIdx[mid:] {
				result = append(result, e.id)
			}
			sort.Ints(result)
			return result
		}

		if guess.email > prefix {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}

	return result
}

func (db *DB) getBirthLtIdxEntries(birth int) []int {
	var result []int
	low := 0
	high := len(db.birthIdx) - 1

	for low <= high {
		mid := (low + high) / 2
		guess := db.birthIdx[mid]
		if guess.birth <= birth && mid+1 < len(db.birthIdx) && db.birthIdx[mid+1].birth > birth {
			for _, e := range db.birthIdx[:mid+1] {
				result = append(result, e.id)
			}
			sort.Ints(result)
			return result
		}

		if guess.birth > birth {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}

	return result
}

func (db *DB) getBirthGtIdxEntries(birth int) []int {
	var result []int
	low := 0
	high := len(db.birthIdx) - 1

	for low <= high {
		mid := (low + high) / 2
		guess := db.birthIdx[mid]
		if guess.birth >= birth && mid-1 >= 0 && db.birthIdx[mid-1].birth < birth {
			for _, e := range db.birthIdx[mid:] {
				result = append(result, e.id)
			}
			sort.Ints(result)
			return result
		}

		if guess.birth > birth {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return result
}

func (db *DB) getSnamePrefixIds(prefix string) []int {
	start := true
	var currentNode trieNode
	for _, char := range prefix {
		if start {
			if _, ok := db.snamePrefixIdx.next[char]; !ok {
				return []int{}
			}
			currentNode = db.snamePrefixIdx.next[char]
			start = false
		} else {
			if _, ok := currentNode.next[char]; !ok {
				return []int{}
			}
			currentNode = currentNode.next[char]
		}
	}
	sort.Ints(*currentNode.ids)
	return *currentNode.ids
}

func (db *DB) LoadMinData(accounts []models.Account) {
	db.mu.Lock()
	for _, account := range accounts {
		accountMin := models.AccountMin{
			Email:   account.Email,
			Phone:   account.Phone,
			Sex:     account.Sex,
			Birth:   account.Birth,
			Joined:  account.Joined,
			Premium: account.Premium,
			Likes:   account.Likes}

		accountMin.FName = len(db.fnames)
		for k, v := range db.fnames {
			if v == account.FName {
				accountMin.FName = k
				break
			}
		}
		if accountMin.FName == len(db.fnames) {
			db.fnames[accountMin.FName] = account.FName
		}

		accountMin.SName = len(db.snames)
		for k, v := range db.snames {
			if v == account.SName {
				accountMin.SName = k
				break
			}
		}
		if accountMin.SName == len(db.snames) {
			db.snames[accountMin.SName] = account.SName
		}

		accountMin.Country = len(db.countries)
		for k, v := range db.countries {
			if v == account.Country {
				accountMin.Country = k
				break
			}
		}
		if accountMin.Country == len(db.countries) {
			db.countries[accountMin.Country] = account.Country
		}

		accountMin.City = len(db.cities)
		for k, v := range db.cities {
			if v == account.City {
				accountMin.City = k
				break
			}
		}
		if accountMin.City == len(db.cities) {
			db.cities[accountMin.City] = account.City
		}

		switch account.Status {
		case "заняты":
			accountMin.Status = 1
		case "всё сложно":
			accountMin.Status = 2
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
		accountMin.Interests = interests

		db.accountsMin[account.ID] = accountMin
	}
	db.mu.Unlock()
	//runtime.GC()
}

func (db *DB) CreateIndexes(now int) bool {
	db.mu.Lock()
	for k, v := range db.accountsMin {

		db.sexIdx[v.Sex] = append(db.sexIdx[v.Sex], k)

		if _, ok := db.statusIdx[db.status[v.Status]]; !ok {
			db.statusIdx[db.status[v.Status]] = []int{}
		}
		db.statusIdx[db.status[v.Status]] = append(db.statusIdx[db.status[v.Status]], k)

		switch v.Status {
		case 0:
			db.statusNeqIdx[db.status[1]] = append(db.statusNeqIdx[db.status[1]], k)
			db.statusNeqIdx[db.status[2]] = append(db.statusNeqIdx[db.status[2]], k)
		case 1:
			db.statusNeqIdx[db.status[0]] = append(db.statusNeqIdx[db.status[0]], k)
			db.statusNeqIdx[db.status[2]] = append(db.statusNeqIdx[db.status[2]], k)
		case 2:
			db.statusNeqIdx[db.status[0]] = append(db.statusNeqIdx[db.status[0]], k)
			db.statusNeqIdx[db.status[1]] = append(db.statusNeqIdx[db.status[1]], k)
		}

		if _, ok := db.fnameIdx[db.fnames[v.FName]]; !ok {
			db.fnameIdx[db.fnames[v.FName]] = []int{}
		}
		db.fnameIdx[db.fnames[v.FName]] = append(db.fnameIdx[db.fnames[v.FName]], k)

		if db.fnames[v.FName] != "" {
			db.fnameNotNullIdx = append(db.fnameNotNullIdx, k)
		}

		if _, ok := db.snameIdx[db.snames[v.SName]]; !ok {
			db.snameIdx[db.snames[v.SName]] = []int{}
		}
		db.snameIdx[db.snames[v.SName]] = append(db.snameIdx[db.snames[v.SName]], k)

		if db.snames[v.SName] != "" {
			db.snameNotNullIdx = append(db.snameNotNullIdx, k)
		}

		phoneCode := ""
		if v.Phone != "" {
			s := strings.Split(v.Phone, "(")
			s = strings.Split(s[1], ")")
			phoneCode = s[0]
		}
		if _, ok := db.phoneCodeIdx[phoneCode]; !ok {
			db.phoneCodeIdx[phoneCode] = []int{}
		}
		db.phoneCodeIdx[phoneCode] = append(db.phoneCodeIdx[phoneCode], k)

		if _, ok := db.countryIdx[db.countries[v.Country]]; !ok {
			db.countryIdx[db.countries[v.Country]] = []int{}
		}
		db.countryIdx[db.countries[v.Country]] = append(db.countryIdx[db.countries[v.Country]], k)

		if db.countries[v.Country] != "" {
			db.countryNotNullIdx = append(db.countryNotNullIdx, k)
		}

		if _, ok := db.cityIdx[db.cities[v.City]]; !ok {
			db.cityIdx[db.cities[v.City]] = []int{}
		}
		db.cityIdx[db.cities[v.City]] = append(db.cityIdx[db.cities[v.City]], k)

		if db.cities[v.City] != "" {
			db.cityNotNullIdx = append(db.cityNotNullIdx, k)
		}

		db.emailIdx = append(db.emailIdx, emailIdxEntry{v.Email, k})

		domain := strings.Split(v.Email, "@")[1]
		if _, ok := db.emailDomainIdx[domain]; !ok {
			db.emailDomainIdx[domain] = []int{}
		}
		db.emailDomainIdx[domain] = append(db.emailDomainIdx[domain], k)

		if db.snames[v.SName] != "" {
			start := true
			var currentNode trieNode
			for _, char := range db.snames[v.SName] {
				if start {
					if _, ok := db.snamePrefixIdx.next[char]; !ok {
						db.snamePrefixIdx.next[char] = trieNode{next: make(map[int32]trieNode), ids: &[]int{}}
					}
					currentNode = db.snamePrefixIdx.next[char]
					*currentNode.ids = append(*currentNode.ids, k)
					start = false
				} else {
					if _, ok := currentNode.next[char]; !ok {
						currentNode.next[char] = trieNode{next: make(map[int32]trieNode), ids: &[]int{}}
					}
					currentNode = currentNode.next[char]
					*currentNode.ids = append(*currentNode.ids, k)
				}
			}
		}

		db.birthIdx = append(db.birthIdx, birthIdxEntry{v.Birth, k})

		year := time.Unix(int64(v.Birth), 0).Year()
		if _, ok := db.birthYearIdx[year]; !ok {
			db.birthYearIdx[year] = []int{}
		}
		db.birthYearIdx[year] = append(db.birthYearIdx[year], k)

		for _, interest := range v.Interests {
			if _, ok := db.interestsIdx[db.interests[interest]]; !ok {
				db.interestsIdx[db.interests[interest]] = []int{}
			}
			db.interestsIdx[db.interests[interest]] = append(db.interestsIdx[db.interests[interest]], k)
		}

		for _, like := range v.Likes {
			if _, ok := db.likesIdx[like.ID]; !ok {
				db.likesIdx[like.ID] = []int{}
			}
			db.likesIdx[like.ID] = append(db.likesIdx[like.ID], k)
		}

		if v.Premium == nil {
			if _, ok := db.premiumIdx[0]; !ok {
				db.premiumIdx[0] = []int{}
			}
			db.premiumIdx[0] = append(db.premiumIdx[0], k)
		} else {
			if _, ok := db.premiumIdx[1]; !ok {
				db.premiumIdx[1] = []int{}
			}
			db.premiumIdx[1] = append(db.premiumIdx[1], k)
			if v.PremiumNow(now) {
				if _, ok := db.premiumIdx[2]; !ok {
					db.premiumIdx[2] = []int{}
				}
				db.premiumIdx[2] = append(db.premiumIdx[2], k)
			}
		}

		db.ids = append(db.ids, k)
	}

	sort.Ints(db.ids)

	for _, v := range db.sexIdx {
		sort.Ints(v)
	}

	for _, v := range db.statusIdx {
		sort.Ints(v)
	}

	for _, v := range db.statusNeqIdx {
		sort.Ints(v)
	}

	for _, v := range db.fnameIdx {
		sort.Ints(v)
	}

	sort.Ints(db.fnameNotNullIdx)

	for _, v := range db.snameIdx {
		sort.Ints(v)
	}

	sort.Ints(db.snameNotNullIdx)

	for _, v := range db.phoneCodeIdx {
		sort.Ints(v)
	}

	for _, v := range db.countryIdx {
		sort.Ints(v)
	}

	sort.Ints(db.countryNotNullIdx)

	for _, v := range db.cityIdx {
		sort.Ints(v)
	}

	sort.Ints(db.cityNotNullIdx)

	sort.Slice(db.emailIdx, func(i, j int) bool {
		return db.emailIdx[i].email < db.emailIdx[j].email
	})

	for _, v := range db.emailDomainIdx {
		sort.Ints(v)
	}

	sort.Slice(db.birthIdx, func(i, j int) bool {
		return db.birthIdx[i].birth < db.birthIdx[j].birth
	})

	for _, v := range db.birthYearIdx {
		sort.Ints(v)
	}

	for _, v := range db.interestsIdx {
		sort.Ints(v)
	}

	for _, v := range db.likesIdx {
		sort.Ints(v)
	}

	for _, v := range db.premiumIdx {
		sort.Ints(v)
	}

	db.mu.Unlock()
	//runtime.GC()

	log.Println("indexes size", utils.Sizeof(
		db.sexIdx,
		db.statusIdx,
		db.fnameIdx,
		db.snameIdx,
		db.phoneCodeIdx,
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

	log.Println("db size", utils.Sizeof(db.accountsMin))

	return true
}

func (db *DB) Find(query M) models.Accounts {
	//log.Println("[DEBUG] query", query)
	res := make([][]int, 0)
	projection := make(map[string]void)
	for k, v := range query {
		switch k {
		case "sex_eq":
			res = append(res, db.sexIdx[v.(string)])
			projection["sex"] = void{}
		case "status_eq":
			res = append(res, db.statusIdx[v.(string)])
			projection["status"] = void{}
		case "status_neq":
			res = append(res, db.statusNeqIdx[v.(string)])
			projection["status"] = void{}
		case "fname_eq":
			res = append(res, db.fnameIdx[v.(string)])
			projection["fname"] = void{}
		case "fname_any":
			r := db.fnameIdx[v.([]string)[0]]
			if len(v.([]string)) > 1 {
				for i := 1; i < len(v.([]string)); i++ {
					r = sumSlicesUnique(r, db.fnameIdx[v.([]string)[i]])
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
				//todo make phone null idx
				r := make([]int, 0)
				for kp, vp := range db.phoneCodeIdx {
					if kp != "" {
						r = sumSlicesUnique(r, vp)
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
			r := db.cityIdx[v.([]string)[0]]
			if len(v.([]string)) > 1 {
				for i := 1; i < len(v.([]string)); i++ {
					r = sumSlicesUnique(r, db.cityIdx[v.([]string)[i]])
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
			res = append(res, db.getEmailLtIdxEntries(v.(string)))
		case "email_gt":
			res = append(res, db.getEmailGtIdxEntries(v.(string)))
		case "sname_starts":
			res = append(res, db.getSnamePrefixIds(v.(string)))
			projection["sname"] = void{}
		case "phone_code":
			res = append(res, db.phoneCodeIdx[v.(string)])
			projection["phone"] = void{}
		case "birth_lt":
			res = append(res, db.getBirthLtIdxEntries(v.(int)))
			projection["birth"] = void{}
		case "birth_gt":
			res = append(res, db.getBirthGtIdxEntries(v.(int)))
			projection["birth"] = void{}
		case "birth_year":
			if b, ok := db.birthYearIdx[v.(int)]; ok {
				res = append(res, b)
			} else {
				res = append(res, []int{})
			}
			projection["birth"] = void{}
		case "interests_contains":
			r := make([][]int, 0)
			for _, interest := range v.([]string) {
				r = append(r, db.interestsIdx[interest])
			}

			if len(r) == 0 {
				res = append(res, []int{})
				break
			}

			if len(r) == 1 {
				res = append(res, r[0])
				break
			}

			sort.Slice(r, func(i, j int) bool {
				return len(r[i]) < len(r[j])
			})

			var ids []int
		InterestsContainsLoop: //todo https://habr.com/post/250191/
			for _, id := range r[0] {
				for i := 1; i < len(r); i++ {
					if !intBinarySearch(r[i], id) {
						continue InterestsContainsLoop
					}
				}
				ids = append(ids, id)
			}
			res = append(res, ids)
		case "interests_any":
			r := db.interestsIdx[v.([]string)[0]]
			if len(v.([]string)) > 1 {
				for i := 1; i < len(v.([]string)); i++ {
					r = sumSlicesUnique(r, db.interestsIdx[v.([]string)[i]])
				}
			}
			res = append(res, r)
		case "likes_contains":
			r := make([][]int, 0)
			for _, like := range v.([]int) {
				r = append(r, db.likesIdx[like])
			}

			if len(r) == 0 {
				res = append(res, []int{})
				break
			}

			if len(r) == 1 {
				res = append(res, r[0])
				break
			}

			sort.Slice(r, func(i, j int) bool {
				return len(r[i]) < len(r[j])
			})

			var ids []int
		LikesContainsLoop: //todo https://habr.com/post/250191/
			for _, id := range r[0] {
				for i := 1; i < len(r); i++ {
					if !intBinarySearch(r[i], id) {
						continue LikesContainsLoop
					}
				}
				ids = append(ids, id)
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

	limit := query["limit"].(int)
	var ids []int
	accountsMin := make([]models.AccountMin, 0)

	if len(res) == 0 {
		for i := len(db.ids) - 1; i >= len(db.ids)-limit; i-- {
			ids = append(ids, db.ids[i])
			accountsMin = append(accountsMin, db.accountsMin[db.ids[i]])
		}
	} else if len(res) == 1 {
		for _, id := range res[0] {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool {
			return ids[i] > ids[j]
		})

		for i := 0; i < limit && i < len(ids); i++ {
			accountsMin = append(accountsMin, db.accountsMin[ids[i]])
		}
	} else {
		sort.Slice(res, func(i, j int) bool {
			return len(res[i]) < len(res[j])
		})

	MinResLoop: //todo https://habr.com/post/250191/
		for _, k := range res[0] {
			for i := 1; i < len(res); i++ {
				if !intBinarySearch(res[i], k) {
					continue MinResLoop
				}
			}
			ids = append(ids, k)
		}

		sort.Slice(ids, func(i, j int) bool {
			return ids[i] > ids[j]
		})

		for i := 0; i < limit && i < len(ids); i++ {
			accountsMin = append(accountsMin, db.accountsMin[ids[i]])
		}
	}

	accounts := models.Accounts{}
	accounts.Accounts = make([]models.Account, 0)
	for i, accountMin := range accountsMin {
		account := models.Account{ID: ids[i], Email: accountMin.Email}

		if _, ok := projection["fname"]; ok {
			account.FName = db.fnames[accountMin.FName]
		}

		if _, ok := projection["sname"]; ok {
			account.SName = db.snames[accountMin.SName]
		}

		if _, ok := projection["phone"]; ok {
			account.Phone = accountMin.Phone
		}

		if _, ok := projection["sex"]; ok {
			account.Sex = accountMin.Sex
		}

		if _, ok := projection["birth"]; ok {
			account.Birth = accountMin.Birth
		}

		if _, ok := projection["country"]; ok {
			account.Country = db.countries[accountMin.Country]
		}

		if _, ok := projection["city"]; ok {
			account.City = db.cities[accountMin.City]
		}

		if _, ok := projection["status"]; ok {
			account.Status = db.status[accountMin.Status]
		}

		if _, ok := projection["premium"]; ok {
			account.Premium = accountMin.Premium
		}

		accounts.Accounts = append(accounts.Accounts, account)
	}

	return accounts
}

func intBinarySearch(slice []int, element int) bool {
	low := 0
	high := len(slice) - 1

	for low <= high {
		mid := (low + high) / 2
		guess := slice[mid]
		if guess == element {
			return true
		}

		if guess > element {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}

	return false
}

func sumSlicesUnique(slice1, slice2 []int) []int {
	n := len(slice1)
	m := len(slice2)
	var i, j, k int
	result := make([]int, m+n)
	for (i < n) && (j < m) {
		if slice1[i] == slice2[j] {
			result[k] = slice1[i]
			k++
			i++
			j++
		} else if slice1[i] < slice2[j] {
			result[k] = slice1[i]
			k++
			i++
		} else {
			result[k] = slice2[j]
			k++
			j++
		}
	}

	for i < n {
		result[k] = slice1[i]
		k++
		i++
	}
	for j < m {
		result[k] = slice2[j]
		k++
		j++
	}

	return result[:k]
}

func deduplicate(slice []int) []int {
	j := 0
	for i := 1; i < len(slice); i++ {
		if slice[j] == slice[i] {
			continue
		}
		j++
		// preserve the original data
		// in[i], in[j] = in[j], in[i]
		// only set what is required
		slice[j] = slice[i]
	}
	return slice[:j+1]
}
