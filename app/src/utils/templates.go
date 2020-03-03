package utils

import (
	"html/template"
	html "html/template"
	"net/http"
	"os"
	"strings"
)

var (
	wd, err = os.Getwd()
	wdsrc   = wd + "/src/"
)

var rootpath = cleanWD(wdsrc)
var STATIC_ROOT_PATH = rootpath + "/static"

var (
	LOGIN    = STATIC_ROOT_PATH + "/templates/login.html"
	FOOTER   = STATIC_ROOT_PATH + "/templates/footer.html"
	LAYOUT   = STATIC_ROOT_PATH + "/templates/layout.html"
	REGISTER = STATIC_ROOT_PATH + "/templates/register.html"
	TIMELINE = STATIC_ROOT_PATH + "/templates/timeline.html"
)

// GetTemplate returns executable template with header and footer
func GetTemplate(route string) *html.Template {
	return template.Must(template.ParseFiles(route, LAYOUT, FOOTER))
}

func RenderTemplate(w http.ResponseWriter, route string, data interface{}) {
	tmpl := GetTemplate(route)
	tmpl.Execute(w, data)
}

func cleanWD(wd string) string {
	suffix := "/src/src/"
	if strings.HasSuffix(wd, suffix) {
		wd := wd[:len(wd)-4]
		return wd
	} else {
		return wd
	}
}
