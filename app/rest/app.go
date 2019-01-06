package rest

import (
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
	"hlc/app/models"
	"hlc/app/store"
	"hlc/app/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type App struct {
	router *mux.Router
	db     *store.DB
	now    int //current time from options.txt
}

func (a *App) Initialize() {
	a.router = mux.NewRouter()
	a.db = store.NewDB()
	a.initializeRoutes()
}

func (a *App) SetNow(now int) {
	a.now = now
}

func (a *App) LoadData(accounts []models.Account) {
	a.db.LoadData(accounts)
	log.Println("[INFO] all accounts added")
}

func (a *App) CheckDB() {
	recs := a.db.Count()
	log.Println("[INFO] recs added=", recs)
}

func (a *App) CreateIndexes(background bool) {
	log.Println("[INFO] indexing started")
	a.db.CreateIndexes()
	if !background {
		log.Println("[INFO] indexing finished")
	}
}

func (a *App) Run(listenAddr string) {
	log.Println("[INFO] start server on", listenAddr)
	log.Fatal("[ERROR] ", http.ListenAndServe(listenAddr, a.router))
}

func (a *App) initializeRoutes() {
	a.router.HandleFunc("/accounts/filter/", a.filter).Methods(http.MethodGet)
	//a.router.HandleFunc("/accounts/group/", a.group).Methods(http.MethodGet)
	//a.router.HandleFunc("/accounts/{id}/recommend/", a.recommend).Methods(http.MethodGet)
	//a.router.HandleFunc("/accounts/{id}/suggest/", a.suggest).Methods(http.MethodGet)

	// Регистрация pprof-обработчиков
	//a.router.HandleFunc("/debug/pprof/", pprof.Index)
	//a.router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	//a.router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	//a.router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	//a.router.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func (a *App) filter(w http.ResponseWriter, r *http.Request) {
	defer utils.TimeTrack(time.Now(), "request")
	//w.Header().Set("Content-Type", "application/json")
	query := store.M{}
	for k, v := range r.URL.Query() {
		if v[0] == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch k {
		case "sex_eq":
			query["sex_eq"] = v[0]
			continue
		case "email_domain":
			query["email_domain"] = v[0]
			continue
		case "email_lt":
			query["email_lt"] = v[0]
			continue
		case "email_gt":
			query["email_gt"] = v[0]
			continue
		case "status_eq":
			query["status_eq"] = v[0]
			continue
		case "status_neq":
			query["status_neq"] = v[0]
			continue
		case "fname_eq":
			query["fname_eq"] = v[0]
			continue
		case "fname_any":
			query["fname_any"] = strings.Split(v[0], ",")
			continue
		case "fname_null":
			query["fname_null"] = v[0]
			continue
		case "sname_eq":
			query["sname_eq"] = v[0]
			continue
		case "sname_starts":
			query["sname_starts"] = v[0]
			continue
		case "sname_null":
			query["sname_null"] = v[0]
			continue
		case "phone_code":
			query["phone_code"] = v[0]
			continue
		case "phone_null":
			query["phone_null"] = v[0]
			continue
		case "country_eq":
			query["country_eq"] = v[0]
			continue
		case "country_null":
			query["country_null"] = v[0]
			continue
		case "city_eq":
			query["city_eq"] = v[0]
			continue
		case "city_any":
			query["city_any"] = strings.Split(v[0], ",")
			continue
		case "city_null":
			query["city_null"] = v[0]
			continue
		case "birth_lt":
			birth, err := strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			query["birth_lt"] = birth
			continue
		case "birth_gt":
			birth, err := strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			query["birth_gt"] = birth
			continue
		case "birth_year":
			year, err := strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			query["birth_year"] = year
			continue
		case "interests_contains":
			query["interests_contains"] = strings.Split(v[0], ",")
			continue
		case "interests_any":
			query["interests_any"] = strings.Split(v[0], ",")
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
			query["likes_contains"] = likeIds
			continue
		case "premium_now":
			query["premium_now"] = a.now
			continue
		case "premium_null":
			query["premium_null"] = v[0]
			continue
		case "limit":
			limit, err := strconv.Atoi(v[0])
			if err != nil {
				log.Println("[ERROR] ", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if limit < 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			query["limit"] = limit
			continue
		case "query_id":
			continue
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	//log.Println("[DEBUG] query=", query)

	accounts := a.db.Find(query)
	_, _, err := easyjson.MarshalToHTTPResponseWriter(accounts, w)
	//err := json.NewEncoder(w).Encode(a.db.Find(query))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", err)
	}
}

//func (a *App) group(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Content-Type", "application/json")
//	query := bson.M{}
//	var limit, order int
//	keys := make([]string, 0)
//	for k, v := range r.URL.Query() {
//		if v[0] == "" {
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//		switch k {
//		case "sex":
//			query["sex"] = v[0]
//			continue
//		case "birth":
//			year, err := strconv.Atoi(v[0])
//			if err != nil {
//				log.Println("[ERROR] ", err)
//				w.WriteHeader(http.StatusBadRequest)
//				return
//			}
//			query["birth"] = yearInterval(year)
//			continue
//		case "country":
//			query["country"] = v[0]
//			continue
//		case "city":
//			query["city"] = v[0]
//			continue
//		case "joined":
//			year, err := strconv.Atoi(v[0])
//			if err != nil {
//				log.Println("[ERROR] ", err)
//				w.WriteHeader(http.StatusBadRequest)
//				return
//			}
//			query["joined"] = yearInterval(year)
//			continue
//		case "status":
//			query["status"] = v[0]
//			continue
//		case "interests":
//			query["interests"] = bson.M{"$elemMatch": bson.M{"$eq": v[0]}}
//			continue
//		case "likes":
//			likeId, err := strconv.Atoi(v[0])
//			if err != nil {
//				log.Println("[ERROR] ", err)
//				w.WriteHeader(http.StatusBadRequest)
//				return
//			}
//			query["likes"] = bson.M{"$elemMatch": bson.M{"id": likeId}}
//			continue
//		case "limit":
//			var err error
//			limit, err = strconv.Atoi(v[0])
//			if err != nil {
//				log.Println("[ERROR] ", err)
//				w.WriteHeader(http.StatusBadRequest)
//				return
//			}
//			if limit < 0 {
//				w.WriteHeader(http.StatusBadRequest)
//				return
//			}
//			continue
//		case "order":
//			var err error
//			order, err = strconv.Atoi(v[0])
//			if err != nil {
//				w.WriteHeader(http.StatusBadRequest)
//				return
//			}
//
//			if order == 0 {
//				w.WriteHeader(http.StatusBadRequest)
//				return
//			}
//
//			if order != -1 && order != 1 {
//				w.WriteHeader(http.StatusBadRequest)
//				return
//			}
//			continue
//		case "keys":
//			keys = strings.Split(v[0], ",")
//			if len(keys) == 0 {
//				w.WriteHeader(http.StatusBadRequest)
//				return
//			}
//			//validate keys
//			for _, key := range keys {
//				if _, ok := models.Keys[key]; !ok {
//					w.WriteHeader(http.StatusBadRequest)
//					return
//				}
//			}
//			continue
//		case "query_id":
//			continue
//		default:
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//	}
//
//	session := a.mongoSession.Copy()
//	defer session.Close()
//	collection := session.DB(dbName).C(accountsCollectionName)
//
//	groups := models.Groups{}
//	groups.Groups = make([]models.Group, 0)
//
//	groupPipe := bson.M{}
//	projectPipe := bson.M{"_id": 0, "count": 1}
//	sortPipe := bson.M{"count": order}
//	unwind := false
//	for _, key := range keys {
//		groupPipe[key] = "$" + key
//		projectPipe[key] = "$_id." + key
//		sortPipe[key] = order
//		if key == "interests" {
//			unwind = true
//		}
//	}
//
//	var pipeline []bson.M
//	if unwind {
//		pipeline = []bson.M{
//			{"$match": query},
//			{"$unwind": "$interests"},
//			{"$group": bson.M{"_id": groupPipe, "count": bson.M{"$sum": 1}}},
//			{"$project": projectPipe},
//			{"$sort": sortPipe},
//			{"$limit": limit},
//		}
//	} else {
//		pipeline = []bson.M{
//			{"$match": query},
//			{"$group": bson.M{"_id": groupPipe, "count": bson.M{"$sum": 1}}},
//			{"$project": projectPipe},
//			{"$sort": sortPipe},
//			{"$limit": limit},
//		}
//	}
//
//	err := collection.Pipe(pipeline).All(&groups.Groups)
//	if err != nil {
//		log.Println("[ERROR] ", err)
//	}
//
//	err = json.NewEncoder(w).Encode(groups)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		log.Println("[ERROR] ", err)
//	}
//}
//
//func (a *App) recommend(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Content-Type", "application/json")
//
//	id, err := strconv.Atoi(mux.Vars(r)["id"])
//	if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		return
//	}
//
//	session := a.mongoSession.Copy()
//	defer session.Close()
//	collection := session.DB(dbName).C(accountsCollectionName)
//
//	account := models.Account{}
//	err = collection.Find(bson.M{"id": id}).Select(bson.M{
//		"sex":       1,
//		"birth":     1,
//		"interests": 1}).One(&account)
//	if err != nil {
//		log.Println("[ERROR] ", err)
//		if err == mgo.ErrNotFound {
//			w.WriteHeader(http.StatusNotFound)
//			return
//		}
//	}
//
//	query := bson.M{"sex": "f"}
//	if account.Sex == "f" {
//		query["sex"] = "m"
//	}
//
//	query["interests"] = bson.M{"$elemMatch": bson.M{"$in": account.Interests}}
//
//	var limit int
//	for k, v := range r.URL.Query() {
//		if v[0] == "" {
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//		switch k {
//		case "country":
//			query["country"] = v[0]
//			continue
//		case "city":
//			query["city"] = v[0]
//			continue
//		case "limit":
//			var err error
//			limit, err = strconv.Atoi(v[0])
//			if err != nil {
//				log.Println("[ERROR] ", err)
//				w.WriteHeader(http.StatusBadRequest)
//				return
//			}
//			if limit < 0 {
//				w.WriteHeader(http.StatusBadRequest)
//				return
//			}
//			continue
//		case "query_id":
//			continue
//		default:
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//	}
//
//	accounts := models.Accounts{}
//	accounts.Accounts = make([]models.Account, 0)
//
//	err = collection.Find(query).Select(bson.M{
//		"id":        1,
//		"email":     1,
//		"status":    1,
//		"fname":     1,
//		"sname":     1,
//		"birth":     1,
//		"premium":   1,
//		"interests": 1}).All(&accounts.Accounts)
//	if err != nil {
//		log.Println("[ERROR] ", err)
//	}
//
//	account.PrepareInterestsMap()
//	sort.Slice(accounts.Accounts, func(i, j int) bool {
//		return account.CheckCompatibility(accounts.Accounts[i], a.now) > account.CheckCompatibility(accounts.Accounts[j], a.now)
//	})
//
//	if len(accounts.Accounts) > limit {
//		accounts.Accounts = accounts.Accounts[:limit]
//	}
//
//	for i, _ := range accounts.Accounts {
//		accounts.Accounts[i].Interests = []string{}
//	}
//
//	err = json.NewEncoder(w).Encode(accounts)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		log.Println("[ERROR] ", err)
//	}
//}
//
//func (a *App) suggest(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Content-Type", "application/json")
//
//	id, err := strconv.Atoi(mux.Vars(r)["id"])
//	if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		return
//	}
//
//	session := a.mongoSession.Copy()
//	defer session.Close()
//	collection := session.DB(dbName).C(accountsCollectionName)
//
//	account := models.Account{}
//	err = collection.Find(bson.M{"id": id}).Select(bson.M{
//		"sex":   1,
//		"likes": 1}).One(&account)
//	if err != nil {
//		log.Println("[ERROR] ", err)
//		if err == mgo.ErrNotFound {
//			w.WriteHeader(http.StatusNotFound)
//			return
//		}
//	}
//
//	query := bson.M{"sex": account.Sex}
//
//	var limit int
//	for k, v := range r.URL.Query() {
//		if v[0] == "" {
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//		switch k {
//		case "country":
//			query["country"] = v[0]
//			continue
//		case "city":
//			query["city"] = v[0]
//			continue
//		case "limit":
//			var err error
//			limit, err = strconv.Atoi(v[0])
//			if err != nil {
//				log.Println("[ERROR] ", err)
//				w.WriteHeader(http.StatusBadRequest)
//				return
//			}
//			if limit < 0 {
//				w.WriteHeader(http.StatusBadRequest)
//				return
//			}
//			continue
//		case "query_id":
//			continue
//		default:
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//	}
//
//	likeIds := make([]int, 0)
//	for _, like := range account.Likes {
//		likeIds = append(likeIds, like.ID)
//	}
//	query["likes"] = bson.M{"$elemMatch": bson.M{"id": bson.M{"$in": likeIds}}}
//
//	accounts := models.Accounts{}
//	accounts.Accounts = make([]models.Account, 0)
//
//	err = collection.Find(query).Select(bson.M{"likes": 1}).All(&accounts.Accounts)
//	if err != nil {
//		log.Println("[ERROR] ", err)
//	}
//
//	account.PrepareLikesMap()
//	//sort.Slice(accounts.Accounts, func(i, j int) bool {
//	//	return account.CheckSimilarity(accounts.Accounts[i]) > account.CheckSimilarity(accounts.Accounts[j])
//	//})
//	parallelMergeSort(accounts.Accounts, account)
//
//	//ids := make([]int, 0)
//	//for _, a := range accounts.Accounts {
//	//	//log.Println(account.CheckSimilarity(a))todo
//	//	if account.CheckSimilarity(a) == 0 {
//	//		break
//	//	}
//	//	ids = append(ids, account.GetNewIds(a)...)
//	//}
//	//
//	//err = collection.Find(bson.M{"id": bson.M{"$in":ids}}).Select(bson.M{
//	//	"id":1,
//	//	"email":1,
//	//	"status":1,
//	//	"fname":1,
//	//	"sname":1}).All(&accounts.Accounts)
//	//if err != nil {
//	//	log.Println("[ERROR] ", err)
//	//}
//	//
//	//result := models.Accounts{}
//	//result.Accounts = make([]models.Account, 0)
//	//for i:= 0; i < limit; i++ {
//	//	tempAccount, err := accounts.ExtractAccountByID(ids[i])
//	//	if err != nil {
//	//		log.Println("[ERROR] ", err)
//	//		continue
//	//	}
//	//	result.Accounts = append(result.Accounts, tempAccount)
//	//}
//	//
//	//err = json.NewEncoder(w).Encode(result)
//	//if err != nil {
//	//	w.WriteHeader(http.StatusInternalServerError)
//	//	log.Println("[ERROR] ", err)
//	//}
//
//	//if len(accounts.Accounts) > limit {
//	//	accounts.Accounts = accounts.Accounts[:limit]
//	//}
//
//	ids := make([]int, 0)
//	for _, a := range accounts.Accounts {
//		ids = append(ids, account.GetNewIds(a)...)
//		if len(ids) > limit {
//			ids = ids[:limit]
//			break
//		}
//	}
//
//	err = collection.Find(bson.M{"id": bson.M{"$in": ids}}).Select(bson.M{
//		"id":     1,
//		"email":  1,
//		"status": 1,
//		"fname":  1,
//		"sname":  1}).Sort("-id").All(&accounts.Accounts)
//	if err != nil {
//		log.Println("[ERROR] ", err)
//	}
//
//	err = json.NewEncoder(w).Encode(accounts)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		log.Println("[ERROR] ", err)
//	}
//}
//
//func exists(v string) bson.M {
//	switch v {
//	case "0":
//		return bson.M{"$exists": true}
//	case "1":
//		return bson.M{"$exists": false}
//	}
//	return nil
//}
//
//func yearInterval(year int) bson.M {
//	return bson.M{
//		"$gte": time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC).Unix(),
//		"$lt":  time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC).Unix(),
//	}
//}
