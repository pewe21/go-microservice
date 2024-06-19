package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() *PostgresStorage {
	userDB := os.Getenv("POSTGRES_USER")
	passDB := os.Getenv("POSTGRES_PASSWORD")
	databaseDB := os.Getenv("POSTGRES_DB")
	hostDB := os.Getenv("POSTGRES_HOST")
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", userDB, passDB, hostDB, databaseDB)
	log.Println(connStr)
	// connStr := "postgres://postgress:postgress@localhost/pqgotest?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Cannot establish connection to database: ", err.Error())
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Cannot ping to database: ", err.Error())
	}

	log.Println("Connected to database")

	return &PostgresStorage{
		db: db,
	}
}

func (s *PostgresStorage) Init() {
	if err := s.createUserTable(); err != nil {
		log.Fatal(err)
	}
}

func (s *PostgresStorage) createUserTable() error {
	_, err := s.db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY,
            username TEXT UNIQUE NOT NULL,
            name TEXT NOT NULL,
            hashPassword TEXT NOT NULL,
            profile TEXT NOT NULL,
            totalFollower INTEGER DEFAULT 0 NOT NULL,
            totalFollowing INTEGER DEFAULT 0 NOT NULL,
            
            createdAt INTEGER NOT NULL,
            updatedAt INTEGER NOT NULL,
            deletedAt INTEGER
        )`)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) UpdateProfileById(profileUrl, id string) error {
	stmt, err := s.db.Prepare(`
        UPDATE users
        SET
            profile = ?
        WHERE
            id = ?
        `)
	if err != nil {
		return err
	}

	if _, err := stmt.Exec(profileUrl, id); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) IncrementFollowerById(id string) error {
	stmt, err := s.db.Prepare(`
        UPDATE users
        SET
            totalFollower = totalFollower + 1
        WHERE
            id = ?
        `)
	if err != nil {
		return err
	}

	if _, err := stmt.Exec(id); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) DecrementFollowerById(id string) error {
	stmt, err := s.db.Prepare(`
        UPDATE users
        SET
            totalFollower = totalFollower - 1
        WHERE
            id = ?
        `)
	if err != nil {
		return err
	}

	if _, err := stmt.Exec(id); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) IncrementFollowingById(id string) error {
	stmt, err := s.db.Prepare(`
        UPDATE users
        SET
            totalFollowing = totalFollowing + 1
        WHERE
            id = ?
        `)
	if err != nil {
		return err
	}

	if _, err := stmt.Exec(id); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) DecrementFollowingById(id string) error {
	stmt, err := s.db.Prepare(`
        UPDATE users
        SET
            totalFollowing = totalFollowing - 1
        WHERE
            id = ?
        `)
	if err != nil {
		return err
	}

	if _, err := stmt.Exec(id); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) CreateUser(id, username, name, hashPassword, profile string, createdAt, updatedAt int64) error {
	// psql use $1, $2, $3, etc. instead of ? as placeholder
	// http://go-database-sql.org/prepared.html#parameter-placeholder-syntax
	stmt, err := s.db.Prepare(`
        INSERT INTO users (
        id,
        username,
        name,
        hashPassword,
        profile,
        createdAt,
        updatedAt
        ) VALUES ($1,$2,$3,$4,$5,$6,$7);
        `)
	if err != nil {
		return err
	}

	defer stmt.Close()
	log.Println("Create user excec cek error")
	log.Println(id, username, name, hashPassword, profile, createdAt, updatedAt)
	if _, err := stmt.Exec(
		id,
		username,
		name,
		hashPassword,
		profile,
		createdAt,
		updatedAt); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) GetUserPasswordByUsername(username string, user *User) error {
	stmt, err := s.db.Prepare(`
        SELECT 
        id,
        username,
        hashPassword,
        createdAt,
        updatedAt 
        FROM users WHERE username = $1 AND deletedAt IS NULL`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	if err := stmt.QueryRow(username).Scan(
		&user.Id,
		&user.Username,
		&user.HashPassword,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) GetUserPasswordById(id string, user *User) error {
	stmt, err := s.db.Prepare(`
        SELECT 
        id,
        hashPassword,
        createdAt,
        updatedAt 
        FROM users WHERE id = $1 AND deletedAt IS NULL`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	if err := stmt.QueryRow(id).Scan(
		&user.Id,
		&user.HashPassword,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) UpdateUserPasswordById(newPassword, id string) error {
	stmt, err := s.db.Prepare(`
        UPDATE users
        SET 
            hashPassword = $1,
            updatedAt = $2
        WHERE id = $3`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	unixEpoch := time.Now().Unix()

	if _, err := stmt.Exec(newPassword, unixEpoch, id); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) DeleteUserById(id string) error {
	stmt, err := s.db.Prepare(`
        UPDATE users
        SET deletedAt = $1
        WHERE id = $2`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	unixEpoch := time.Now().Unix()

	if _, err := stmt.Exec(unixEpoch, id); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) UpdateUserNameAndProfile(name, profile, id string) error {
	stmt, err := s.db.Prepare(`
        UPDATE users
        SET 
            name = $1,
            profile = $2,
            updatedAt = $3
        WHERE id = $4`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	unixEpoch := time.Now().Unix()

	if _, err := stmt.Exec(name, profile, unixEpoch, id); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) GetUserByUsername(username string, user *ReturnUser) error {
	stmt, err := s.db.Prepare(`
        SELECT 
        id,
        username,
        name,
        profile,
        createdAt,
        updatedAt 
        FROM users WHERE username = $1 AND deletedAt IS NULL`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	if err := stmt.QueryRow(username).Scan(
		&user.Id,
		&user.Username,
		&user.Name,
		&user.Profile,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) GetUserById(id string, user *ReturnUser) error {
	stmt, err := s.db.Prepare(`
        SELECT 
        id,
        username,
        name,
        profile,
        createdAt,
        updatedAt
        FROM users WHERE id = $1`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	if err := stmt.QueryRow(id).Scan(
		&user.Id,
		&user.Username,
		&user.Name,
		&user.Profile,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return err
	}

	return nil
}
