package authentiction

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var secretKey = []byte("development key")
var store = sessions.NewCookieStore(secretKey)

func Auth(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetSessionValue(w, r, "user")
		if user != nil {
			f(w, r)
			return
		}
		http.Redirect(w, r, "/public", http.StatusFound)
	}
}

func getSession(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := store.Get(r, "session")
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
