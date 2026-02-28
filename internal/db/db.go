package db

import (
	"context"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

type User struct {
	ID           int    `db:"id" json:"id"`
	Email        string `db:"email" json:"email"`
	PasswordHash string `db:"password_hash" json:"-"`
	APIToken     string `db:"api_token" json:"api_token"`
	ClientToken  string `db:"client_token" json:"client_token"`
}

type APIKeyRecord struct {
	ID            int    `db:"id"`
	APIKey        string `db:"api_key"`
	AllowedModels string `db:"allowed_models"` // comma separated string e.g. "gpt-3.5-turbo,gpt-4"
	RPM           int    `db:"rpm"`            // requests per minute limit
}

// AllowedModelList returns a slice of allowed models
func (a *APIKeyRecord) AllowedModelList() []string {
	if a.AllowedModels == "" || a.AllowedModels == "*" {
		return []string{"*"}
	}
	return strings.Split(a.AllowedModels, ",")
}

func Connect(dsn string) (*DB, error) {
	conn, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &DB{DB: conn}, nil
}

func (db *DB) InitializeSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS api_keys (
		id SERIAL PRIMARY KEY,
		api_key VARCHAR(100) UNIQUE NOT NULL,
		allowed_models VARCHAR(255) NOT NULL DEFAULT '*',
		rpm INTEGER NOT NULL DEFAULT 60
	);

	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		api_token VARCHAR(100) UNIQUE NOT NULL,
		client_token VARCHAR(100) UNIQUE NOT NULL
	);
	`
	_, err := db.Exec(schema)
	return err
}

func (db *DB) GetAPIKey(ctx context.Context, key string) (*APIKeyRecord, error) {
	var record APIKeyRecord
	err := db.GetContext(ctx, &record, "SELECT * FROM api_keys WHERE api_key=$1", key)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (db *DB) CreateUser(ctx context.Context, email, pwHash, apiToken, clientToken string) error {
	_, err := db.ExecContext(ctx, "INSERT INTO users (email, password_hash, api_token, client_token) VALUES ($1, $2, $3, $4)",
		email, pwHash, apiToken, clientToken)

	if err == nil {
		// Auto-register API key to api_keys table for standard flow limits
		_, err = db.ExecContext(ctx, "INSERT INTO api_keys (api_key, allowed_models) VALUES ($1, '*')", apiToken)
	}
	return err
}

func (db *DB) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := db.GetContext(ctx, &u, "SELECT * FROM users WHERE email=$1", email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
