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

	/*const dropUserTable string = `
		DROP TABLE IF EXISTS user;
	`
	if _, err := db.Exec(dropUserTable); err != nil {
		log.Fatal(err)
	}*/

	const createUserTable string = `
		CREATE TABLE IF NOT EXISTS user (
			userId INTEGER NOT NULL PRIMARY KEY,
			username VARCHAR NOT NULL UNIQUE,
			password VARCHAR NOT NULL
        );
    `

	if _, err := db.Exec(createUserTable); err != nil {
		log.Fatal(err)
	}

	const createItemTable string = `
            CREATE TABLE IF NOT EXISTS item (
                itemId INTEGER NOT NULL PRIMARY KEY,
                userId INTEGER NOT NULL,
                accessKey TEXT,
                FOREIGN KEY(userId) REFERENCES user(userId)
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

func createUser(username string) (int64, error) {
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

func createItem(user int64, accessKey string) (int64, error) {
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

func getUserId(username, password string) (int, error) {
	const userQuery string = `
		SELECT userId FROM user
		WHERE username = ? AND password = ?;
	`
	var userId int
	err := db.QueryRow(userQuery, username, password).Scan(&userId)
	return userId, err
}
