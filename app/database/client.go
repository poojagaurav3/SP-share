package database

import (
	"context"
	"fmt"
	"sync"

	pg "github.com/go-pg/pg"
)

const (
	// pgURL = "postgres://postgres:docker@localhost:5432/spdb?sslmode=disable"
	pgURL = "postgres://postgres:0HYmmozpAU1M4vxoCaSj@sp-share-db.c2ffomew0jtd.us-east-2.rds.amazonaws.com:5432/spdb?sslmode=disable"
)

var client *Client
var initOnce sync.Once

// Client is the struct for database connectivity
type Client struct {
	pgClient *pg.DB
}

// GetPGClient returns the pg-client contained in Client
func (c *Client) GetPGClient() *pg.DB {
	if c == nil {
		return nil
	}
	return c.pgClient
}

// GetClient initializes and returns a new database client
func GetClient() (*Client, error) {
	opt, err := pg.ParseURL(pgURL)
	if err != nil {
		return nil, err
	}

	initOnce.Do(func() {
		client = &Client{
			pgClient: pg.Connect(opt),
		}
		// client.pgClient.AddQueryHook(dbLogger{})
	})

	return client, nil
}

type dbLogger struct{}

func (d dbLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	return c, nil
}

func (d dbLogger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	fmt.Println(q.FormattedQuery())
	return nil
}
