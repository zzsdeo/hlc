package store

import (
	"hlc/app/models"
	"log"
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

func (db *DB) LoadData(accounts []models.Account) {
	//jobs := make(chan models.Account, len(accounts))
	//numOfWorkers := 100
	//for numOfWorkers >= 0 {
	//	go db.accountWorker(jobs)
	//	numOfWorkers--
	//}
	//for i := range accounts {
	//	jobs <- accounts[i]
	//}
	//close(jobs)

	for i := range accounts {
		db.AddAccount(accounts[i])
	}
}

func (db *DB) SortDB() {
	sort.Slice(db.accountsMin, func(i, j int) bool {
		return db.accountsMin[i].ID > db.accountsMin[j].ID
	})
	for _, a := range db.accountsMin[:10] {
		log.Println(a.ID)
	}
}

//func (db *DB) accountWorker(jobs <-chan models.Account) {
//	for j := range jobs {
//		db.AddAccount(j)
//	}
//}

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
	var accountsMin []models.AccountMin
	projection := make(map[string]void)
MainLoop:
	for _, accountMin := range db.accountsMin {
	InnerLoop:
		for k, v := range query {
			switch k {
			case "sex_eq":
				if db.sex[accountMin.Sex] != v.(string) {
					continue MainLoop
				}
				projection["sex"] = void{}
			case "status_eq":
				if db.status[accountMin.Status] != v.(string) {
					continue MainLoop
				}
				projection["status"] = void{}
			case "status_neq":
				if db.status[accountMin.Status] == v.(string) {
					continue MainLoop
				}
				projection["status"] = void{}
			case "fname_eq":
				if db.fnames[accountMin.FName] != v.(string) {
					continue MainLoop
				}
				projection["fname"] = void{}
			case "fname_any":
				for ii := range v.([]string) {
					if db.fnames[accountMin.FName] == v.([]string)[ii] {
						projection["fname"] = void{}
						continue InnerLoop
					}
				}
				continue MainLoop
			case "fname_null":
				switch v.(string) {
				case "0":
					if db.fnames[accountMin.FName] == "" {
						continue MainLoop
					}
				case "1":
					if db.fnames[accountMin.FName] != "" {
						continue MainLoop
					}
				}
				projection["fname"] = void{}
			case "sname_eq":
				if db.snames[accountMin.SName] != v.(string) {
					continue MainLoop
				}
				projection["sname"] = void{}
			case "sname_null":
				switch v.(string) {
				case "0":
					if db.snames[accountMin.SName] == "" {
						continue MainLoop
					}
				case "1":
					if db.snames[accountMin.SName] != "" {
						continue MainLoop
					}
				}
				projection["sname"] = void{}
			case "phone_null":
				switch v.(string) {
				case "0":
					if accountMin.Phone == "" {
						continue MainLoop
					}
				case "1":
					if accountMin.Phone != "" {
						continue MainLoop
					}
				}
				projection["phone"] = void{}
			case "country_eq":
				if db.countries[accountMin.Country] != v.(string) {
					continue MainLoop
				}
				projection["country"] = void{}
			case "country_null":
				switch v.(string) {
				case "0":
					if db.countries[accountMin.Country] == "" {
						continue MainLoop
					}
				case "1":
					if db.countries[accountMin.Country] != "" {
						continue MainLoop
					}
				}
				projection["country"] = void{}
			case "city_eq":
				if db.cities[accountMin.City] != v.(string) {
					continue MainLoop
				}
				projection["city"] = void{}
			case "city_any":
				for ii := range v.([]string) {
					if db.cities[accountMin.City] == v.([]string)[ii] {
						projection["city"] = void{}
						continue InnerLoop
					}
				}
				continue MainLoop
			case "city_null":
				switch v.(string) {
				case "0":
					if db.cities[accountMin.City] == "" {
						continue MainLoop
					}
				case "1":
					if db.cities[accountMin.City] != "" {
						continue MainLoop
					}
				}
				projection["city"] = void{}
			case "email_domain":
				if !strings.Contains(accountMin.Email, "@"+v.(string)) {
					continue MainLoop
				}
			case "email_lt":
				if accountMin.Email > v.(string) {
					continue MainLoop
				}
			case "email_gt":
				if accountMin.Email < v.(string) {
					continue MainLoop
				}
			case "sname_starts":
				if !strings.HasPrefix(db.snames[accountMin.SName], v.(string)) {
					continue MainLoop
				}
				projection["sname"] = void{}
			case "phone_code":
				if !strings.Contains(accountMin.Phone, "("+v.(string)+")") {
					continue MainLoop
				}
				projection["phone"] = void{}
			case "birth_lt":
				if accountMin.Birth > v.(int) {
					continue MainLoop
				}
				projection["birth"] = void{}
			case "birth_gt":
				if accountMin.Birth < v.(int) {
					continue MainLoop
				}
				projection["birth"] = void{}
			case "birth_year":
				if time.Unix(int64(accountMin.Birth), 0).Year() != v.(int) {
					continue MainLoop
				}
				projection["birth"] = void{}
			case "interests_contains":
			InterestsContainsLoop:
				for ii := range v.([]string) {
					for iii := range accountMin.Interests {
						if db.interests[accountMin.Interests[iii]] == v.([]string)[ii] {
							continue InterestsContainsLoop
						}
					}
					continue MainLoop
				}
			case "interests_any":
				for ii := range v.([]string) {
					for iii := range accountMin.Interests {
						if db.interests[accountMin.Interests[iii]] == v.([]string)[ii] {
							continue InnerLoop
						}
					}
				}
				continue MainLoop
			case "likes_contains":
			LikesContainsLoop:
				for ii := range v.([]int) {
					for iii := range accountMin.Likes {
						if accountMin.Likes[iii].ID == v.([]int)[ii] {
							continue LikesContainsLoop
						}
					}
					continue MainLoop
				}

			case "premium_now":
				if !accountMin.PremiumNow(v.(int)) {
					continue MainLoop
				}
				projection["premium"] = void{}
			case "premium_null":
				switch v.(string) {
				case "0":
					if accountMin.Premium == nil {
						continue MainLoop
					}
				case "1":
					if accountMin.Premium != nil {
						continue MainLoop
					}
				}
				projection["premium"] = void{}
			}
		}
		accountsMin = append(accountsMin, accountMin)
	}

	limit := query["limit"].(int)
	accounts := models.Accounts{}
	accounts.Accounts = make([]models.Account, 0)
	for i := 0; i < limit && i < len(accountsMin); i++ {
		account := models.Account{ID: accountsMin[i].ID, Email: accountsMin[i].Email}

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
