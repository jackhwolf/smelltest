package main

import (
	"log"
	"net/http"
	"smelltest/api"
	"smelltest/backend"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"message": "not found"}`))
}

func main() {
	backend.MainMakeAllTables()

	r := mux.NewRouter()
	r.HandleFunc("", notFound)

	api.BuildUserRouter(r)
	api.BuildSmellsRouter(r)
	// srv := &http.Server{
	// 	Handler: r,
	// 	Addr:    "127.0.0.1:8081",
	// 	// Good practice: enforce timeouts for servers you create!
	// 	WriteTimeout: 15 * time.Second,
	// 	ReadTimeout:  15 * time.Second,
	// }
	// log.Fatal(srv.ListenAndServe())
	log.Fatal(http.ListenAndServe(":8081", handlers.CORS(handlers.AllowedOrigins([]string{"*"}))(r)))
}
