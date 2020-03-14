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
var staticRootPath = rootpath + "/static"

var (
	Login    = staticRootPath + "/templates/login.html"
	Footer   = staticRootPath + "/templates/footer.html"
	Layout   = staticRootPath + "/templates/layout.html"
	Register = staticRootPath + "/templates/register.html"
	Timeline = staticRootPath + "/templates/timeline.html"
)

// GetTemplate returns executable template with header and footer
func GetTemplate(route string) *html.Template {
	return template.Must(template.ParseFiles(route, Layout, Footer))
}

func RenderTemplate(w http.ResponseWriter, route string, data interface{}) {
	tmpl := GetTemplate(route)
	tmpl.Execute(w, data)
}

func cleanWD(wd string) string {
	suffix := "/src/src/"
	if strings.HasSuffix(wd, suffix) {
		return wd[:len(wd)-4]
	} else {
		return wd
	}
}
