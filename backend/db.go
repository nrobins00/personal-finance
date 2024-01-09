package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func createDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	const createUserTable string = `
        CREATE TABLE IF NOT EXISTS user (
            id INTEGER NOT NULL PRIMARY KEY
        );   
    `

	if _, err := db.Exec(createUserTable); err != nil {
		log.Fatal(err)
	}

	const createItemTable string = `
            CREATE TABLE IF NOT EXISTS item (
                id INTEGER NOT NULL PRIMARY KEY,
                userId INTEGER NOT NULL,
                accessKey TEXT,
                FOREIGN KEY(userId) REFERENCES user(id)
            );
    `
	if _, err := db.Exec(createItemTable); err != nil {
		log.Fatal(err)
	}
	return db

	// const insertAccessKey string = `
	//     INSERT INTO
	// `
}

func createItem(user string, accessKey string) (int64, error) {
    const storeToken = `
        INSERT INTO item (userId, accessKey)
        VALUES (?, ?)
    `
    result, err := db.Exec(storeToken, user, accessKey)
    if err != nil {
        return -1, err
    }
    insertedId, err := result.LastInsertId()
    if err != nil {
        return -1, nil
    }
    return insertedId, nil
}

