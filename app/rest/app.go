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
		account.MongoID = bson.NewObjectId()
		err := collection.Insert(&account)
		if err != nil {
			log.Println("[ERROR] index=", i, err)
		}
	}
	log.Println("[INFO] all accounts added")
}

func (a *App) Run(listenAddr string) {
	log.Println("[INFO] start server on", listenAddr)
	log.Fatal("[ERROR] ", http.ListenAndServe(listenAddr, a.router))
}

func (a *App) initializeRoutes() {
	a.router.HandleFunc("/ping/", a.ping).Methods(http.MethodGet)

	a.router.HandleFunc("/accounts/filter/", a.filter).Methods(http.MethodGet)
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

	queryMap := make(map[string]interface{})

	var limit int

	for k, v := range r.URL.Query() {
		switch k {
		case "sex_eq":
			queryMap["sex"] = v[0]
			continue
		case "email_domain":
			regex := make(map[string]string)
			regex["$regex"] = "(@" + v[0] + ")"
			queryMap["email"] = regex
			continue
		case "email_lt":
			lt := make(map[string]string)
			lt["$lt"] = v[0]
			queryMap["email"] = lt
			continue
		case "email_gt":
			gt := make(map[string]string)
			gt["$gt"] = v[0]
			queryMap["email"] = gt
			continue
		case "status_eq":
			queryMap["status"] = v[0]
			continue
		case "status_neq":
			ne := make(map[string]string)
			ne["$ne"] = v[0]
			queryMap["status"] = ne
			continue
		case "fname_eq":
			queryMap["fname"] = v[0]
			continue
		case "fname_any":
			in := make(map[string][]string)
			in["$in"] = strings.Split(v[0], ",")
			queryMap["fname"] = in
			continue
		case "fname_null":
			queryMap["fname"] = exists(v[0])
			continue
		case "sname_eq":
			queryMap["sname"] = v[0]
			continue
		case "sname_starts":
			regex := make(map[string]string)
			regex["$regex"] = "^" + v[0]
			queryMap["sname"] = regex
			continue
		case "sname_null":
			queryMap["sname"] = exists(v[0])
			continue
		case "phone_code":
			regex := make(map[string]string)
			regex["$regex"] = "(\\(" + v[0] + "\\))"
			queryMap["phone"] = regex
			continue
		case "phone_null":
			queryMap["phone"] = exists(v[0])
			continue
		case "country_eq":
			queryMap["country"] = v[0]
			continue
		case "country_null":
			queryMap["country"] = exists(v[0])
			continue
		case "city_eq":
			queryMap["city"] = v[0]
			continue
		case "city_any":
			in := make(map[string][]string)
			in["$in"] = strings.Split(v[0], ",")
			queryMap["city"] = in
			continue
		case "city_null":
			queryMap["city"] = exists(v[0])
			continue
		case "birth_lt":
			lt := make(map[string]int)
			var err error
			lt["$lt"], err = strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			queryMap["birth"] = lt
			continue
		case "birth_gt":
			gt := make(map[string]int)
			var err error
			gt["$gt"], err = strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			queryMap["birth"] = gt
			continue
		case "birth_year":
			year, err := strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			interval := make(map[string]int64)
			interval["$gte"] = time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
			interval["$lt"] = time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
			queryMap["birth"] = interval
			continue
		case "interests_contains":
			all := make(map[string][]string)
			all["$all"] = strings.Split(v[0], ",")
			queryMap["interests"] = all
			continue
		case "interests_any":
			elemMatch := make(map[string]map[string][]string)
			in := make(map[string][]string)
			in["$in"] = strings.Split(v[0], ",")
			elemMatch["$elemMatch"] = in
			queryMap["interests"] = elemMatch
			continue
		case "likes_contains":
			likes := strings.Split(v[0], ",")
			likeIds := make([]int, 0)
			for _, like := range likes {
				l, err := strconv.Atoi(like)
				if err != nil { //todo check it
					log.Println("[ERROR] ", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				likeIds = append(likeIds, l)
			}
			//log.Println("[DEBUG] ", likeIds)
			all := make(map[string][]int)
			all["$all"] = likeIds
			like := make(map[string]map[string][]int)
			like["id"] = all
			elemMatch := make(map[string]map[string]map[string][]int)
			elemMatch["$elemMatch"] = like
			queryMap["likes"] = elemMatch
			continue
		case "premium_now":
			//mongo find {"$and":["premium.start":{"$lt": 123}, "premium.finish":{"$gt": 123}]}
			lt := make(map[string]int)
			lt["$lt"] = a.now
			gt := make(map[string]int)
			gt["$gt"] = a.now
			start := make(map[string]map[string]int)
			start["premium.start"] = lt
			finish := make(map[string]map[string]int)
			finish["premium.finish"] = gt
			interval := make([]map[string]map[string]int, 0)
			interval = append(interval, start, finish)
			queryMap["$and"] = interval
			continue
		case "premium_null":
			queryMap["premium"] = exists(v[0])
			continue
		case "limit":
			var err error
			limit, err = strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case "query_id":
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	//log.Println("[DEBUG] queryMap=", queryMap)
	//log.Println("[DEBUG] limit=", limit)

	session := a.mongoSession.Copy()
	defer session.Close()
	collection := session.DB(dbName).C(accountsCollectionName)

	selector := make(map[string]int)
	selector["id"] = 1
	selector["email"] = 1
	for k, _ := range queryMap {
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

	err := collection.Find(queryMap).Limit(limit).Sort("-id").Select(selector).All(&filterResponse.Accounts)

	if err != nil {
		log.Println("[ERROR] ", err, queryMap)
	}

	err = json.NewEncoder(w).Encode(filterResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", err)
	}

}

func (a *App) group(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	queryMap := make(map[string]interface{})

	var limit int

	for k, v := range r.URL.Query() {
		switch k {
		case "id": //todo
		case "email":
			queryMap["email"] = v[0]
		case "fname":
			queryMap["fname"] = v[0]
		case "sname":
			queryMap["sname"] = v[0]
		case "phone":
			queryMap["phone"] = v[0]
		case "sex":
			queryMap["sex"] = v[0]
		case "birth":
			year, err := strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			interval := make(map[string]int64)
			interval["$gte"] = time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
			interval["$lt"] = time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
			queryMap["birth"] = interval
		case "country":
			queryMap["country"] = v[0]
		case "city":
			queryMap["city"] = v[0]
		case "joined":
			queryMap["joined"] = v[0]
		case "status":
			queryMap["status"] = v[0]
		case "interests":
			queryMap["interests"] = v[0]
		case "premium":
			queryMap["premium"] = v[0]
		case "likes":
			queryMap["likes"] = v[0]
		case "limit":
			var err error
			limit, err = strconv.Atoi(v[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case "query_id":
		default:
			w.WriteHeader(http.StatusBadRequest)
			return

		}

	}
}

func exists(v string) map[string]bool {
	exists := make(map[string]bool)
	switch v {
	case "0":
		exists["$exists"] = true
		return exists
	case "1":
		exists["$exists"] = false
		return exists
	}
	return nil
}

func xz(v string) {

}
