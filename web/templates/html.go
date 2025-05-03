package templates

import (
	"embed"
	"html/template"
	"io"
	"strings"
)

//go:embed *
var files embed.FS

var funcs = template.FuncMap{
	"uppercase": func(v string) string {
		return strings.ToUpper(v)
	},
}

func parse(file string) *template.Template {
	return template.Must(
		template.New("layout.html").Funcs(funcs).ParseFS(files, "layout.html", file))
}

var (
	home = parse("home.html")
	user = parse("user.html")
)

type HomeParams struct {
}

func Home(w io.Writer) error {
	return home.Execute(w, HomeParams{})
}

type UserParams struct {
	picture  any
	nickname any
}

func User(w io.Writer, p any) error {
	return user.Execute(w, p)
}
