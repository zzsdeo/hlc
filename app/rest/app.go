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

type filterResponse struct {
	Accounts []models.Account `json:"accounts"`
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

	//todo validate params

	queryMap := make(map[string]interface{})

	var limit int

	for k, v := range mux.Vars(r) {
		switch k {
		case "sex_eq":
			queryMap["sex"] = v
			continue
		case "email_domain":
			regex := make(map[string]string)
			regex["$regex"] = v
			queryMap["email"] = regex
			continue
		case "email_lt":
			lt := make(map[string]string)
			lt["$lt"] = v
			queryMap["email"] = lt
			continue
		case "email_gt":
			gt := make(map[string]string)
			gt["$gt"] = v
			queryMap["email"] = gt
			continue
		case "status_eq":
			queryMap["status"] = v
			continue
		case "status_neq":
			neq := make(map[string]string)
			neq["$neq"] = v
			queryMap["status"] = neq
			continue
		case "fname_eq":
			queryMap["fname"] = v
			continue
		case "fname_any":
			in := make(map[string][]string)
			in["$in"] = strings.Split(v, ",")
			queryMap["fname"] = in
			continue
		case "fname_null":
			queryMap["fname"] = exists(v)
			continue
		case "sname_eq":
			queryMap["sname"] = v
			continue
		case "sname_starts":
			regex := make(map[string]string)
			regex["$regex"] = "^" + v
			queryMap["sname"] = regex
			continue
		case "sname_null":
			queryMap["sname"] = exists(v)
			continue
		case "phone_code":
			regex := make(map[string]string)
			regex["$regex"] = "(" + v + ")"
			queryMap["phone"] = regex
			continue
		case "phone_null":
			queryMap["phone"] = exists(v)
			continue
		case "country_eq":
			queryMap["country"] = v
			continue
		case "country_null":
			queryMap["country"] = exists(v)
			continue
		case "city_eq":
			queryMap["city"] = v
			continue
		case "city_any":
			in := make(map[string][]string)
			in["$in"] = strings.Split(v, ",")
			queryMap["city"] = in
			continue
		case "city_null":
			queryMap["city"] = exists(v)
			continue
		case "birth_lt":
			lt := make(map[string]int)
			lt["$lt"], _ = strconv.Atoi(v)
			queryMap["birth"] = lt
			continue
		case "birth_gt":
			gt := make(map[string]int)
			gt["$gt"], _ = strconv.Atoi(v)
			queryMap["birth"] = gt
			continue
		case "birth_year":
			year, _ := strconv.Atoi(v)
			interval := make(map[string]int64)
			interval["$gte"] = time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
			interval["$lt"] = time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
			queryMap["birth"] = interval
			continue
		case "interests_contains":
			all := make(map[string][]string)
			all["$all"] = strings.Split(v, ",")
			queryMap["interests"] = all
			continue
		case "interests_any":
			elemMatch := make(map[string]map[string][]string)
			in := make(map[string][]string)
			in["$in"] = strings.Split(v, ",")
			elemMatch["$elemMatch"] = in
			queryMap["interests"] = elemMatch
			continue
		case "likes_contains":
			//mongo find {likes: {id: {$in: [1, 2, 3]}}}
			likes := strings.Split(v, ",")
			likeIds := make([]int, len(likes))
			for _, like := range likes {
				l, _ := strconv.Atoi(like)
				likeIds = append(likeIds, l)
			}
			in := make(map[string][]int)
			in["$in"] = likeIds
			like := make(map[string]map[string][]int)
			like["id"] = in
			queryMap["likes"] = like
			continue
		case "premium_now":
			//mongo find {premium: {$and: [{start: {$lt: 123}}, {finish: {$gt: 123}}]}}
			lt := make(map[string]int)
			lt["$lt"] = a.now
			gt := make(map[string]int)
			gt["$gt"] = a.now
			start := make(map[string]map[string]int)
			start["start"] = lt
			finish := make(map[string]map[string]int)
			finish["finish"] = gt
			and := make(map[string][]map[string]map[string]int)
			interval := make([]map[string]map[string]int, 2)
			interval = append(interval, start, finish)
			and["$and"] = interval
			queryMap["premium"] = and
			continue
		case "premium_null":
			queryMap["premium"] = exists(v)
			continue
		case "limit":
			limit, _ = strconv.Atoi(v)
		}
	}

	session := a.mongoSession.Copy()
	defer session.Close()
	collection := session.DB(dbName).C(accountsCollectionName)

	selector := make(map[string]int)

	for k, _ := range queryMap {
		selector[k] = 1
	}

	filterResponse := filterResponse{}

	err := collection.Find(queryMap).Limit(limit).Sort("id").Select(selector).All(&filterResponse.Accounts)

	if err != nil {
		log.Println("[ERROR] ", err)
	}

	err = json.NewEncoder(w).Encode(filterResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", err)
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
