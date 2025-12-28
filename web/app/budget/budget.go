package budgets

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/nrobins00/personal-finance/platform/database"
	"github.com/nrobins00/personal-finance/web/templates"
)

func BudgetsHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		context := r.Context()
		session := context.Value("session").(*sessions.Session)

		userId := session.Values["userId"].(int64)

		budget, err := db.GetBudget(int(userId))

		if err != nil {
			log.Println("error geting items: ", err)
		}

		params := templates.BudgetParams{
			CurrentBudget: budget,
		}

		err = templates.Budget(w, params)

		if err != nil {
			slog.Error(err.Error())
		}
	}
}
