package middleware

import "net/http"

type middleware func(func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request)

func Combine(funcs ...middleware) middleware {
	return func(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		for _, m := range funcs {
			f = m(f)
		}
		return f
	}
}
