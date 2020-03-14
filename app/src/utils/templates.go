package utils

import (
	"html/template"
	"net/http"
	"os"
	"strings"
)

var (
	wd, _ = os.Getwd()
	wdsrc = wd + "/src/"
)

var rootpath = cleanWD(wdsrc)
var staticRootPath = rootpath + "/static"

var (
	Login    = staticRootPath + "/templates/login.html"
	Footer   = staticRootPath + "/templates/footer.html"
	Layout   = staticRootPath + "/templates/layout.html"
	Register = staticRootPath + "/templates/register.html"
	Timeline = staticRootPath + "/templates/timeline.html"
)

func RenderTemplate(w http.ResponseWriter, route string, data interface{}) {
	templ, err := template.ParseFiles(route, Layout, Footer)
	templ = template.Must(templ, err)
	templ.Execute(w, data)
}

func cleanWD(wd string) string {
	suffix := "/src/src/"
	if strings.HasSuffix(wd, suffix) {
		return wd[:len(wd)-4]
	} else {
		return wd
	}
}
