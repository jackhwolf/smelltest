package api

import (
	"log"
	"net/http"
	"smelltest/backend"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// ReverseLookupResponse is used for results from
// querying the RLTable
type ReverseLookupResponse struct {
	*ReverseLookup
	Exists bool
}

// ReverseLookupItem looks up an item by hash key in the RLTable and
// returns a pointer to a response struct
func ReverseLookupItem(hkv string) *ReverseLookupResponse {
	rlt := backend.MakeReverseLookupTable(false)
	users := rlt.Lookup(rlt.Hkey.Name, hkv)
	if len(users) == 0 {
		return &ReverseLookupResponse{BlankReverseLookup(), false}
	}
	rl := BlankReverseLookup()
	err := dynamodbattribute.UnmarshalMap(users[0], &rl)
	if err != nil {
		panic(err)
	}
	return &ReverseLookupResponse{rl, true}
}

// AuthorizeTokenAttemptResponse is what we send after an auth attempt
// to either block client or give them a token
type AuthorizeTokenAttemptResponse struct {
	Token, Username string
	Success         bool
}

// AuthorizeToken checks that a token exists and returns a pointer
// to an AuthorizeTokenAttemptResponse with relevant info
func AuthorizeToken(token string) *AuthorizeTokenAttemptResponse {
	lookup := ReverseLookupItem(token)
	if !lookup.Exists {
		return &AuthorizeTokenAttemptResponse{"", "", false}
	}
	return &AuthorizeTokenAttemptResponse{token, lookup.ReverseLookup.ReverseValue, true}

}

// AuthenticateUnPwResponse is what we send after user attemps to
type AuthenticateUnPwResponse struct {
	*User
	Success bool
}

// AuthenticateUnPw checks that the provided username password
// combo exists and returns a pointer to an AuthenticateUnPwResponse
func AuthenticateUnPw(username, password string) *AuthenticateUnPwResponse {
	lookup := ReverseLookupItem(username)
	if lookup.Exists {
		uid := lookup.ReverseLookup.ReverseValue
		ut := backend.MakeUserTable(false)
		users := ut.Lookup(ut.Hkey.Name, uid)
		user := BlankUser()
		err := dynamodbattribute.UnmarshalMap(users[0], &user)
		if err != nil {
			panic(err)
		}
		if user.Password == password {
			return &AuthenticateUnPwResponse{user, true}
		}
	}
	return &AuthenticateUnPwResponse{BlankUser(), false}
}

// AuthHandler is to authenticate requests
func AuthHandler(w http.ResponseWriter, req *http.Request) (*ReverseLookupResponse, bool) {
	token := req.Header.Get("X-Access-Token")
	w.Header().Set("Content-Type", "application/json")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "missing X-Access-Token header"}`))
		return &ReverseLookupResponse{nil, false}, false
	}
	lookup := ReverseLookupItem(token)
	if !lookup.Exists {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "unauthorized"}`))
		return lookup, false
	}
	return lookup, true
}

// MuxWrappable will be the type of our handler funcs that we want wrapped for auth, etc
type MuxWrappable func(w http.ResponseWriter, req *http.Request) (int, error)

func makeErrMap() map[int]string {
	errMap := make(map[int]string)
	errMap[http.StatusRequestEntityTooLarge] = "Request too large!"
	errMap[http.StatusUnauthorized] = "Unauthorized!"
	return errMap
}

// helper to print hit
func logEndpoint(req *http.Request) {
	log.Println(req.Method + ": " + req.URL.RequestURI())
}

// Wrapped will wrap all of the handler functions to check requests
func (fn MuxWrappable) Wrapped(tokenCheck bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		logEndpoint(req)
		if tokenCheck {
			_, ok := AuthHandler(w, req)
			if !ok {
				return // AuthHandler takes care of this stuff
			}
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		status, err := fn(w, req)
		if err != nil {
			emap := makeErrMap()
			log.Println(err)
			http.Error(w, emap[status], status)
		}
	}
}
