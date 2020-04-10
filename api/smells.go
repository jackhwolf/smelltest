package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// handler for when client wants to logout, uses AuthHandler
func getSmells(w http.ResponseWriter, req *http.Request) (int, error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	ss := GetAllSmells()
	WriteStruct(w, ss)
	return http.StatusOK, nil
}

// BuildSmellsRouter builds the mux router for the api
func BuildSmellsRouter(r *mux.Router) {
	api := r.PathPrefix("/v1/smells").Subrouter().StrictSlash(true)
	api.HandleFunc("/", MuxWrappable(getSmells).Wrapped(false)).Methods(http.MethodGet)
}
