package pgconn

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Config struct {
	Host         string
	Port         uint
	DB           string
	Username     string
	Password     string
	MaxIdleConns int
	MaxOpenConns int
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		c.Host, c.Port, c.DB, c.Username, c.Password)
}

func NewDB(ctx context.Context, config *Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.DSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
