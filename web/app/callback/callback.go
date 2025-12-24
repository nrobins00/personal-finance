// web/app/callback/callback.go

package callback

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"

	"github.com/nrobins00/personal-finance/platform/authenticator"
	"github.com/nrobins00/personal-finance/platform/database"
)

// Handler for our callback.
func Handler(auth *authenticator.Authenticator, db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value("session").(*sessions.Session)

		if r.URL.Query().Get("state") != session.Values["state"] {
			w.Write([]byte("Invalid state parameter."))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Exchange an authorization code for a token.
		token, err := auth.Exchange(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			//w.Write([]byte("Failed to exchange an authorization code for a token."))
			fmt.Fprint(w, "Failed to exchange an authorization code for a token.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		idToken, err := auth.VerifyIDToken(r.Context(), token)
		if err != nil {
			fmt.Fprint(w, "Failed to verify ID Token.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			fmt.Fprint(w, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userInfo, err := auth.GetUserInfo(r.Context(), token)
		if err != nil {
			fmt.Fprint(w, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userId, err := db.GetUserIdByEmail(userInfo.Email)
		if err != nil {
			// no user found, so create one
			userId, err = db.CreateUser(userInfo.Email)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error: %v", err)
			}
		}

		profile["UserId"] = userId
		session.Values["userId"] = userId
		session.Values["access_token"] = token.AccessToken
		session.Values["profile"] = profile

		if err := session.Save(r, w); err != nil {
			fmt.Fprint(w, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Redirect to logged in page.
		http.Redirect(w, r, "/auth/user", http.StatusTemporaryRedirect)
	}
}
