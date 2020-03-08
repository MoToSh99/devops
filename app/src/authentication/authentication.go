package authentication

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var SECRET_KEY = []byte("development key")
var STORE = sessions.NewCookieStore(SECRET_KEY)

func getSession(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := STORE.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return session
}

func PutSessionValue(w http.ResponseWriter, r *http.Request, key string, value interface{}) error {
	session := getSession(w, r)
	session.Values[key] = value
	return session.Save(r, w)
}

func GetSessionValue(w http.ResponseWriter, r *http.Request, key string) interface{} {
	session := getSession(w, r)
	return session.Values[key] //this should be able to fail!, error from getSession should be propagated
}

func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session := getSession(w, r)
	session.AddFlash("You were logged out")
	session.Options.MaxAge = -1
	return session.Save(r, w)
}

func Flash(w http.ResponseWriter, r *http.Request, msg string) {
	session := getSession(w, r)
	session.AddFlash(msg)
}
