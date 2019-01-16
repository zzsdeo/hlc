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
	mu                *sync.Mutex
	ids               []int
	sexIdx            map[string]map[int]void
	statusIdx         map[string]map[int]void
	statusNeqIdx      map[string]map[int]void
	fnameIdx          map[string]map[int]void
	fnameNotNullIdx   map[int]void
	snameIdx          map[string]map[int]void
	snameNotNullIdx   map[int]void
	phoneCodeIdx      map[string]map[int]void
	countryIdx        map[string]map[int]void
	countryNotNullIdx map[int]void
	cityIdx           map[string]map[int]void
	cityNotNullIdx    map[int]void
	emailIdx          []emailIdxEntry
	emailDomainIdx    map[string]map[int]void
	snamePrefixIdx    trieNode
	birthIdx          []birthIdxEntry
	birthYearIdx      map[int]map[int]void
	interestsIdx      map[string]map[int]void
	likesIdx          map[int]map[int]void
	premiumIdx        map[byte]map[int]void // 0-null, 1-not_null, 2-now
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
	ids  map[int]void
}

func NewDB() *DB {
	return &DB{
		mu:                &sync.Mutex{},
		ids:               make([]int, 0),
		sexIdx:            map[string]map[int]void{},
		statusIdx:         map[string]map[int]void{},
		statusNeqIdx:      map[string]map[int]void{"свободны": {}, "заняты": {}, "всё сложно": {}},
		fnameIdx:          map[string]map[int]void{},
		fnameNotNullIdx:   map[int]void{},
		snameIdx:          map[string]map[int]void{},
		snameNotNullIdx:   map[int]void{},
		phoneCodeIdx:      map[string]map[int]void{},
		countryIdx:        map[string]map[int]void{},
		countryNotNullIdx: map[int]void{},
		cityIdx:           map[string]map[int]void{},
		cityNotNullIdx:    map[int]void{},
		emailIdx:          []emailIdxEntry{},
		emailDomainIdx:    map[string]map[int]void{},
		snamePrefixIdx:    trieNode{next: make(map[int32]trieNode), ids: make(map[int]void)},
		birthIdx:          []birthIdxEntry{},
		birthYearIdx:      map[int]map[int]void{},
		interestsIdx:      map[string]map[int]void{},
		likesIdx:          map[int]map[int]void{},
		premiumIdx:        map[byte]map[int]void{},

		minData: minData{
			accountsMin: map[int]models.AccountMin{},
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

func (db *DB) getSnamePrefixIds(prefix string) map[int]void {
	start := true
	var currentNode trieNode
	for _, char := range prefix {
		if start {
			if _, ok := db.snamePrefixIdx.next[char]; !ok {
				return make(map[int]void)
			}
			currentNode = db.snamePrefixIdx.next[char]
			start = false
		} else {
			if _, ok := currentNode.next[char]; !ok {
				return make(map[int]void)
			}
			currentNode = currentNode.next[char]
		}
	}
	return currentNode.ids
}

func (db *DB) LoadMinData(accounts []models.Account) {
	db.mu.Lock()
	for _, account := range accounts {
		accountMin := models.AccountMin{
			Email:   account.Email,
			Phone:   account.Phone,
			Birth:   account.Birth,
			Joined:  account.Joined,
			Premium: account.Premium,
			//Likes:   account.Likes,
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

		db.accountsMin[account.ID] = accountMin
	}
	db.mu.Unlock()
	runtime.GC()
}

//func (db *DB) LoadMinData2(accounts []models.Account, now int) {
//	//db.mu.Lock()
//	for i := range accounts {
//		db.AddAccount(accounts[i])
//		db.UpdateIndex(accounts[i], now)
//	}
//
//	sort.Slice(db.emailIdx, func(i, j int) bool {
//		return db.emailIdx[i].email < db.emailIdx[j].email
//	})
//	sort.Slice(db.birthIdx, func(i, j int) bool {
//		return db.birthIdx[i].birth < db.birthIdx[j].birth
//	})
//	sort.Slice(db.ids, func(i, j int) bool {
//		return db.ids[i] > db.ids[j]
//	})
//	//db.mu.Unlock()
//	//runtime.GC()
//}

func (db *DB) LoadMinData2(accounts []models.Account, now int) {
	jobs := make(chan models.Account, len(accounts))
	numOfWorkers := 100
	for numOfWorkers >= 0 {
		go db.accountWorker(jobs)
		go db.indexWorker(jobs, now)
		numOfWorkers--
	}
	for i := range accounts {
		jobs <- accounts[i]
	}
	close(jobs)
}

func (db *DB) SortSlices() {
	sort.Slice(db.emailIdx, func(i, j int) bool {
		return db.emailIdx[i].email < db.emailIdx[j].email
	})
	sort.Slice(db.birthIdx, func(i, j int) bool {
		return db.birthIdx[i].birth < db.birthIdx[j].birth
	})
	sort.Slice(db.ids, func(i, j int) bool {
		return db.ids[i] > db.ids[j]
	})
}

func (db *DB) accountWorker(jobs <-chan models.Account) {
	for j := range jobs {
		db.AddAccount(j)
	}
}

func (db *DB) indexWorker(jobs <-chan models.Account, now int) {
	for j := range jobs {
		db.UpdateIndex(j, now)
	}
}

func (db *DB) CalculateData() {
	log.Println("db size", utils.Sizeof(db.accountsMin))

	log.Println("db.sexIdx", utils.Sizeof(db.sexIdx))
	log.Println("db.statusIdx", utils.Sizeof(db.statusIdx))
	log.Println("db.fnameIdx", utils.Sizeof(db.fnameIdx))
	log.Println("db.snameIdx", utils.Sizeof(db.snameIdx))
	log.Println("db.phoneCodeIdx", utils.Sizeof(db.phoneCodeIdx))
	log.Println("db.countryIdx", utils.Sizeof(db.countryIdx))
	log.Println("db.cityIdx", utils.Sizeof(db.cityIdx))
	log.Println("db.emailIdx", utils.Sizeof(db.emailIdx))
	log.Println("db.emailDomainIdx", utils.Sizeof(db.emailDomainIdx))
	log.Println("db.snamePrefixIdx", utils.Sizeof(db.snamePrefixIdx))
	log.Println("db.birthIdx", utils.Sizeof(db.birthIdx))
	log.Println("db.birthYearIdx", utils.Sizeof(db.birthYearIdx))
	log.Println("db.interestsIdx", utils.Sizeof(db.interestsIdx))
	log.Println("db.likesIdx", utils.Sizeof(db.likesIdx))
	log.Println("db.premiumIdx", utils.Sizeof(db.premiumIdx))
}

func (db *DB) AddAccount(account models.Account) {
	db.mu.Lock()
	accountMin := models.AccountMin{
		Email:   account.Email,
		Phone:   account.Phone,
		Birth:   account.Birth,
		Joined:  account.Joined,
		Premium: account.Premium,
		//Likes:   account.Likes,
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

	db.accountsMin[account.ID] = accountMin

	db.mu.Unlock()
}

func (db *DB) UpdateIndex(account models.Account, now int) {
	db.mu.Lock()
	if _, ok := db.sexIdx[account.Sex]; !ok {
		db.sexIdx[account.Sex] = map[int]void{}
	}
	db.sexIdx[account.Sex][account.ID] = void{}

	if _, ok := db.statusIdx[account.Status]; !ok {
		db.statusIdx[account.Status] = map[int]void{}
	}
	db.statusIdx[account.Status][account.ID] = void{}

	switch account.Status {
	case "свободны":
		db.statusNeqIdx[db.status[1]][account.ID] = void{}
		db.statusNeqIdx[db.status[2]][account.ID] = void{}
	case "заняты":
		db.statusNeqIdx[db.status[0]][account.ID] = void{}
		db.statusNeqIdx[db.status[2]][account.ID] = void{}
	case "всё сложно":
		db.statusNeqIdx[db.status[0]][account.ID] = void{}
		db.statusNeqIdx[db.status[1]][account.ID] = void{}
	}

	if _, ok := db.fnameIdx[account.FName]; !ok {
		db.fnameIdx[account.FName] = map[int]void{}
	}
	db.fnameIdx[account.FName][account.ID] = void{}

	if account.FName != "" {
		db.fnameNotNullIdx[account.ID] = void{}
	}

	if _, ok := db.snameIdx[account.SName]; !ok {
		db.snameIdx[account.SName] = map[int]void{}
	}
	db.snameIdx[account.SName][account.ID] = void{}

	if account.SName != "" {
		db.snameNotNullIdx[account.ID] = void{}
	}

	phoneCode := ""
	if account.Phone != "" {
		s := strings.Split(account.Phone, "(")
		s = strings.Split(s[1], ")")
		phoneCode = s[0]
	}
	if _, ok := db.phoneCodeIdx[phoneCode]; !ok {
		db.phoneCodeIdx[phoneCode] = map[int]void{}
	}
	db.phoneCodeIdx[phoneCode][account.ID] = void{}

	if _, ok := db.countryIdx[account.Country]; !ok {
		db.countryIdx[account.Country] = map[int]void{}
	}
	db.countryIdx[account.Country][account.ID] = void{}

	if account.Country != "" {
		db.countryNotNullIdx[account.ID] = void{}
	}

	if _, ok := db.cityIdx[account.City]; !ok {
		db.cityIdx[account.City] = map[int]void{}
	}
	db.cityIdx[account.City][account.ID] = void{}

	if account.City != "" {
		db.cityNotNullIdx[account.ID] = void{}
	}

	db.emailIdx = append(db.emailIdx, emailIdxEntry{account.Email, account.ID})

	domain := strings.Split(account.Email, "@")[1]
	if _, ok := db.emailDomainIdx[domain]; !ok {
		db.emailDomainIdx[domain] = map[int]void{}
	}
	db.emailDomainIdx[domain][account.ID] = void{}

	if account.SName != "" {
		start := true
		var currentNode trieNode
		for _, char := range account.SName {
			if start {
				if _, ok := db.snamePrefixIdx.next[char]; !ok {
					db.snamePrefixIdx.next[char] = trieNode{next: make(map[int32]trieNode), ids: make(map[int]void)}
				}
				currentNode = db.snamePrefixIdx.next[char]
				currentNode.ids[account.ID] = void{}
				start = false
			} else {
				if _, ok := currentNode.next[char]; !ok {
					currentNode.next[char] = trieNode{next: make(map[int32]trieNode), ids: make(map[int]void)}
				}
				currentNode = currentNode.next[char]
				currentNode.ids[account.ID] = void{}
			}
		}
	}

	db.birthIdx = append(db.birthIdx, birthIdxEntry{account.Birth, account.ID})

	year := time.Unix(int64(account.Birth), 0).Year()
	if _, ok := db.birthYearIdx[year]; !ok {
		db.birthYearIdx[year] = map[int]void{}
	}
	db.birthYearIdx[year][account.ID] = void{}

	for _, interest := range account.Interests {
		if _, ok := db.interestsIdx[interest]; !ok {
			db.interestsIdx[interest] = map[int]void{}
		}
		db.interestsIdx[interest][account.ID] = void{}
	}

	for _, like := range account.Likes {
		if _, ok := db.likesIdx[like.ID]; !ok {
			db.likesIdx[like.ID] = map[int]void{}
		}
		db.likesIdx[like.ID][account.ID] = void{}
	}

	if account.Premium == nil {
		if _, ok := db.premiumIdx[0]; !ok {
			db.premiumIdx[0] = map[int]void{}
		}
		db.premiumIdx[0][account.ID] = void{}
	} else {
		if _, ok := db.premiumIdx[1]; !ok {
			db.premiumIdx[1] = map[int]void{}
		}
		db.premiumIdx[1][account.ID] = void{}
		if account.PremiumNow(now) {
			if _, ok := db.premiumIdx[2]; !ok {
				db.premiumIdx[2] = map[int]void{}
			}
			db.premiumIdx[2][account.ID] = void{}
		}
	}

	db.ids = append(db.ids, account.ID)

	//sort.Slice(db.emailIdx, func(i, j int) bool {
	//	return db.emailIdx[i].email < db.emailIdx[j].email
	//})
	//sort.Slice(db.birthIdx, func(i, j int) bool {
	//	return db.birthIdx[i].birth < db.birthIdx[j].birth
	//})
	//sort.Slice(db.ids, func(i, j int) bool {
	//	return db.ids[i] > db.ids[j]
	//})

	db.mu.Unlock()
}

func (db *DB) CreateIndexes(now int) bool {
	db.mu.Lock()
	for k, v := range db.accountsMin {
		if _, ok := db.sexIdx[db.sex[v.Sex]]; !ok {
			db.sexIdx[db.sex[v.Sex]] = map[int]void{}
		}
		db.sexIdx[db.sex[v.Sex]][k] = void{}

		if _, ok := db.statusIdx[db.status[v.Status]]; !ok {
			db.statusIdx[db.status[v.Status]] = map[int]void{}
		}
		db.statusIdx[db.status[v.Status]][k] = void{}

		switch v.Status {
		case 0:
			db.statusNeqIdx[db.status[1]][k] = void{}
			db.statusNeqIdx[db.status[2]][k] = void{}
		case 1:
			db.statusNeqIdx[db.status[0]][k] = void{}
			db.statusNeqIdx[db.status[2]][k] = void{}
		case 2:
			db.statusNeqIdx[db.status[0]][k] = void{}
			db.statusNeqIdx[db.status[1]][k] = void{}
		}

		if _, ok := db.fnameIdx[db.fnames[v.FName]]; !ok {
			db.fnameIdx[db.fnames[v.FName]] = map[int]void{}
		}
		db.fnameIdx[db.fnames[v.FName]][k] = void{}

		if db.fnames[v.FName] != "" {
			db.fnameNotNullIdx[k] = void{}
		}

		if _, ok := db.snameIdx[db.snames[v.SName]]; !ok {
			db.snameIdx[db.snames[v.SName]] = map[int]void{}
		}
		db.snameIdx[db.snames[v.SName]][k] = void{}

		if db.snames[v.SName] != "" {
			db.snameNotNullIdx[k] = void{}
		}

		phoneCode := ""
		if v.Phone != "" {
			s := strings.Split(v.Phone, "(")
			s = strings.Split(s[1], ")")
			phoneCode = s[0]
		}
		if _, ok := db.phoneCodeIdx[phoneCode]; !ok {
			db.phoneCodeIdx[phoneCode] = map[int]void{}
		}
		db.phoneCodeIdx[phoneCode][k] = void{}

		if _, ok := db.countryIdx[db.countries[v.Country]]; !ok {
			db.countryIdx[db.countries[v.Country]] = map[int]void{}
		}
		db.countryIdx[db.countries[v.Country]][k] = void{}

		if db.countries[v.Country] != "" {
			db.countryNotNullIdx[k] = void{}
		}

		if _, ok := db.cityIdx[db.cities[v.City]]; !ok {
			db.cityIdx[db.cities[v.City]] = map[int]void{}
		}
		db.cityIdx[db.cities[v.City]][k] = void{}

		if db.cities[v.City] != "" {
			db.cityNotNullIdx[k] = void{}
		}

		db.emailIdx = append(db.emailIdx, emailIdxEntry{v.Email, k})

		domain := strings.Split(v.Email, "@")[1]
		if _, ok := db.emailDomainIdx[domain]; !ok {
			db.emailDomainIdx[domain] = map[int]void{}
		}
		db.emailDomainIdx[domain][k] = void{}

		if db.snames[v.SName] != "" {
			start := true
			var currentNode trieNode
			for _, char := range db.snames[v.SName] {
				if start {
					if _, ok := db.snamePrefixIdx.next[char]; !ok {
						db.snamePrefixIdx.next[char] = trieNode{next: make(map[int32]trieNode), ids: make(map[int]void)}
					}
					currentNode = db.snamePrefixIdx.next[char]
					currentNode.ids[k] = void{}
					start = false
				} else {
					if _, ok := currentNode.next[char]; !ok {
						currentNode.next[char] = trieNode{next: make(map[int32]trieNode), ids: make(map[int]void)}
					}
					currentNode = currentNode.next[char]
					currentNode.ids[k] = void{}
				}
			}
		}

		db.birthIdx = append(db.birthIdx, birthIdxEntry{v.Birth, k})

		year := time.Unix(int64(v.Birth), 0).Year()
		if _, ok := db.birthYearIdx[year]; !ok {
			db.birthYearIdx[year] = map[int]void{}
		}
		db.birthYearIdx[year][k] = void{}

		for _, interest := range v.Interests {
			if _, ok := db.interestsIdx[db.interests[interest]]; !ok {
				db.interestsIdx[db.interests[interest]] = map[int]void{}
			}
			db.interestsIdx[db.interests[interest]][k] = void{}
		}

		for _, like := range v.Likes {
			if _, ok := db.likesIdx[like.ID]; !ok {
				db.likesIdx[like.ID] = map[int]void{}
			}
			db.likesIdx[like.ID][k] = void{}
		}

		if v.Premium == nil {
			if _, ok := db.premiumIdx[0]; !ok {
				db.premiumIdx[0] = map[int]void{}
			}
			db.premiumIdx[0][k] = void{}
		} else {
			if _, ok := db.premiumIdx[1]; !ok {
				db.premiumIdx[1] = map[int]void{}
			}
			db.premiumIdx[1][k] = void{}
			if v.PremiumNow(now) {
				if _, ok := db.premiumIdx[2]; !ok {
					db.premiumIdx[2] = map[int]void{}
				}
				db.premiumIdx[2][k] = void{}
			}
		}

		db.ids = append(db.ids, k)
	}
	runtime.GC()

	sort.Slice(db.emailIdx, func(i, j int) bool {
		return db.emailIdx[i].email < db.emailIdx[j].email
	})
	sort.Slice(db.birthIdx, func(i, j int) bool {
		return db.birthIdx[i].birth < db.birthIdx[j].birth
	})
	sort.Slice(db.ids, func(i, j int) bool {
		return db.ids[i] > db.ids[j]
	})
	db.mu.Unlock()

	runtime.GC()

	//log.Println("indexes size", utils.Sizeof(
	//	db.sexIdx,
	//	db.statusIdx,
	//	db.fnameIdx,
	//	db.snameIdx,
	//	db.phoneCodeIdx,
	//	db.countryIdx,
	//	db.cityIdx,
	//	db.emailIdx,
	//	db.emailDomainIdx,
	//	db.snamePrefixIdx,
	//	db.birthIdx,
	//	db.birthYearIdx,
	//	db.interestsIdx,
	//	db.likesIdx,
	//	db.premiumIdx))
	//
	return true
}

func (db *DB) Find(query M) models.Accounts {
	//log.Println("[DEBUG] query", query)
	res := make([]map[int]void, 0)
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
