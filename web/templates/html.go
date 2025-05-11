package templates

import (
	"embed"
	"html/template"
	"io"
	"strings"

	"github.com/nrobins00/personal-finance/types"
)

//go:embed *
var files embed.FS

var funcs = template.FuncMap{
	"uppercase": func(v string) string {
		return strings.ToUpper(v)
	},
}

func parse(filesToParse ...string) *template.Template {
	//templates := template.Must(template.ParseFS(files, "navbar.html"))
	filesToParse = append(filesToParse, "layout.html")
	return template.Must(template.New("layout.html").Funcs(funcs).ParseFS(files, filesToParse...))
	//return template.Must(
	//	template.New("layout.html").Funcs(funcs).ParseFS(files, "layout.html", file))
}

var (
	//navbar       = parse("navbar.html")
	login        = parse("login.html")
	user         = parse("user.html")
	accounts     = parse("accounts.html")
	budget       = parse("budget.html")
	transactions = parse("transactions.html")
	updateLink   = parse("updateLink.html")
	home         = parse("home.html", "navbar.html", "transactions.html")
)

type LoginParams struct {
}

func Login(w io.Writer) error {
	return login.Execute(w, LoginParams{})
}

type UserParams struct {
	picture  any
	nickname any
}

func User(w io.Writer, p any) error {
	return user.Execute(w, p)
}

type HomeParams struct {
	Spent            float32
	Budget           float32
	Transactions     []types.Transaction
	UserId           int64
	Page             string
	MoreTransactions bool
	Categories       []string
	SortCol          string
	SortDir          string
}

func Home(w io.Writer, p any) error {
	return home.Execute(w, p)
}
