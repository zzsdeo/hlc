package rest

import (
	"encoding/json"
	"github.com/globalsign/mgo"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
)

type App struct {
	router          *mux.Router
	mongoSession *mgo.Session
}

func (a *App) Initialize(mongoAddr string) {
	a.router = mux.NewRouter()

	session, err := mgo.Dial(mongoAddr)
	if err != nil {
		log.Fatal("[ERROR] ", err)
	}
	a.mongoSession = session

	a.initializeRoutes()
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

			continue
			case:"fname_null"
			case:"sname_eq"
			case:"sname_starts"
			case:"sname_null"
			case:"phone_code"
			case:"phone_null"
			case:"country_eq"
			case:"country_null"
			case:"city_eq"
			case:"city_any"
			case:"city_null"
			case:"birth_lt"
			case:"birth_gt"
			case:"birth_year"
			case:"interests_contains"
			case:"interests_any"
			case:"likes_contains"
			case:"premium_now"
			case:"premium_null"
			case:"limit"
		}
	}

}







func (a *App) getSpecs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	specs, err := a.specsRepository.GetSpecs()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, err)
	}

	err = json.NewEncoder(w).Encode(specs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, err)
	}
}

func (a *App) getSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	spec, err := a.specsRepository.GetSpec(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println("[ERROR] ", r, err)
		return
	}

	err = json.NewEncoder(w).Encode(spec)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, err)
	}
}

func (a *App) createSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var encodedSpec models.Spec
	err := json.NewDecoder(r.Body).Decode(&encodedSpec)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, err)
		return
	}

	spec, err := a.specsRepository.CreateSpec(encodedSpec)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, encodedSpec, err)
		return
	}

	err = json.NewEncoder(w).Encode(spec)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, spec, err)
	}
}

func (a *App) updateSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var encodedSpec models.Spec
	err := json.NewDecoder(r.Body).Decode(&encodedSpec)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, err)
		return
	}

	spec, err := a.specsRepository.UpdateSpec(encodedSpec)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, spec, err)
		return
	}

	err = json.NewEncoder(w).Encode(spec)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, spec, err)
	}
}

func (a *App) deleteSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	err := a.specsRepository.DeleteSpec(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *App) createItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var encodedItem models.Item
	err := json.NewDecoder(r.Body).Decode(&encodedItem)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, err)
		return
	}

	params := mux.Vars(r)

	item, err := a.specsRepository.CreateItem(params["id"], encodedItem)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, item, err)
		return
	}

	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, item, err)
	}
}
