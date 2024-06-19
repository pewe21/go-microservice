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
	// connStr := "postgres://postgress:postgress@localhost/pqgotest?sslmode=disable"
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", userDB, passDB, hostDB, databaseDB)
	log.Println(connStr)
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
	if err := s.createPostTable(); err != nil {
		log.Fatal(err)
	}
}

func (s *PostgresStorage) createPostTable() error {
	_, err := s.db.Exec(`
        CREATE TABLE IF NOT EXISTS posts (
            id TEXT PRIMARY KEY,
            image TEXT,
            body TEXT NOT NULL,
            idUser TEXT NOT NULL,
            username TEXT NOT NULL,
            name TEXT NOT NULL,
            profile TEXT NOT NULL,
            totalLikes INTEGER DEFAULT 0 NOT NULL,
            totalReplies INTEGER DEFAULT 0 NOT NULL,
            
            createdAt INTEGER NOT NULL,
            updatedAt INTEGER NOT NULL,
            deletedAt INTEGER
        )`)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) UpdatePostBody(id, body, userid string) error {
	// psql use $1, $2, $3, etc. instead of ? as placeholder
	// http://go-database-sql.org/prepared.html#parameter-placeholder-syntax
	stmt, err := s.db.Prepare(`
        UPDATE posts
        SET
            body = $1,
            updatedAt = $2
        WHERE 
            id = $3
            AND idUser = $4
            AND deletedAt IS NULL
        `)
	if err != nil {
		return err
	}

	unixEpoch := time.Now().Unix()

	_, err = stmt.Exec(body, unixEpoch, id, userid)
	if err != nil {
		return err
	}

	return nil
}

// listPostByUser --> nampilin list post yang dibuat oleh user
func (s *PostgresStorage) ListPostByUser(cursor int64, userId string, limit int32, posts *[]Post) error {
	queryStr := `
        SELECT
            id,
            image,
            body,
            idUser,
            username,
            name,
            profile,
            totalLikes,
            totalReplies,
            createdAt,
            updatedAt
        FROM
            posts 
        WHERE
            idUser = $1
            AND deletedAt IS NULL
            AND createdAt < $2
        ORDER BY
            createdAt DESC
        LIMIT $3`

	stmt, err := s.db.Prepare(queryStr)
	if err != nil {
		log.Println("Stmt error:", err)
		return err
	}

	defer stmt.Close()

	if cursor == 0 {
		cursor = 922337203685477
	}

	rows, err := stmt.Query(userId, cursor, limit)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(
			&post.Id,
			&post.Image,
			&post.Body,
			&post.IdUser,
			&post.Username,
			&post.Name,
			&post.Profile,
			&post.TotalLikes,
			&post.TotalReplies,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return err
		}
		*posts = append(*posts, post)

	}

	return nil
}

// listPosts --> nampilin list post
func (s *PostgresStorage) ListPost(cursor int64, limit int32, posts *[]Post) error {
	queryStr := `
        SELECT
            id,
            image,
            body,
            idUser,
            username,
            name,
            profile,
            totalLikes,
            totalReplies,
            createdAt,
            updatedAt
        FROM
            posts 
        WHERE
            deletedAt IS NULL
            AND createdAt < $1
        ORDER BY
            createdAt DESC
        LIMIT $2
        `
	stmt, err := s.db.Prepare(queryStr)
	if err != nil {
		log.Println("Error when creating stmt in listPosts", err)
	}
	defer stmt.Close()

	if cursor == 0 {
		cursor = 922337203685477
	}

	rows, err := stmt.Query(cursor, limit)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(
			&post.Id,
			&post.Image,
			&post.Body,
			&post.IdUser,
			&post.Username,
			&post.Name,
			&post.Profile,
			&post.TotalLikes,
			&post.TotalReplies,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return err
		}
		*posts = append(*posts, post)

	}

	return nil
}

// getPostById --> nampilin satu post
func (s *PostgresStorage) GetPostById(id string, post *Post) error {
	stmt, err := s.db.Prepare(`
        SELECT
            id,
            image,
            body,
            idUser,
            username,
            name,
            profile,
            totalLikes,
            totalReplies,
            createdAt,
            updatedAt
        FROM
            posts 
        WHERE
            id = $1
            AND deletedAt IS NULL
        LIMIT 1
        `)
	if err != nil {
		return err
	}

	defer stmt.Close()

	if err := stmt.QueryRow(id).Scan(
		&post.Id,
		&post.Image,
		&post.Body,
		&post.IdUser,
		&post.Username,
		&post.Name,
		&post.Profile,
		&post.TotalLikes,
		&post.TotalReplies,
		&post.CreatedAt,
		&post.UpdatedAt,
	); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) DeletePostById(id, userId string) error {
	stmt, err := s.db.Prepare(`
        UPDATE 
            posts
        SET 
            deletedAt = $1
        WHERE 
            id = $2
            AND idUser = $3
            AND deletedAt IS NULL
        `)
	if err != nil {
		return err
	}

	defer stmt.Close()

	unixEpoch := time.Now().Unix()

	if _, err := stmt.Exec(unixEpoch, id, userId); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) UpdateUserDetail(idUser, profile, name string) error {
	log.Println(idUser, profile, name)
	stmt, err := s.db.Prepare(`
        UPDATE posts
        SET
            name = $1,
            profile = $2,
            updatedAt = $3
        WHERE 
            idUser = $4
            AND deletedAt IS NULL
        `)
	if err != nil {
		return err
	}

	defer stmt.Close()

	unixEpoch := time.Now().Unix()

	if _, err := stmt.Exec(name, profile, unixEpoch, idUser); err != nil {
		return err
	}

	log.Println("done updating user detail in post service")

	return nil
}

func (s *PostgresStorage) CreatePost(id, image, body, idUser, username, name, profile string) error {
	stmt, err := s.db.Prepare(`
        INSERT INTO posts (
            id,
            image,
            body,
            idUser,
            username,
            name,
            profile,
            
            createdAt,
            updatedAt
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
        `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	unixEpoch := time.Now().Unix()

	if _, err := stmt.Exec(
		id,
		image,
		body,
		idUser,
		username,
		name,
		profile,
		unixEpoch,
		unixEpoch); err != nil {
		return err
	}

	return nil
}
