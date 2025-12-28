package database

import "fmt"

func (db DB) GetBudget(userId int) (float32, error) {
	const budgetQuery string = `
		SELECT amount FROM budget
		WHERE userId = ?
	`
	var amount float32
	row := db.QueryRow(budgetQuery, userId)
	err := row.Scan(&amount)
	if err != nil {
		return 0, err
	}
	return amount, nil
}

func (db DB) InsertBudget(userId int, amount float32) error {
	const deleteQuery string = `
		DELETE FROM budget
		WHERE userId = ?
	`
	const insertQuery string = `
		INSERT INTO budget (userId, amount)
		VALUES (?, ?);
	`
	_, err := db.Exec(deleteQuery, userId)

	if err != nil {
		return err
	}

	_, err = db.Exec(insertQuery, userId, amount)
	return err
}

func (db DB) InsertSubBudget(userId int, category string, amount float32) error {
	const deleteQuery string = `
		DELETE FROM subBudget
		WHERE userId = ?
		AND category = ?
	`
	const insertQuery string = `
		INSERT INTO subBudget (budgetKey, category, amount)
		SELECT budgetKey, ?, ?
		FROM budget
		WHERE userId = ?
		;
	`
	_, err := db.Exec(deleteQuery, userId, category)

	if err != nil {
		return err
	}

	_, err = db.Exec(insertQuery, category, amount, userId)
	return err
}

func (db DB) GetSpendingsForLastMonth(userId int) (float32, error) {
	const transQuery string = `
		SELECT coalesce(sum(amount), 0)
		FROM transax
		JOIN account on transax.accountKey = account.accountKey
		JOIN item on item.itemKey = account.itemKey
		WHERE item.userId = ?
		AND transax.authorizedDttm > DATE('now', '-1 month')
		AND transax.amount > 0
	`

	row := db.QueryRow(transQuery, userId)
	var sum float32
	err := row.Scan(&sum)
	if err != nil {
		return -1, err
	}
	return sum, nil
}

func (db DB) GetPrimaryCategories() ([]string, error) {
	const query string = `
		SELECT DISTINCT primaryCategory
		FROM transactionCategory;
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	categories := make([]string, 0)
	defer rows.Close()
	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err != nil {
			fmt.Println(err)
		}
		categories = append(categories, cat)
	}

	return categories, nil
}
