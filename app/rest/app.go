package rest

import (
	"encoding/json"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
	"hlc/app/models"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	dbName                 = "hlc"
	accountsCollectionName = "accounts"
)

type App struct {
	router       *mux.Router
	mongoSession *mgo.Session
	now          int //current time from options.txt
}

func (a *App) Initialize(mongoAddr string) {
	a.router = mux.NewRouter()

	session, err := mgo.Dial(mongoAddr)
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}
	a.mongoSession = session

	//todo make indexes

	a.initializeRoutes()
}

func (a *App) SetNow(now int) {
	a.now = now
}

func (a *App) DropCollection() {
	session := a.mongoSession.Copy()
	defer session.Close()
	collection := session.DB(dbName).C(accountsCollectionName)
	err := collection.DropCollection()
	if err != nil {
		log.Println("[ERROR] ", err)
	}
}

func (a *App) LoadData(accounts []models.Account) {
	session := a.mongoSession.Copy()
	defer session.Close()
	collection := session.DB(dbName).C(accountsCollectionName)

	for i, account := range accounts {
		err := collection.Insert(&account)
		if err != nil {
			log.Println("[ERROR] index=", i, err)
		}
	}
	log.Println("[INFO] all accounts added")
}

func (a *App) CheckDB() {
	session := a.mongoSession.Copy()
	defer session.Close()
	collection := session.DB(dbName).C(accountsCollectionName)
	recs, err := collection.Find(nil).Count()
	if err != nil {
		log.Println("[ERROR] ", err)
	}
	log.Println("[INFO] recs added=", recs)
}

func (a *App) CreateIndex(background bool) {
	log.Println("[INFO] indexing started")

	session := a.mongoSession.Copy()
	defer session.Close()
	collection := session.DB(dbName).C(accountsCollectionName)

	err := collection.EnsureIndex(mgo.Index{
		Key:        []string{"id"},
		Background: background,
		Sparse:     true,
	})

	if err != nil {
		log.Println("[ERROR] ", err)
	}

	err = collection.EnsureIndex(mgo.Index{
		Key:        []string{"country"},
		Background: background,
		Sparse:     true,
	})

	if err != nil {
		log.Println("[ERROR] ", err)
	}

	err = collection.EnsureIndex(mgo.Index{
		Key:        []string{"city"},
		Background: background,
		Sparse:     true,
	})

	if err != nil {
		log.Println("[ERROR] ", err)
	}

	err = collection.EnsureIndex(mgo.Index{
		Key:        []string{"birth"},
		Background: background,
		Sparse:     true,
	})

	if err != nil {
		log.Println("[ERROR] ", err)
	}

	err = collection.EnsureIndex(mgo.Index{
		Key:        []string{"interests"},
		Background: background,
		Sparse:     true,
	})

	if err != nil {
		log.Println("[ERROR] ", err)
	}

	err = collection.EnsureIndex(mgo.Index{
		Key:        []string{"likes"},
		Background: background,
		Sparse:     true,
	})

	if err != nil {
		log.Println("[ERROR] ", err)
	}

	err = collection.EnsureIndex(mgo.Index{
		Key:        []string{"joined"},
		Background: background,
		Sparse:     true,
	})

	if err != nil {
		log.Println("[ERROR] ", err)
	}

	if !background {
		log.Println("[INFO] indexing finished")
	}
}

func (a *App) Run(listenAddr string) {
	log.Println("[INFO] start server on", listenAddr)
	log.Fatal("[ERROR] ", http.ListenAndServe(listenAddr, a.router))
}

func (a *App) initializeRoutes() {
	a.router.HandleFunc("/ping/", a.ping).Methods(http.MethodGet)

	a.router.HandleFunc("/accounts/filter/", a.filter).Methods(http.MethodGet)
	a.router.HandleFunc("/accounts/group/", a.group).Methods(http.MethodGet)
	a.router.HandleFunc("/accounts/{id}/recommend/", a.recommend).Methods(http.MethodGet)
}

func (a *App) ping(w http.ResponseWriter, r *http.Request) {
	_, err := io.WriteString(w, "pong")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, err)
	}
}

func (a *App) filter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := bson.M{}
	var limit int
	for k, v := range r.URL.Query() {
		if v[0] == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch k {
		case "sex_eq":
			query["sex"] = v[0]
			continue
		case "email_domain":
			query["email"] = bson.M{"$regex": "(@" + v[0] + ")"}
			continue
		case "email_lt":
			query["email"] = bson.M{"$lt": v[0]}
			continue
		case "email_gt":
			query["email"] = bson.M{"$gt": v[0]}
			continue
		case "status_eq":
			query["status"] = v[0]
			continue
		case "status_neq":
			query["status"] = bson.M{"$ne": v[0]}
			continue
		case "fname_eq":
			query["fname"] = v[0]
			continue
		case "fname_any":
			query["fname"] = bson.M{"$in": strings.Split(v[0], ",")}
			continue
		case "fname_null":
			query["fname"] = exists(v[0])
			continue
		case "sname_eq":
			query["sname"] = v[0]
			continue
		case "sname_starts":
			query["sname"] = bson.M{"$regex": "^" + v[0]}
			continue
		case "sname_null":
			query["sname"] = exists(v[0])
			continue
		case "phone_code":
			query["phone"] = bson.M{"$regex": "(\\(" + v[0] + "\\))"}
			continue
		case "phone_null":
			query["phone"] = exists(v[0])
			continue
		case "country_eq":
			query["country"] = v[0]
			continue
		case "country_null":
			query["country"] = exists(v[0])
			continue
		case "city_eq":
			query["city"] = v[0]
			continue
		case "city_any":
			query["city"] = bson.M{"$in": strings.Split(v[0], ",")}
			continue
		case "city_null":
			query["city"] = exists(v[0])
			continue
		case "birth_lt":
			birth, err := strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			query["birth"] = bson.M{"$lt": birth}
			continue
		case "birth_gt":
			birth, err := strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			query["birth"] = bson.M{"$gt": birth}
			continue
		case "birth_year":
			year, err := strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			query["birth"] = yearInterval(year)
			continue
		case "interests_contains":
			query["interests"] = bson.M{"$all": strings.Split(v[0], ",")}
			continue
		case "interests_any":
			query["interests"] = bson.M{"$elemMatch": bson.M{"$in": strings.Split(v[0], ",")}}
			continue
		case "likes_contains":
			likes := strings.Split(v[0], ",")
			likeIds := make([]int, 0)
			for _, like := range likes {
				l, err := strconv.Atoi(like)
				if err != nil {
					log.Println("[ERROR] ", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				likeIds = append(likeIds, l)
			}
			//log.Println("[DEBUG] ", likeIds)
			query["likes"] = bson.M{"$elemMatch": bson.M{"id": bson.M{"$all": likeIds}}}
			continue
		case "premium_now":
			query["$and"] = []bson.M{{"premium.start": bson.M{"$lt": a.now}}, {"premium.finish": bson.M{"$gt": a.now}}}
			continue
		case "premium_null":
			query["premium"] = exists(v[0])
			continue
		case "limit":
			var err error
			limit, err = strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if limit < 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			continue
		case "query_id":
			continue
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	//log.Println("[DEBUG] query=", query)
	//log.Println("[DEBUG] limit=", limit)

	session := a.mongoSession.Copy()
	defer session.Close()
	collection := session.DB(dbName).C(accountsCollectionName)

	selector := make(map[string]int)
	selector["id"] = 1
	selector["email"] = 1
	for k, _ := range query {
		if k != "interests" && k != "likes" {
			if k == "$and" {
				selector["premium"] = 1
			} else {
				selector[k] = 1
			}
		}
	}

	//log.Println("[DEBUG] selector=", selector)

	filterResponse := models.Accounts{}
	filterResponse.Accounts = make([]models.Account, 0)

	err := collection.Find(query).Limit(limit).Sort("-id").Select(selector).All(&filterResponse.Accounts)

	if err != nil {
		log.Println("[ERROR] ", err, query)
	}

	err = json.NewEncoder(w).Encode(filterResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", err)
	}

}

func (a *App) group(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := bson.M{}
	var limit, order int
	keys := make([]string, 0)
	for k, v := range r.URL.Query() {
		if v[0] == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch k {
		case "sex":
			query["sex"] = v[0]
			continue
		case "birth":
			year, err := strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			query["birth"] = yearInterval(year)
			continue
		case "country":
			query["country"] = v[0]
			continue
		case "city":
			query["city"] = v[0]
			continue
		case "joined":
			year, err := strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			query["joined"] = yearInterval(year)
			continue
		case "status":
			query["status"] = v[0]
			continue
		case "interests":
			query["interests"] = bson.M{"$elemMatch": bson.M{"$eq": v[0]}}
			continue
		case "likes":
			likeId, err := strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			query["likes"] = bson.M{"$elemMatch": bson.M{"id": likeId}}
			continue
		case "limit":
			var err error
			limit, err = strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if limit < 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			continue
		case "order":
			var err error
			order, err = strconv.Atoi(v[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if order == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if order != -1 && order != 1 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			continue
		case "keys":
			keys = strings.Split(v[0], ",")
			if len(keys) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			//validate keys
			for _, key := range keys {
				if _, ok := models.Keys[key]; !ok {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			continue
		case "query_id":
			continue
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	session := a.mongoSession.Copy()
	defer session.Close()
	collection := session.DB(dbName).C(accountsCollectionName)

	groups := models.Groups{}
	groups.Groups = make([]models.Group, 0)

	groupPipe := bson.M{}
	projectPipe := bson.M{"_id": 0, "count": 1}
	sortPipe := bson.M{"count": order}
	unwind := false
	for _, key := range keys {
		groupPipe[key] = "$" + key
		projectPipe[key] = "$_id." + key
		sortPipe[key] = order
		if key == "interests" {
			unwind = true
		}
	}

	var pipeline []bson.M
	if unwind {
		pipeline = []bson.M{
			{"$match": query},
			{"$unwind": "$interests"},
			{"$group": bson.M{"_id": groupPipe, "count": bson.M{"$sum": 1}}},
			{"$project": projectPipe},
			{"$sort": sortPipe},
			{"$limit": limit},
		}
	} else {
		pipeline = []bson.M{
			{"$match": query},
			{"$group": bson.M{"_id": groupPipe, "count": bson.M{"$sum": 1}}},
			{"$project": projectPipe},
			{"$sort": sortPipe},
			{"$limit": limit},
		}
	}

	err := collection.Pipe(pipeline).All(&groups.Groups)
	if err != nil {
		log.Println("[ERROR] ", err)
	}

	err = json.NewEncoder(w).Encode(groups)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", err)
	}
}

func (a *App) recommend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	session := a.mongoSession.Copy()
	defer session.Close()
	collection := session.DB(dbName).C(accountsCollectionName)

	account := models.Account{}
	err = collection.Find(bson.M{"id": id}).Select(bson.M{
		"sex":       1,
		"birth":     1,
		"interests": 1}).One(&account)
	if err != nil {
		log.Println("[ERROR] ", err)
		if err == mgo.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	query := bson.M{"sex": "f"}
	if account.Sex == "f" {
		query["sex"] = "m"
	}

	query["interests"] = bson.M{"$elemMatch": bson.M{"$in": account.Interests}}

	var limit int
	for k, v := range r.URL.Query() {
		if v[0] == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch k {
		case "country":
			query["country"] = v[0]
			continue
		case "city":
			query["city"] = v[0]
			continue
		case "limit":
			var err error
			limit, err = strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if limit < 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			continue
		case "query_id":
			continue
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	accounts := models.Accounts{}
	accounts.Accounts = make([]models.Account, 0)

	err = collection.Find(query).Select(bson.M{
		"id":        1,
		"email":     1,
		"status":    1,
		"fname":     1,
		"sname":     1,
		"birth":     1,
		"premium":   1,
		"interests": 1}).All(&accounts.Accounts)
	if err != nil {
		log.Println("[ERROR] ", err)
	}

	account.PrepareInterestsMap()
	sort.Slice(accounts.Accounts, func(i, j int) bool {
		return account.CheckCompatibility(accounts.Accounts[i], a.now) > account.CheckCompatibility(accounts.Accounts[j], a.now)
	})

	if len(accounts.Accounts) > limit {
		accounts.Accounts = accounts.Accounts[:limit]
	}

	for i, _ := range accounts.Accounts {
		accounts.Accounts[i].Interests = []string{}
	}

	err = json.NewEncoder(w).Encode(accounts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", err)
	}
}

func exists(v string) bson.M {
	switch v {
	case "0":
		return bson.M{"$exists": true}
	case "1":
		return bson.M{"$exists": false}
	}
	return nil
}

func yearInterval(year int) bson.M {
	return bson.M{
		"$gte": time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC).Unix(),
		"$lt":  time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
}
