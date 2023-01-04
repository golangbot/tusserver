package database

import (
	"database/sql"
	"time"
)

var (
	MaxRetries   uint = 10
	MaxRetryTime uint = 2 // assume in seconds
)

// check retry logic with timeout
// time.After returns channel
func GetConnection(dsn string) (*sql.DB, error) {
	var count uint = 1
retry:
	db, err := sql.Open("postgres", dsn)
	if err != nil && count <= MaxRetries {
		time.Sleep(time.Second * time.Duration(MaxRetryTime))
		count++
		goto retry
	}
	return db, err
}

// docker run -d --name pg -p 5432:5432 -e POSTGRES_PASSWORD=contactsdb_admin -e POSTGRES_USER=contactsdb_user -e POSTGRES_DB=contactsdb postgres

// docker run -d --name pgadmin -p --network mynetwork 58080:8080 adminer

//
