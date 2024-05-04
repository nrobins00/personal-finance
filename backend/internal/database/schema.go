package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

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
			category TEXT,
			authorizedDttm TEXT,
			FOREIGN KEY(category) REFERENCES transactionCategory(detailedName),
			FOREIGN KEY(accountKey) REFERENCES account(accountKey)
		);
	`
	if _, err := db.Exec(createTransactionTable); err != nil {
		log.Fatal("createTransactionTable", err)
	}

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

	const createBudgetTable string = `
		CREATE TABLE IF NOT EXISTS budget (
			budgetKey INTEGER NOT NULL PRIMARY KEY,
			userId INTEGER NOT NULL,
			amount REAL,
			FOREIGN KEY(userId) REFERENCES user(userId)
		);
	`
	if _, err := db.Exec(createBudgetTable); err != nil {
		log.Fatal("createBudgetTable", err)
	}

	return &DB{db}
}
