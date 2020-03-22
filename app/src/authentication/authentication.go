package authentication

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var secretKey = []byte("development key")
var store = sessions.NewCookieStore(secretKey)

func getSession(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return session
}

//PutSessionValue Saves the given key and value in the current (cookie based) session of the request.
func PutSessionValue(w http.ResponseWriter, r *http.Request, key string, value interface{}) error {
	session := getSession(w, r)
	session.Values[key] = value
	return session.Save(r, w)
}

//GetSessionValue Retrieves value to the given key in the current (cookie based) session of the request.
func GetSessionValue(w http.ResponseWriter, r *http.Request, key string) interface{} {
	session := getSession(w, r)
	return session.Values[key] //this should be able to fail!, error from getSession should be propagated
}

//ClearSession Deletes all key/value pairs in the current (cookie based) session of the request.
func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session := getSession(w, r)
	session.AddFlash("You were logged out")
	session.Options.MaxAge = -1
	return session.Save(r, w)
}

//Flash Flashes the current (cookie based) session of the request.
func Flash(w http.ResponseWriter, r *http.Request, msg string) {
	session := getSession(w, r)
	session.AddFlash(msg)
}
