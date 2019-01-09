package rest

import (
	"bytes"
	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
	"hlc/app/models"
	"hlc/app/store"
	"log"
	"strconv"
	"strings"
)

type queryKeys struct {
	SexEq             []byte
	EmailDomain       []byte
	EmailLt           []byte
	EmailGt           []byte
	StatusEq          []byte
	StatusNeq         []byte
	FNameEq           []byte
	FNameAny          []byte
	FNameNull         []byte
	SNameEq           []byte
	SNameStarts       []byte
	SNameNull         []byte
	PhoneCode         []byte
	PhoneNull         []byte
	CountryEq         []byte
	CountryNull       []byte
	CityEq            []byte
	CityAny           []byte
	CityNull          []byte
	BirthLt           []byte
	BirthGt           []byte
	BirthYear         []byte
	InterestsContains []byte
	InterestsAny      []byte
	LikesContains     []byte
	PremiumNow        []byte
	PremiumNull       []byte
	Limit             []byte
	QueryId           []byte
}

type paths struct {
	filterPath []byte
}

type App struct {
	queryKeys
	paths
	db  *store.DB
	now int //current time from options.txt
}

func (a *App) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	if ctx.IsGet() {
		if bytes.Equal(ctx.Path(), a.filterPath) {
			a.filter(ctx)
		}
	} else {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
	}
}

func (a *App) Initialize(now int) {
	a.db = store.NewDB()
	a.now = now

	a.filterPath = []byte("/accounts/filter/")

	a.SexEq = []byte("sex_eq")
	a.EmailDomain = []byte("email_domain")
	a.EmailLt = []byte("email_lt")
	a.EmailGt = []byte("email_gt")
	a.StatusEq = []byte("status_eq")
	a.StatusNeq = []byte("status_neq")
	a.FNameEq = []byte("fname_eq")
	a.FNameAny = []byte("fname_any")
	a.FNameNull = []byte("fname_null")
	a.SNameEq = []byte("sname_eq")
	a.SNameStarts = []byte("sname_starts")
	a.SNameNull = []byte("sname_null")
	a.PhoneCode = []byte("phone_code")
	a.PhoneNull = []byte("phone_null")
	a.CountryEq = []byte("country_eq")
	a.CountryNull = []byte("country_null")
	a.CityEq = []byte("city_eq")
	a.CityAny = []byte("city_any")
	a.CityNull = []byte("city_null")
	a.BirthLt = []byte("birth_lt")
	a.BirthGt = []byte("birth_gt")
	a.BirthYear = []byte("birth_year")
	a.InterestsContains = []byte("interests_contains")
	a.InterestsAny = []byte("interests_any")
	a.LikesContains = []byte("likes_contains")
	a.PremiumNow = []byte("premium_now")
	a.PremiumNull = []byte("premium_null")
	a.Limit = []byte("limit")
	a.QueryId = []byte("query_id")
}

func (a *App) LoadData(accounts []models.Account) {
	a.db.LoadData(accounts)
	//a.db.LoadMinData(accounts)
	log.Println("[INFO] added ", len(accounts), " accounts")
}

func (a *App) CreateIndexes() {
	log.Println("[INFO] indexing started")
	a.db.CreateIndexes(a.now)
	log.Println("[INFO] indexing finished")
}

func (a *App) Run(listenAddr string) {
	log.Println("[INFO] start server on", listenAddr)
	log.Fatal("[ERROR] ", fasthttp.ListenAndServe(listenAddr, a.HandleFastHTTP))
}

//func (a *App) initializeRoutes() {
//	a.router.HandleFunc("/accounts/filter/", a.filter).Methods(http.MethodGet)
//	//a.router.HandleFunc("/accounts/group/", a.group).Methods(http.MethodGet)
//	//a.router.HandleFunc("/accounts/{id}/recommend/", a.recommend).Methods(http.MethodGet)
//	//a.router.HandleFunc("/accounts/{id}/suggest/", a.suggest).Methods(http.MethodGet)
//
//	// Регистрация pprof-обработчиков
//	//a.router.HandleFunc("/debug/pprof/", pprof.Index)
//	//a.router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
//	//a.router.HandleFunc("/debug/pprof/profile", pprof.Profile)
//	//a.router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
//	//a.router.HandleFunc("/debug/pprof/trace", pprof.Trace)
//}

func (a *App) filter(ctx *fasthttp.RequestCtx) {
	//defer utils.TimeTrack(time.Now(), ctx.QueryArgs().Peek("query_id"))
	ctx.SetContentType("application/json")
	query := store.M{}
	isBadArg := false
	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		if isBadArg {
			return
		} else if len(value) == 0 {
			isBadArg = true
			return
		} else if bytes.Equal(key, a.SexEq) {
			query["sex_eq"] = string(value)
			return
		} else if bytes.Equal(key, a.EmailDomain) {
			query["email_domain"] = string(value)
			return
		} else if bytes.Equal(key, a.EmailLt) {
			query["email_lt"] = string(value)
			return
		} else if bytes.Equal(key, a.EmailGt) {
			query["email_gt"] = string(value)
			return
		} else if bytes.Equal(key, a.StatusEq) {
			query["status_eq"] = string(value)
			return
		} else if bytes.Equal(key, a.StatusNeq) {
			query["status_neq"] = string(value)
			return
		} else if bytes.Equal(key, a.FNameEq) {
			query["fname_eq"] = string(value)
			return
		} else if bytes.Equal(key, a.FNameAny) {
			query["fname_any"] = strings.Split(string(value), ",")
			return
		} else if bytes.Equal(key, a.FNameNull) {
			query["fname_null"] = string(value)
			return
		} else if bytes.Equal(key, a.SNameEq) {
			query["sname_eq"] = string(value)
			return
		} else if bytes.Equal(key, a.SNameStarts) {
			query["sname_starts"] = string(value)
			return
		} else if bytes.Equal(key, a.SNameNull) {
			query["sname_null"] = string(value)
			return
		} else if bytes.Equal(key, a.PhoneCode) {
			query["phone_code"] = string(value)
			return
		} else if bytes.Equal(key, a.PhoneNull) {
			query["phone_null"] = string(value)
			return
		} else if bytes.Equal(key, a.CountryEq) {
			query["country_eq"] = string(value)
			return
		} else if bytes.Equal(key, a.CountryNull) {
			query["country_null"] = string(value)
			return
		} else if bytes.Equal(key, a.CityEq) {
			query["city_eq"] = string(value)
			return
		} else if bytes.Equal(key, a.CityAny) {
			query["city_any"] = strings.Split(string(value), ",")
			return
		} else if bytes.Equal(key, a.CityNull) {
			query["city_null"] = string(value)
			return
		} else if bytes.Equal(key, a.BirthLt) {
			birth, err := strconv.Atoi(string(value))
			if err != nil {
				log.Println("[ERROR] ", err)
				isBadArg = true
				return
			}
			query["birth_lt"] = birth
			return
		} else if bytes.Equal(key, a.BirthGt) {
			birth, err := strconv.Atoi(string(value))
			if err != nil {
				log.Println("[ERROR] ", err)
				isBadArg = true
				return
			}
			query["birth_gt"] = birth
			return
		} else if bytes.Equal(key, a.BirthYear) {
			year, err := strconv.Atoi(string(value))
			if err != nil {
				log.Println("[ERROR] ", err)
				isBadArg = true
				return
			}
			query["birth_year"] = year
			return
		} else if bytes.Equal(key, a.InterestsContains) {
			query["interests_contains"] = strings.Split(string(value), ",")
			return
		} else if bytes.Equal(key, a.InterestsAny) {
			query["interests_any"] = strings.Split(string(value), ",")
			return
		} else if bytes.Equal(key, a.LikesContains) {
			likes := strings.Split(string(value), ",")
			likeIds := make([]int, 0)
			for _, like := range likes {
				l, err := strconv.Atoi(like)
				if err != nil {
					log.Println("[ERROR] ", err)
					isBadArg = true
					return
				}
				likeIds = append(likeIds, l)
			}
			//log.Println("[DEBUG] ", likeIds)
			query["likes_contains"] = likeIds
			return
		} else if bytes.Equal(key, a.PremiumNow) {
			query["premium_now"] = a.now
			return
		} else if bytes.Equal(key, a.PremiumNull) {
			query["premium_null"] = string(value)
			return
		} else if bytes.Equal(key, a.Limit) {
			limit, err := strconv.Atoi(string(value))
			if err != nil {
				log.Println("[ERROR] ", err)
				isBadArg = true
				return
			}
			if limit < 0 {
				isBadArg = true
				return
			}
			query["limit"] = limit
			return
		} else if bytes.Equal(key, a.QueryId) {
			return
		} else {
			isBadArg = true
			return
		}
	})

	//log.Println("[DEBUG] query=", query)

	if isBadArg {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	accounts := a.db.Find(query)

	b, err := easyjson.Marshal(&accounts)
	//err := json.NewEncoder(w).Encode(a.db.Find(query))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		log.Println("[ERROR] ", err)
		return
	}

	_, err = ctx.Write(b)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		log.Println("[ERROR] ", err)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
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
