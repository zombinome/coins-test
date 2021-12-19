package db

import "github.com/jackc/pgx"

// Initializes new pgx connection pool from connection string
//	cs - connection string
// Returns connection pool object
func NewConnectionPool(cs string) (*pgx.ConnPool, error) {
	connCfg, err := pgx.ParseConnectionString(cs)
	if err != nil {
		return nil, err
	}

	poolCfg := pgx.ConnPoolConfig{
		ConnConfig:     connCfg,
		MaxConnections: 16,
	}

	return pgx.NewConnPool(poolCfg)
}
