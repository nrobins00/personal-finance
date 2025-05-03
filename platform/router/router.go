// platform/router/router.go

package router

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/nrobins00/personal-finance/web/app/login"
	"github.com/nrobins00/personal-finance/web/app/logout"
	"github.com/nrobins00/personal-finance/web/templates"

	"github.com/nrobins00/personal-finance/platform/authenticator"
	"github.com/nrobins00/personal-finance/web/app/callback"
	"github.com/nrobins00/personal-finance/web/app/user"
)

func getSessionMiddleware(store *sessions.CookieStore) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := store.Get(r, "session-name")

			r = r.WithContext(context.WithValue(r.Context(), "session", session))
			next.ServeHTTP(w, r)
		})
	}
}

func isAuthenticatedMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value("session").(*sessions.Session)
		if session.Values["Profile"] == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// New registers the routes and returns the router.
func New(auth *authenticator.Authenticator) *mux.Router {
	router := mux.NewRouter()

	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := sessions.NewCookieStore([]byte("secret")) // TODO: use actual secret

	router.Use(getSessionMiddleware(store))
	authR := router.PathPrefix("/auth/").Subrouter()
	authR.Use(isAuthenticatedMiddleWare)

	router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("web/static"))))
	// templ := template.Must(template.New("").ParseGlob("web/template/*"))
	// templ.Execute()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("hit home")
		templates.Home(w)
	})
	router.HandleFunc("/login", login.Handler(auth)).Methods("GET")
	router.HandleFunc("/callback", callback.Handler(auth)).Methods("GET")
	router.HandleFunc("/logout", logout.Handler).Methods("GET")
	authR.HandleFunc("/user", user.Handler).Methods("GET")

	return router
}
