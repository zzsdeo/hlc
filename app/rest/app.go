package rest

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
)

type App struct {
	router          *mux.Router
	specsRepository store.SpecsRepository
}

//func (a *App) Initialize() {
//	a.router = mux.NewRouter()
//	s := store.MockStore{}
//	s.Initialize()
//	a.specsRepository = &s
//	a.initializeRoutes()
//}

func (a *App) Initialize(url string) {
	a.router = mux.NewRouter()
	s := store.MongoStore{}
	err := s.Initialize(url)
	if err != nil {
		//log.Fatal("[ERROR] ", err)
		log.Println("[ERROR] ", err)
	}
	a.specsRepository = &s
	a.initializeRoutes()
}

func (a *App) Run(url string) {
	log.Println("[INFO] start server on", url)
	log.Fatal("[ERROR] ", http.ListenAndServe(url, a.router))
}

func (a *App) initializeRoutes() {
	a.router.HandleFunc("/api/v1/ping", a.ping).Methods(http.MethodGet)
	a.router.HandleFunc("/api/v1/specs", a.getSpecs).Methods(http.MethodGet)
	a.router.HandleFunc("/api/v1/specs/{id}", a.getSpec).Methods(http.MethodGet)
	a.router.HandleFunc("/api/v1/specs", a.createSpec).Methods(http.MethodPost)
	a.router.HandleFunc("/api/v1/specs/{id}", a.updateSpec).Methods(http.MethodPut)
	a.router.HandleFunc("/api/v1/specs/{id}", a.deleteSpec).Methods(http.MethodDelete)

	a.router.HandleFunc("/api/v1/specs/{id}/items", a.createItem).Methods(http.MethodPost)
	//a.router.HandleFunc("/api/v1/specs/{id}/items/{item_id}", a.deleteItem).Methods(http.MethodDelete) todo
}

func (a *App) ping(w http.ResponseWriter, r *http.Request) {
	_, err := io.WriteString(w, "pong")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] ", r, err)
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
