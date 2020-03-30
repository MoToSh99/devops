package authentication

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/matt035343/devops/app/src/log"
)

var secretKey = []byte("development key")
var store = sessions.NewCookieStore(secretKey)

func getSession(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := store.Get(r, "session")
	if err != nil {
		log.ErrorErr("Error getting session", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return session
}

//PutSessionValue Saves the given key and value in the current (cookie based) session of the request.
func PutSessionValue(w http.ResponseWriter, r *http.Request, key string, value interface{}) error {
	session := getSession(w, r)
	session.Values[key] = value
	err := session.Save(r, w)
	log.ErrorErr("Error saving session key %s and value %v", err, key, value)
	return err
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
	err := session.Save(r, w)
	log.ErrorErr("Error saving  cleared session", err)
	return err
}

//Flash Flashes the current (cookie based) session of the request.
func Flash(w http.ResponseWriter, r *http.Request, msg string) {
	session := getSession(w, r)
	session.AddFlash(msg)
}
