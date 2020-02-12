package utils

import (
	"html/template"
	html "html/template"
)

const STATIC_ROOT_PATH = "./src/static"

const (
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
