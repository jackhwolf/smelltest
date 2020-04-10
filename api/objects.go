package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// WriteStruct will write the contents of a given struct to the http response
func WriteStruct(w http.ResponseWriter, dst interface{}) {
	js, err := json.Marshal(dst)
	if err != nil {
		panic(err)
	}
	w.Write(js)
	// fmt.Fprintf(w, "%+v\n", dst)
}

////////////////////////
// data struct for user
////////////////////////

// User struct will hold data for each user
type User struct {
	ID, ID2            string
	Username, Password string
	JoinedAt           int64
}

// Setup sets the ID and ID2 of this struct
func (u *User) Setup() {
	if u.ID == "" {
		id := uuid.New().String()
		u.ID, u.ID2 = id, id
	}
	u.JoinedAt = time.Now().UnixNano()
}

// GetIDs returns the ID and ID2 of this struct
func (u *User) GetIDs() (string, string) {
	return u.ID, u.ID2
}

// BlankUser returns a pointer to a zero-init'd User struct
func BlankUser() *User {
	d := &User{}
	return d
}

/////////////////////////////////
// data struct for reverse lookup
/////////////////////////////////

// ReverseLookup struct will hold data for each reverse lookup
type ReverseLookup struct {
	ReverseKey, ReverseValue string
}

// BlankReverseLookup returns a pointer to a zero-init'd User struct
func BlankReverseLookup() *ReverseLookup {
	rk := &ReverseLookup{}
	return rk
}

///////////////////////////////
// data structs for smell stuff
///////////////////////////////

// Smell is a struct to hold info abt a specific smell
type Smell struct {
	Name, Desc string
}

// SmellsStruct is a struct to hold all smells
type SmellsStruct struct {
	Smells map[string]*Smell
	N      int
}

// AddSmell will take the params of a new Smell and add if not present
func (ss *SmellsStruct) AddSmell(name, desc string) bool {
	_, ok := ss.Smells[name]
	if !ok {
		ss.Smells[name] = &Smell{name, desc}
		ss.N++
	}
	return !ok
}

// DelSmell will take the name of a Smell and delete if present
func (ss *SmellsStruct) DelSmell(name string) bool {
	_, ok := ss.Smells[name]
	if ok {
		delete(ss.Smells, name)
		ss.N--
	}
	return ok
}

// GetAllSmells is a helper to return a full SmellsStruct
// this is where all pre-defined smells should be added, either
// manually or loaded in from a file (TODO, PREFERRED)
func GetAllSmells() *SmellsStruct {
	ss := &SmellsStruct{}
	ss.Smells = make(map[string]*Smell)
	ss.AddSmell("A", "Smell A")
	ss.AddSmell("B", "Smell B")
	ss.AddSmell("C", "Smell C")
	return ss
}

// SmellEntry is a struct for when users enter info
// about a smell test they did at home
type SmellEntry struct {
	ID, UID   string
	Ratings   *map[string]int
	CreatedAt int64
}

// Setup sets the ID and ID2 of this struct
func (se *SmellEntry) Setup(uid string) {
	if se.ID == "" {
		id := uuid.New().String()
		se.ID, se.UID = id, uid
		se.CreatedAt = time.Now().UnixNano()
	}
}
