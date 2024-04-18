package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
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

	// const insertAccessKey string = `
	//     INSERT INTO
	//
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
	var cursor string
	err := db.QueryRow(cursorQuery).Scan(&cursor)
	if err != nil {
		return "", err
	}
	return cursor, nil
}

type Account struct {
	Base    plaid.AccountBase
	ItemKey int
}

func (db *DB) InsertAccounts(userId int64, accounts []Account) {
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
			acc.Base.AccountId, userId, acc.ItemKey, acc.Base.GetMask(), acc.Base.GetName(),
			*acc.Base.GetBalances().Available.Get(), *acc.Base.GetBalances().Current.Get(),
			acc.Base.Balances.GetLastUpdatedDatetime())
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
