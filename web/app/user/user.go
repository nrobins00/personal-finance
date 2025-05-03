// web/app/user/user.go

package user

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/nrobins00/personal-finance/web/templates"
)

// Handler for our logged-in user page.
func Handler(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	session := context.Value("session").(*sessions.Session)
	profile := session.Values["profile"]

	templates.User(w, profile)
}
