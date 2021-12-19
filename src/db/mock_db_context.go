package db

import (
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"

	servErr "test/coins/errors"
)

// Wrapper around sql.Rows, which is requred to normalize difference in contract
// between sql.Rows and pgx.Rows
type sqlRowsWrapper struct {
	rows *sql.Rows
}

func (wr sqlRowsWrapper) Close() {
	wr.rows.Close()
}

func (wr sqlRowsWrapper) Next() bool {
	return wr.rows.Next()
}

func (wr sqlRowsWrapper) Scan(args ...interface{}) error {
	return wr.rows.Scan(args...)
}

type mockDbContext struct {
	Mock *sqlmock.Sqlmock

	db          *sql.DB
	transaction *sql.Tx
}

// Creates new mockDbContext that is used to mock db interaction with database
// 	setup - setup function that is called so mock can be set up
// Returns created mockDbContext
func CreateMockDbContext(setup func(mock sqlmock.Sqlmock)) (*mockDbContext, error) {
	sqlDb, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	setup(mock)

	tran, err := sqlDb.Begin()
	if err != nil {
		return nil, err
	}

	return &mockDbContext{
		Mock:        &mock,
		db:          sqlDb,
		transaction: tran,
	}, nil
}

func (dbContext mockDbContext) Release() error {
	return dbContext.transaction.Rollback()
}

func (dbContext mockDbContext) Save() error {
	return dbContext.transaction.Commit()
}

func (dbContext mockDbContext) Query(sql string, sqlParams []interface{}, mapper QueryMapper) error {
	rows, err := dbContext.db.Query(sql, sqlParams...)
	if err != nil {
		return servErr.ErrDatabaseError(err)
	}

	return mapper(sqlRowsWrapper{rows})
}

func (dbContext mockDbContext) Execute(sql string, sqlParams ...interface{}) (int64, error) {
	tag, err := dbContext.db.Exec(sql, sqlParams...)
	if err != nil {
		return -1, servErr.ErrDatabaseError(err)
	}

	rowsAffected, err := tag.RowsAffected()
	if err != nil {
		return -1, servErr.ErrDatabaseError(err)
	}

	return rowsAffected, nil
}
