package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nrobins00/personal-finance/internal/types"
	"github.com/plaid/plaid-go/plaid"
)

type DB struct {
	*sql.DB
}

func CreateDatabase(filename string) *DB {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	/*const dropUserTable string = `
		DROP TABLE IF EXISTS user;
	`
	if _, err := db.Exec(dropUserTable); err != nil {
		log.Fatal(err)
	}*/

	const createUserTable string = `
		CREATE TABLE IF NOT EXISTS user (
			userId INTEGER NOT NULL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
        );
    `

	if _, err := db.Exec(createUserTable); err != nil {
		log.Fatal("createUserTable", err)
	}

	const createItemTable string = `
            CREATE TABLE IF NOT EXISTS item (
				itemKey INTEGER NOT NULL PRIMARY KEY,
                itemId TEXT NOT NULL,
                userId INTEGER NOT NULL,
                accessKey TEXT,
				cursor TEXT,
                FOREIGN KEY(userId) REFERENCES user(userId)
            );
	`
	if _, err := db.Exec(createItemTable); err != nil {
		log.Fatal("createItemTable", err)
	}

	const createTransactionTable string = `
		CREATE TABLE IF NOT EXISTS transax (
			transactionKey INTEGER NOT NULL PRIMARY KEY,
			transactionId TEXT NOT NULL,
			accountKey INTEGER,
			amount REAL,
			categoryId INTEGER,
			authorizedDttm TEXT,
			FOREIGN KEY(categoryId) REFERENCES transactionCategory(categoryId),
			FOREIGN KEY(accountKey) REFERENCES account(accountKey)
		);
	`
	if _, err := db.Exec(createTransactionTable); err != nil {
		log.Fatal("createTransactionTable", err)
	}
	//const createTransactionCategoryTable string = `
	//	CREATE TABLE IF NOT EXISTS transactionCategory (
	//		categoryId INTEGER NOT NULL PRIMARY KEY,
	//		primary TEXT,
	//		detailed TEXT,
	//	);
	//`

	const createAccountTable string = `
		CREATE TABLE IF NOT EXISTS account (
			accountKey INTEGER NOT NULL PRIMARY KEY,
			accountId TEXT NOT NULL,
			userId INTEGER,
			itemKey INTEGER,
			mask TEXT,
			name TEXT,
			availableBalance REAL,
			currentBalance REAL,
			lastUpdatedDttm TEXT,
			FOREIGN KEY(userId) REFERENCES user(userId),
			FOREIGN KEY(itemKey) REFERENCES item(itemKey)
		);
	`
	if _, err := db.Exec(createAccountTable); err != nil {
		log.Fatal("createAccountTable", err)
	}
	return &DB{db}
}

func (db *DB) CreateUser(username string) (int64, error) {
	const createUser = `
		INSERT INTO user (username, password)
		VALUES (?, ?)
	`
	result, err := db.Exec(createUser, username, "x")
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

func (db DB) GetItemsForUser(userId int64) ([]types.Item, error) {
	const query string = `
        SELECT itemId, accessKey FROM item where userId = ?
    `
	rows, err = db.Query(query, userId)
	if err != nil {
		return []types.Item{}, err
	}
	defer rows.Close()
	items := make([]types.Item, 0)
	for rows.Next() {
		var itemId, accessKey string
		if err := rows.Scan(&itemId, &accessKey); err != nil {
			log.Fatal(err)
		}
		items = append(items, types.Item{
			ItemId:    itemId,
			AccessKey: accessKey,
		})
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return items
}

func (db *DB) UpdateTransactions(itemId string, added, modified []plaid.Transaction, removed []plaid.RemovedTransaction, cursor string) error {
	const updateCursor string = `
		UPDATE item 
		SET cursor = ?
		WHERE itemId = ?
	`

	_, err := db.Exec(cursor, itemId)
	if err != nil {
		return err
	}

	const addTransaction string = `
		INSERT INTO transaction (transactionId, accountId, amount, categoryKey, authorizedDttm)
		VALUES (?, ?, ?, ?, ?)
	`

	const findCategoryKey string = `
		SELECT categoryKey
		FROM transactionCategory
		WHERE primaryName = ? AND detailedName = ?
	`
	for _, val := range added {
		//find category key
		category := val.GetPersonalFinanceCategory()
		row := db.QueryRow(findCategoryKey, category.Primary, category.Detailed)
		var catKey int
		err := row.Scan(&catKey)
		if err != nil {
			log.Printf("Couldn't find category key for values %s and %s",
				category.Primary, category.Detailed)
		}
		//insert transaction
		_, err = db.Exec(addTransaction, val.GetTransactionId(), val.GetAccountId(),
			val.GetAmount(), catKey, val.GetAuthorizedDatetime())

		if err != nil {
			log.Println(err.Error())
			return err
		}
	}
	return nil
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

func (db *DB) InsertAccounts(userId int64, accounts []types.Account) {
	if len(accounts) == 0 {
		return
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
			acc.AccountId, userId, acc.ItemKey, acc.Mask, acc.Name,
			acc.AvailableBalance, acc.CurrentBalance,
			acc.LastUpdatedDttm.String())
		query += valueString
		if accountNum < len(accounts)-1 {
			query += ", "
		}
	}
	fmt.Println(query)

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("insert accounts", err)
	}
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
		var accessKey, itemId string
		var cursor sql.NullString
		if err := rows.Scan(&itemKey, &itemId, &accessKey, &cursor); err != nil {
			log.Fatal(err)
		}
		items = append(items, types.Item{
			ItemKey:   itemKey,
			ItemId:    itemId,
			AccessKey: accessKey,
			Cursor:    cursor.String,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
