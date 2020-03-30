package utils

import (
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/matt035343/devops/app/src/log"
)

var (
	wd, _ = os.Getwd()
	wdsrc = wd + "/src/"
)

var rootpath = cleanWD(wdsrc)
var staticRootPath = rootpath + "/static"

var (
	//Login Constant URL path to login page HTML template.
	Login = staticRootPath + "/templates/login.html"
	//Footer Constant URL path to footer HTML template.
	Footer = staticRootPath + "/templates/footer.html"
	//Layout Constant URL path to header HTML template.
	Layout = staticRootPath + "/templates/layout.html"
	//Register Constant URL path to register page HTML template.
	Register = staticRootPath + "/templates/register.html"
	//Timeline Constant URL path to timeline page HTML template.
	Timeline = staticRootPath + "/templates/timeline.html"
)

//RenderTemplate Renders a HTML template to the user using the given data and the route to wanted HTML template.
func RenderTemplate(w http.ResponseWriter, route string, data interface{}) error {
	templ, err := template.ParseFiles(route, Layout, Footer)
	templ = template.Must(templ, err)
	err = templ.Execute(w, data)
	log.ErrorErr("Error rendering HTML template", err)
	return err
}

func cleanWD(wd string) string {
	suffix := "/src/src/"
	if strings.HasSuffix(wd, suffix) {
		return wd[:len(wd)-4]
	} else {
		return wd
	}
}
