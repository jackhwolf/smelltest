package api

import (
	"encoding/json"
	"net/http"
	"smelltest/backend"

	"github.com/gorilla/mux"
	"github.com/segmentio/ksuid"
)

// helper function to make a new token
func newToken() string {
	return ksuid.New().String()
}

// handler for when client wants to signup
func postNewUser(w http.ResponseWriter, req *http.Request) (int, error) {
	ut := backend.MakeUserTable(false)
	// load request json data into a Dog struct and throw an err to client
	// if they send us fields not declared in Dog
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	u := BlankUser()
	err := decoder.Decode(&u)
	if err != nil {
		panic(err)
	}
	lookup := ReverseLookupItem(u.Username)
	w.Header().Set("Content-Type", "application/json")
	if lookup.Exists {
		w.WriteHeader(http.StatusUnauthorized)
		WriteStruct(w, &AuthorizeTokenAttemptResponse{"", "", false})
	} else {
		// give user IDs and add them to table
		u.Setup()
		ut.AddItem(u)
		// add mapping of username --> ID for login
		rlt := backend.MakeReverseLookupTable(false)
		rte := &ReverseLookup{u.Username, u.ID}
		rlt.AddItem(rte)
		// add mapping of token --> ID for auth
		token := newToken()
		rte = &ReverseLookup{token, u.ID}
		rlt.AddItem(rte)
		w.WriteHeader(http.StatusCreated)
		WriteStruct(w, &AuthorizeTokenAttemptResponse{token, u.Username, true})
	}
	return 1, nil
}

type loginCreds struct {
	Username, Password string
}

// handler for when client wants to login
func loginUser(w http.ResponseWriter, req *http.Request) (int, error) {
	// log request json data into a loginCreds struct to scan for un/pw combo
	lc := &loginCreds{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&lc)
	if err != nil {
		panic(err)
	}
	auth := AuthenticateUnPw(lc.Username, lc.Password)
	w.Header().Set("Content-Type", "application/json")
	if auth.Success {
		token := newToken()
		rlt := backend.MakeReverseLookupTable(false)
		rte := &ReverseLookup{token, auth.User.ID}
		rlt.AddItem(rte)
		w.WriteHeader(http.StatusOK)
		WriteStruct(w, &AuthorizeTokenAttemptResponse{token, auth.User.Username, true})
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		WriteStruct(w, &AuthorizeTokenAttemptResponse{"", "", false})
	}
	return 1, nil
}

// handler for when client wants to logout
func logoutUser(w http.ResponseWriter, req *http.Request) (int, error) {
	// log request json data into a loginCreds struct to scan for un/pw combo
	token := req.Header.Get("X-Access-Token")
	lookup := ReverseLookupItem(token)
	w.Header().Set("Content-Type", "application/json")
	if lookup.Exists {
		rlt := backend.MakeReverseLookupTable(false)
		rlt.Delete(token, lookup.ReverseValue)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "logout"}`))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "unauthorized"}`))
	}
	return 1, nil
}

// handler for when client wants to delete account
func deleteUser(w http.ResponseWriter, req *http.Request) (int, error) {
	// decode request json data into a loginCreds struct to scan for un/pw combo
	lc := &loginCreds{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&lc)
	if err != nil {
		panic(err)
	}
	auth := AuthenticateUnPw(lc.Username, lc.Password)
	w.Header().Set("Content-Type", "application/json")
	if auth.Success {
		rlt := backend.MakeReverseLookupTable(false)
		rlt.Delete(lc.Username, auth.User.ID)
		ut := backend.MakeUserTable(false)
		ut.Delete(auth.User.ID, auth.User.ID2)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "delete"}`))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "unauthorized"}`))
	}
	return 1, nil
}

// BuildUserRouter builds the mux router for the api
func BuildUserRouter(r *mux.Router) {
	api := r.PathPrefix("/v1/user").Subrouter()
	api.HandleFunc("/", MuxWrappable(postNewUser).Wrapped(false)).Methods(http.MethodPost)
	api.HandleFunc("/login/", MuxWrappable(loginUser).Wrapped(false)).Methods(http.MethodPost)
	api.HandleFunc("/logout/", MuxWrappable(logoutUser).Wrapped(false)).Methods(http.MethodGet)
	api.HandleFunc("/", MuxWrappable(deleteUser).Wrapped(false)).Methods(http.MethodDelete)
}
