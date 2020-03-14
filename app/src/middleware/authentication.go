package middleware

import (
	"net/http"

	"github.com/matt035343/devops/app/src/authentication"
)

func Auth(f handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		user := authentication.GetSessionValue(w, r, "user")
		if user != nil {
			f(w, r)
			return
		}
		http.Redirect(w, r, "/public", http.StatusFound)
	}
}