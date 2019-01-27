package store

import (
	"database/sql"
	"fmt"
	"hlc/app/models"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const createDB = `
CREATE TABLE accounts (
	id INTEGER PRIMARY KEY NOT NULL,
	email TEXT NOT NULL,
	fname TEXT NOT NULL,
	sname TEXT NOT NULL,
	phone TEXT NOT NULL,
	sex TEXT NOT NULL,
	birth INTEGER NOT NULL,
	country TEXT NOT NULL,
	city TEXT NOT NULL,
	joined INTEGER NOT NULL,
	status TEXT NOT NULL
) WITHOUT ROWID;
CREATE TABLE likes (
	id INTEGER NOT NULL,
	ts INTEGER NOT NULL,
	accountid INTEGER NOT NULL,
	FOREIGN KEY (accountid) REFERENCES accounts(id),
	FOREIGN KEY (id) REFERENCES accounts(id)
);
CREATE TABLE interests (
	interest TEXT NOT NULL,
	accountid INTEGER NOT NULL,
	FOREIGN KEY (accountid) REFERENCES accounts(id)
);
CREATE TABLE premiums (
	start INTEGER NOT NULL,
	finish INTEGER NOT NULL,
	premium INTEGER NOT NULL,
	accountid INTEGER NOT NULL,
	FOREIGN KEY (accountid) REFERENCES accounts(id)
);
`

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

	sql *sql.DB
}

type void struct{}

type M map[string]interface{}

func NewDB() (*DB, error) {

	sql, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	sql.SetMaxOpenConns(0)
	err = sql.Ping()
	if err != nil {
		return nil, err
	}

	_, err = sql.Exec(createDB)
	if err != nil {
		return nil, err
	}

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
		sql: sql,
	}, nil
}

func (db *DB) LoadData(accounts []models.Account, now int) {
	for i := range accounts {
		db.AddAccount(accounts[i], now)
	}
}

func (db *DB) AddAccount(account models.Account, now int) {
	db.sql.Exec("INSERT INTO accounts VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		account.ID,
		account.Email,
		account.FName,
		account.SName,
		account.Phone,
		account.Sex,
		account.Birth,
		account.Country,
		account.City,
		account.Joined,
		account.Status,
	)
	for i := range account.Interests {
		db.sql.Exec("INSERT INTO interests VALUES (?, ?)",
			account.Interests[i], account.ID)
	}
	for i := range account.Likes {
		db.sql.Exec("INSERT INTO likes VALUES (?, ?, ?)",
			account.Likes[i].ID, account.Likes[i].TS, account.ID)
	}

	if account.Premium != nil {
		db.sql.Exec("INSERT INTO premiums VALUES (?, ?, ?, ?)",
			account.Premium.Start, account.Premium.Finish, account.PremiumNow(now), account.ID)
	} else {
		db.sql.Exec("INSERT INTO premiums VALUES (?, ?, ?, ?)",
			0, 0, account.PremiumNow(now), account.ID)
	}
}

func (db *DB) Find(query M) models.Accounts {
	projection := make(map[string]void)
	accounts := models.Accounts{}
	accounts.Accounts = make([]models.Account, 0)

	queryString := "SELECT id, email"
	from := " FROM accounts"
	whereClause := "WHERE"

	for k, v := range query {
		switch k {
		case "sex_eq":
			whereClause += " sex = '" + v.(string) + "' AND"
			projection["sex"] = void{}
		case "status_eq":
			whereClause += " status = '" + v.(string) + "' AND"
			projection["status"] = void{}
		case "status_neq":
			whereClause += " status != '" + v.(string) + "' AND"
			projection["status"] = void{}
		case "fname_eq":
			whereClause += " fname = '" + v.(string) + "' AND"
			projection["fname"] = void{}
		case "fname_any":
			c := "('" + v.([]string)[0] + "'"
			if len(v.([]string)) > 1 {
				for i := 1; i < len(v.([]string)); i++ {
					c += ", '" + v.([]string)[i] + "'"
				}
			}
			c += ")"
			whereClause += " fname IN " + c + " AND"
			projection["fname"] = void{}
		case "fname_null":
			switch v.(string) {
			case "0":
				whereClause += " fname != '' AND"
			case "1":
				whereClause += " fname = '' AND"
			}
			projection["fname"] = void{}
		case "sname_eq":
			whereClause += " sname = '" + v.(string) + "' AND"
			projection["sname"] = void{}
		case "sname_null":
			switch v.(string) {
			case "0":
				whereClause += " sname != '' AND"
			case "1":
				whereClause += " sname = '' AND"
			}
			projection["sname"] = void{}
		case "phone_null":
			switch v.(string) {
			case "0":
				whereClause += " phone != '' AND"
			case "1":
				whereClause += " phone = '' AND"
			}
			projection["phone"] = void{}
		case "country_eq":
			whereClause += " country = '" + v.(string) + "' AND"
			projection["country"] = void{}
		case "country_null":
			switch v.(string) {
			case "0":
				whereClause += " country != '' AND"
			case "1":
				whereClause += " country = '' AND"
			}
			projection["country"] = void{}
		case "city_eq":
			whereClause += " city = '" + v.(string) + "' AND"
			projection["city"] = void{}
		case "city_any":
			c := "('" + v.([]string)[0] + "'"
			if len(v.([]string)) > 1 {
				for i := 1; i < len(v.([]string)); i++ {
					c += ", '" + v.([]string)[i] + "'"
				}
			}
			c += ")"
			whereClause += " city IN " + c + " AND"
			projection["city"] = void{}
		case "city_null":
			switch v.(string) {
			case "0":
				whereClause += " city != '' AND"
			case "1":
				whereClause += " city = '' AND"
			}
			projection["city"] = void{}
		case "email_domain":
			whereClause += " email LIKE '%@" + v.(string) + "%' AND"
		case "email_lt":
			whereClause += " email < '" + v.(string) + "' AND"
		case "email_gt":
			whereClause += " email > '" + v.(string) + "' AND"
		case "sname_starts":
			whereClause += " sname LIKE '" + v.(string) + "%' AND"
			projection["sname"] = void{}
		case "phone_code":
			whereClause += " phone LIKE '%(" + v.(string) + ")%' AND"
			projection["phone"] = void{}
		case "birth_lt":
			whereClause += " birth < " + v.(string) + " AND"
			projection["birth"] = void{}
		case "birth_gt":
			whereClause += " birth > " + v.(string) + " AND"
			projection["birth"] = void{}
		case "birth_year":
			start := time.Date(v.(int), time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
			finish := time.Date(v.(int)+1, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
			whereClause += " birth >= " + strconv.Itoa(int(start)) + " AND birth < " + strconv.Itoa(int(finish)) + " AND"
			//whereClause += " (birth BETWEEN " + strconv.Itoa(start) + " AND " + strconv.Itoa(finish) + ") AND"
			projection["birth"] = void{}
		case "interests_contains":
			from += " JOIN interests ON interests.accountid = accounts.id"
			c := "('" + v.([]string)[0] + "'"
			if len(v.([]string)) > 1 {
				for i := 1; i < len(v.([]string)); i++ {
					c += ", '" + v.([]string)[i] + "'"
				}
			}
			c += ")"
			whereClause += " interest ALL " + c + " AND"
		case "interests_any":
			from += " INNER JOIN interests ON interests.accountid = accounts.id"
			c := "('" + v.([]string)[0] + "'"
			if len(v.([]string)) > 1 {
				for i := 1; i < len(v.([]string)); i++ {
					c += ", '" + v.([]string)[i] + "'"
				}
			}
			c += ")"
			whereClause += " interest IN " + c + " AND"
		case "likes_contains":
			from += " JOIN likes ON likes.accountid = accounts.id"
			c := "(" + v.([]string)[0]
			if len(v.([]string)) > 1 {
				for i := 1; i < len(v.([]string)); i++ {
					c += ", " + v.([]string)[i]
				}
			}
			c += ")"
			whereClause += " likes.id ALL " + c + " AND"
		case "premium_now":
			from += " JOIN premiums ON premiums.accountid = accounts.id"
			whereClause += " premium = 2 AND"
			projection["premium"] = void{}
		case "premium_null":
			from += " JOIN premiums ON premiums.accountid = accounts.id"
			switch v.(string) {
			case "0":
				whereClause += " premium = 1 AND premium = 2 AND"
			case "1":
				whereClause += " premium = 0 AND"
			}
			projection["premium"] = void{}
		}
	}

	if _, ok := projection["fname"]; ok {
		queryString += ", fname"
	}

	if _, ok := projection["sname"]; ok {
		queryString += ", sname"
	}

	if _, ok := projection["phone"]; ok {
		queryString += ", phone"
	}

	if _, ok := projection["sex"]; ok {
		queryString += ", sex"
	}

	if _, ok := projection["birth"]; ok {
		queryString += ", birth"
	}

	if _, ok := projection["country"]; ok {
		queryString += ", country"
	}

	if _, ok := projection["city"]; ok {
		queryString += ", city"
	}

	if _, ok := projection["status"]; ok {
		queryString += ", status"
	}

	if _, ok := projection["premium"]; ok {
		queryString += ", premiums.start, premiums.finish"
	}

	if whereClause == "WHERE" {
		whereClause = ""
	} else {
		whereClause = strings.TrimRight(whereClause, "AND")
	}

	queryString += from + " " + whereClause + "ORDER BY id DESC LIMIT " + query["limit"].(string)

	fmt.Println(queryString) //TODO:

	rows, err := db.sql.Query(queryString)
	//rows, err := db.sql.Query("SELECT id, email, country FROM accounts WHERE country = 'Испмаль' ORDER BY id DESC LIMIT 20")
	if err != nil {
		log.Println("ERROR ", err)
		return accounts
	}

	for rows.Next() {
		account := &models.Account{}
		columns, err := rows.Columns()
		if err != nil {
			log.Println("ERROR ", err)
			return accounts
		}
		dest := make([]interface{}, 0)
		for i := range columns {
			switch columns[i] {
			case "id":
				dest = append(dest, &account.ID)
			case "email":
				dest = append(dest, &account.Email)
			case "fname":
				dest = append(dest, &account.FName)
			case "sname":
				dest = append(dest, &account.SName)
			case "phone":
				dest = append(dest, &account.Phone)
			case "sex":
				dest = append(dest, &account.Sex)
			case "birth":
				dest = append(dest, &account.Birth)
			case "country":
				dest = append(dest, &account.Country)
			case "city":
				dest = append(dest, &account.City)
			case "status":
				dest = append(dest, &account.Status)
			case "premium":
				dest = append(dest, &account.Premium.Start)
				dest = append(dest, &account.Premium.Finish)
			}
		}
		err = rows.Scan(dest...)
		accounts.Accounts = append(accounts.Accounts, *account)
	}
	rows.Close()

	return accounts
}

func (db *DB) Group(query M) models.Groups {
	limit := query["limit"].(int)
	keys := query["keys"].([]string)
	order := query["order"].(int)
	groupsMap := make(map[models.Group]int)
MainLoop:
	for _, accountMin := range db.accountsMin {
	InnerLoop:
		for k, v := range query {
			switch k {
			case "sex":
				if db.sex[accountMin.Sex] != v.(string) {
					continue MainLoop
				}
			case "status":
				if db.status[accountMin.Status] != v.(string) {
					continue MainLoop
				}
			case "country":
				if db.countries[accountMin.Country] != v.(string) {
					continue MainLoop
				}
			case "city":
				if db.cities[accountMin.City] != v.(string) {
					continue MainLoop
				}
			case "birth":
				if accountMin.CheckBirth(v.(int)) {
					continue InnerLoop
				}
				continue MainLoop
			case "joined":
				if accountMin.CheckJoined(v.(int)) {
					continue InnerLoop
				}
				continue MainLoop
			case "interests":
				for iii := range accountMin.Interests {
					if db.interests[accountMin.Interests[iii]] == v.(string) {
						continue InnerLoop
					}
				}
				continue MainLoop
			case "likes":
				for iii := range accountMin.Likes {
					if accountMin.Likes[iii].ID == v.(int) {
						continue InnerLoop
					}
				}
				continue MainLoop
			}
		}

		unwind := false
		group := models.Group{}
		for i := range keys {
			switch keys[i] {
			case "sex":
				group.Sex = db.sex[accountMin.Sex]
			case "status":
				group.Status = db.status[accountMin.Status]
			case "interests":
				unwind = true
			case "country":
				group.Country = db.countries[accountMin.Country]
			case "city":
				group.City = db.cities[accountMin.City]
			}
		}

		if !unwind {
			groupsMap[group]++
		} else {
			for i := range accountMin.Interests {
				group.Interests = db.interests[accountMin.Interests[i]]
				groupsMap[group]++
			}
		}

	}

	result := models.Groups{}
	result.Groups = []models.Group{}

	for k, v := range groupsMap {
		k.Count = v
		result.Groups = append(result.Groups, k)
	}

	if order == 1 {
		sort.Slice(result.Groups, func(i, j int) bool {
			if result.Groups[i].Count == result.Groups[j].Count {
				for k := range keys {
					switch keys[k] {
					case "sex":
						if result.Groups[i].Sex == result.Groups[j].Sex {
							continue
						}
						return result.Groups[i].Sex < result.Groups[j].Sex
					case "status":
						if result.Groups[i].Status == result.Groups[j].Status {
							continue
						}
						return result.Groups[i].Status < result.Groups[j].Status
					case "interests":
						if result.Groups[i].Interests == result.Groups[j].Interests {
							continue
						}
						return result.Groups[i].Interests < result.Groups[j].Interests
					case "country":
						if result.Groups[i].Country == result.Groups[j].Country {
							continue
						}
						return result.Groups[i].Country < result.Groups[j].Country
					case "city":
						if result.Groups[i].City == result.Groups[j].City {
							continue
						}
						return result.Groups[i].City < result.Groups[j].City
					}
				}
			}
			return result.Groups[i].Count < result.Groups[j].Count
		})
	} else {
		sort.Slice(result.Groups, func(i, j int) bool {
			if result.Groups[i].Count == result.Groups[j].Count {
				for k := range keys {
					switch keys[k] {
					case "sex":
						if result.Groups[i].Sex == result.Groups[j].Sex {
							continue
						}
						return result.Groups[i].Sex > result.Groups[j].Sex
					case "status":
						if result.Groups[i].Status == result.Groups[j].Status {
							continue
						}
						return result.Groups[i].Status > result.Groups[j].Status
					case "interests":
						if result.Groups[i].Interests == result.Groups[j].Interests {
							continue
						}
						return result.Groups[i].Interests > result.Groups[j].Interests
					case "country":
						if result.Groups[i].Country == result.Groups[j].Country {
							continue
						}
						return result.Groups[i].Country > result.Groups[j].Country
					case "city":
						if result.Groups[i].City == result.Groups[j].City {
							continue
						}
						return result.Groups[i].City > result.Groups[j].City
					}
				}
			}
			return result.Groups[i].Count > result.Groups[j].Count
		})
	}

	if len(result.Groups) > limit {
		result.Groups = result.Groups[:limit]
	}

	return result
}

func (db *DB) Recommend(id, now int, query M) (models.Accounts, bool) {
	accounts := models.Accounts{}
	accounts.Accounts = []models.Account{}

	i := sort.Search(len(db.accountsMin), func(i int) bool {
		return db.accountsMin[i].ID <= id
	})

	var accountMinForFind models.AccountMin

	if i < len(db.accountsMin) && db.accountsMin[i].ID == id {
		accountMinForFind = db.accountsMin[i]
	} else {
		return accounts, false
	}

	var accountsMin []models.AccountMin

MainLoop:
	for _, accountMin := range db.accountsMin {
		if accountMin.Sex == accountMinForFind.Sex {
			continue MainLoop
		}

		for k, v := range query {
			switch k {
			case "city":
				if db.cities[accountMin.City] != v.(string) {
					continue MainLoop
				}
			case "country":
				if db.countries[accountMin.Country] != v.(string) {
					continue MainLoop
				}
			}
		}

		found := false
		for ii := range accountMinForFind.Interests {
			for iii := range accountMin.Interests {
				if accountMin.Interests[iii] == accountMinForFind.Interests[ii] {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			continue MainLoop
		}

		accountsMin = append(accountsMin, accountMin)
	}

	sort.Slice(accountsMin, func(i, j int) bool {
		compi := accountMinForFind.CheckCompatibility(accountsMin[i], now)
		compj := accountMinForFind.CheckCompatibility(accountsMin[j], now)
		if compi == compj {
			return accountsMin[i].ID < accountsMin[j].ID
		}
		return compi > compj
	})

	for i := 0; i < query["limit"].(int) && i < len(accountsMin); i++ {
		accounts.Accounts = append(accounts.Accounts, models.Account{
			ID:      accountsMin[i].ID,
			Email:   accountsMin[i].Email,
			Status:  db.status[accountsMin[i].Status],
			FName:   db.fnames[accountsMin[i].FName],
			SName:   db.snames[accountsMin[i].SName],
			Birth:   accountsMin[i].Birth,
			Premium: accountsMin[i].Premium,
		})
	}

	return accounts, true
}

func (db *DB) Suggest(id int, query M) (models.Accounts, bool) {
	accounts := models.Accounts{}
	accounts.Accounts = []models.Account{}

	i := sort.Search(len(db.accountsMin), func(i int) bool {
		return db.accountsMin[i].ID <= id
	})

	var accountMinForFind models.AccountMin

	if i < len(db.accountsMin) && db.accountsMin[i].ID == id {
		accountMinForFind = db.accountsMin[i]
	} else {
		return accounts, false
	}

	var accountsMin []models.AccountMin

MainLoop:
	for _, accountMin := range db.accountsMin {
		if accountMin.Sex != accountMinForFind.Sex {
			continue MainLoop
		}

		for k, v := range query {
			switch k {
			case "city":
				if db.cities[accountMin.City] != v.(string) {
					continue MainLoop
				}
			case "country":
				if db.countries[accountMin.Country] != v.(string) {
					continue MainLoop
				}
			}
		}

		found := false
		for ii := range accountMinForFind.Likes {
			for iii := range accountMin.Likes {
				if accountMin.Likes[iii].ID == accountMinForFind.Likes[ii].ID {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			continue MainLoop
		}

		accountsMin = append(accountsMin, accountMin)
	}

	accountMinForFind.PrepareLikesMap()

	parallelMergeSort(accountsMin, accountMinForFind)

	ids := make([]int, 0)
	for i := range accountsMin {
		ids = append(ids, accountMinForFind.GetNewIds(accountsMin[i])...)
		if len(ids) > query["limit"].(int) {
			ids = ids[:query["limit"].(int)]
			break
		}
	}

	for k := range ids {
		ii := sort.Search(len(db.accountsMin), func(i int) bool {
			return db.accountsMin[i].ID <= ids[k]
		})

		accounts.Accounts = append(accounts.Accounts, models.Account{
			ID:     db.accountsMin[ii].ID,
			Email:  db.accountsMin[ii].Email,
			Status: db.status[db.accountsMin[ii].Status],
			FName:  db.fnames[db.accountsMin[ii].FName],
			SName:  db.snames[db.accountsMin[ii].SName],
		})
	}

	return accounts, true
}
