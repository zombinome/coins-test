package db

import (
	"context"
	"time"

	"github.com/jackc/pgx"

	servErr "test/coins/errors"
)

type pgxDbContext struct {
	connectionPool *pgx.ConnPool

	connection *pgx.Conn

	transaction *pgx.Tx

	ctxCancelFn context.CancelFunc
}

// Creates new DbContext object
//	connPool    - connection pool to be used to acquire connections
// 	tranTimeout - transaction timeout
// Returns new DbContext object with acquired connection and initialized transaction,
func CreateContext(connPool *pgx.ConnPool, tranTimeout time.Duration) (DbContext, error) {
	conn, err := connPool.AcquireEx(context.Background())
	if err != nil {
		return nil, servErr.ErrDatabaseError(err)
	}

	// Making sure transaction will be interrupted by timeout
	tranCtx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	opts := pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	}

	tran, err := conn.BeginEx(tranCtx, &opts)
	if err != nil {
		ctxCancel()
		connPool.Release(conn)
		return nil, servErr.ErrDatabaseError(err)
	}

	return pgxDbContext{
		connectionPool: connPool,
		connection:     conn,
		transaction:    tran,
		ctxCancelFn:    ctxCancel,
	}, nil
}

func (db pgxDbContext) Release() error {
	// As per docs, should be save to call Rollback() even after Commit() call
	var err = db.transaction.Rollback()
	if err != nil {
		return servErr.ErrDatabaseError(err)
	}

	db.connectionPool.Release(db.connection)

	return nil
}

func (db pgxDbContext) Query(sql string, args []interface{}, mapper QueryMapper) error {
	rows, err := db.transaction.Query(sql, args...)
	if err != nil {
		return err
	}

	defer rows.Close()

	return mapper(rows)
}

func (db pgxDbContext) Execute(sql string, args ...interface{}) (int64, error) {
	tag, err := db.transaction.Exec(sql, args...)
	if err != nil {
		return -1, err
	}

	return tag.RowsAffected(), nil
}

func (db pgxDbContext) Save() error {
	var err = db.transaction.Commit()
	if err != nil {
		return servErr.ErrDatabaseError(err)
	}

	return nil
}
