package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nrobins00/personal-finance/internal/types"
)

type DB struct {
	*sql.DB
}

func (db *DB) CreateUser(username string, password string) (int64, error) {
	const createUser = `
		INSERT INTO user (username, password)
		VALUES (?, ?)
	`
	result, err := db.Exec(createUser, username, password)
	if err != nil {
		return -1, err
	}
	return result.LastInsertId()
}

func (db *DB) CreateItem(user int64, itemId string, accessKey string) (int64, error) {
	const storeToken = `
		INSERT INTO item (itemId, userId, accessKey)
		VALUES (?, ?, ?)
	`
	result, err := db.Exec(storeToken, itemId, user, accessKey)
	if err != nil {
		return -1, err
	}
	insertedId, err := result.LastInsertId()
	if err != nil {
		return -1, nil
	}
	return insertedId, nil
}

func (db *DB) UpdateTransactions(itemId string, added, modified, removed []types.Transaction, cursor string) error {

	const updateCursor string = `
		UPDATE item 
		SET cursor = ?
		WHERE itemId = ?
	`

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = db.Exec(updateCursor, cursor, itemId)
	if err != nil {
		return err
	}

	const addTransaction string = `
		INSERT INTO transax (transactionId, amount, authorizedDttm, accountKey, category)
		VALUES
	`

	accountKeyMap, err := db.buildAccountKeyMap(added)
	if err != nil {
		return err
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString(addTransaction)
	validTransactionFound := false
	for ix, val := range added {
		accountKey, found := accountKeyMap[val.AccountId]
		if !found {
			fmt.Printf("no account found for transaction: %+v\n", val)
			continue
		}
		validTransactionFound = true

		valString := fmt.Sprintf("\n('%v', %v, '%v', %v, '%v')", val.TransactionId, val.Amount,
			val.AuthorizedDttm, //.String()
			accountKey, val.CategoryDetailed)
		queryBuilder.WriteString(valString)

		if ix < len(added)-1 {
			queryBuilder.WriteString(",")
		}

	}
	if validTransactionFound {
		query := queryBuilder.String()
		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}

	//TODO: updates and deletes

	tx.Commit()

	return nil
}

func (db DB) GetTransactionsForUser(userId int, limit int, offset int) ([]types.Transaction, error) {
	const transQuery string = `
		SELECT transactionId, 
		    amount, 
		    category.PrimaryName, 
		    category.detailedName, 
		    authorizedDttm, 
		    account.name
		FROM transax
		JOIN account on account.accountKey = transax.accountKey
		JOIN item on item.itemKey = account.itemKey
		JOIN transactionCategory category
		    ON category.detailedName = transax.category
		WHERE item.userId = ?
		--AND transax.authorizedDttm > DATE('now', '-1 month')
	`
	var rows *sql.Rows
	var err error
	if limit > 0 {
		limitQuery := transQuery + "\nLIMIT ? OFFSET ?"
		rows, err = db.Query(limitQuery, userId, limit, offset)

	} else {
		rows, err = db.Query(transQuery, userId)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	transactions := make([]types.Transaction, 0)
	for rows.Next() {
		var (
			transactionId    string
			amount           float32
			primaryCategory  string
			detailedCategory string
			authorizedDttm   string
			accountName      string
		)
		err := rows.Scan(&transactionId,
			&amount,
			&primaryCategory,
			&detailedCategory,
			&authorizedDttm,
			&accountName)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, types.Transaction{
			TransactionId:    transactionId,
			Amount:           amount,
			CategoryPrimary:  primaryCategory,
			CategoryDetailed: detailedCategory,
			AuthorizedDttm:   authorizedDttm,
			AccountName:      accountName,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (db *DB) GetUserId(username, password string) (int, error) {
	const userQuery string = `
		SELECT userId FROM user
		WHERE username = ? AND password = ?;
	`
	var userId int
	err := db.QueryRow(userQuery, username, password).Scan(&userId)
	return userId, err
}

func (db *DB) GetLatestCursor(itemId string) (string, error) {
	const cursorQuery string = `
		SELECT cursorId FROM item
		WHERE itemId = ?
	`
	var cursor sql.NullString
	err := db.QueryRow(cursorQuery).Scan(&cursor)
	if err != nil {
		return "", err
	}
	var retString string
	if cursor.Valid {
		retString = cursor.String
	}
	return retString, nil
}

func (db *DB) InsertAccounts(userId int64, itemKey int, accounts []types.Account) error {
	if len(accounts) == 0 {
		return fmt.Errorf("No accounts passed")
	}
	query := `
		INSERT INTO account (
			accountId, 
			userId, 
			itemKey, 
			mask,
			name,
			availableBalance, 
			currentBalance, 
			lastUpdatedDttm
		) VALUES
	`
	for accountNum, acc := range accounts {
		valueString := fmt.Sprintf("('%s', %d, %d, '%s', '%s', %f, %f, '%s')",
			acc.AccountId, userId, itemKey, acc.Mask, acc.Name,
			acc.AvailableBalance, acc.CurrentBalance,
			acc.LastUpdatedDttm.String())
		query += valueString
		if accountNum < len(accounts)-1 {
			query += ", "
		}
	}
	fmt.Println(query)

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

/*

	rows, err := db.Query(accQuery, key)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var accountId int
		var name string
		var availBal, curBal float32
		if err := rows.Scan(&accountId, &name, &availBal, &curBal); err != nil {
			rows.Close()
			log.Fatal(err)
		}
		accounts = append(accounts, account{accountId, name, availBal, curBal})
	}
	rows.Close()
*/

func (db *DB) GetAllAccounts(itemKey int) ([]types.Account, error) {
	const accQuery string = `
		SELECT accountId, name, availableBalance, currentBalance
		FROM account
		WHERE itemKey = ?
	`
	rows, err := db.Query(accQuery, itemKey)
	if err != nil {
		return nil, err
	}
	accounts := make([]types.Account, 0)
	defer rows.Close()
	for rows.Next() {
		var accountId, name string
		var availableBalance, currentBalance float32
		if err := rows.Scan(&accountId, &name, &availableBalance, &currentBalance); err != nil {
			return nil, err
		}
		accounts = append(accounts, types.Account{
			AccountId:        accountId,
			Name:             name,
			AvailableBalance: availableBalance,
			CurrentBalance:   currentBalance,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (db DB) GetAllItemsForUser(userId int64) ([]types.Item, error) {
	const itemQuery string = `
		SELECT itemKey, itemId, accessKey, cursor FROM item where userId = ?
	`

	rows, err := db.Query(itemQuery, userId)
	if err != nil {
		return nil, err
	}
	items := make([]types.Item, 0)
	defer rows.Close()
	for rows.Next() {
		var itemKey int
		var accessToken, itemId string
		var cursor sql.NullString
		if err := rows.Scan(&itemKey, &itemId, &accessToken, &cursor); err != nil {
			log.Fatal(err)
		}
		items = append(items, types.Item{
			ItemKey:     itemKey,
			ItemId:      itemId,
			AccessToken: accessToken,
			Cursor:      cursor.String,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (db DB) buildAccountKeyMap(transactions []types.Transaction) (map[string]int, error) {
	const baseQuery string = `
		SELECT accountKey, accountId
		FROM account
		WHERE accountId in (`

	accountIds := make(map[string]bool)
	for _, val := range transactions {
		accountIds[val.AccountId] = true
	}

	var accountQuerySb strings.Builder
	accountQuerySb.WriteString(baseQuery)
	ids := make([]any, 0)
	for key := range accountIds {
		accountQuerySb.WriteString("?, ")
		ids = append(ids, key)
	}
	accountQuery := accountQuerySb.String()
	accountQuery = strings.TrimSuffix(accountQuery, ", ")
	accountQuery = accountQuery + ")"

	fmt.Println("idString: ", accountQuery)

	rows, err := db.Query(accountQuery, ids...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	accountKeyIdMap := make(map[string]int)
	for rows.Next() {
		var accountKey int
		var accountId string
		if err := rows.Scan(&accountKey, &accountId); err != nil {
			return nil, err
		}
		fmt.Println(accountKey, " ", accountId)
		accountKeyIdMap[accountId] = accountKey
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accountKeyIdMap, nil
}

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
	const insertQuery string = `
		INSERT INTO budget (userId, amount)
		VALUES (?, ?);
	`
	_, err := db.Exec(insertQuery, userId, amount)
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
	`

	row := db.QueryRow(transQuery, userId)
	var sum float32
	err := row.Scan(&sum)
	if err != nil {
		return -1, err
	}
	return sum, nil
}
