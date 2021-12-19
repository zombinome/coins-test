package db

type QueryResultRows interface {
	Close()

	Next() bool

	Scan(values ...interface{}) error
}

type QueryMapper = func(rows QueryResultRows) error

// DbContext is wrapper around connection and transaction
// used to avoid copy-paste of infrastructure code needed to begin/commit transaction
// DbContext should be used for queries that should be creaeted in scope of single transaction
// It is not thread safe
type DbContext interface {
	// Releases DbContext, rolliing back transaction (if not committed yet)
	// and returning current connection to connection pool
	Release() error

	// Commits current transaction
	Save() error

	// Executes sql query that is expected to return some data from database
	// 	sql         - sql query
	//  sqlParams   - sql parameters to be used with sql query
	//  queryMapper - mapper func used to read and map retrieved rows
	// Returns error if some error occured. If there is some database-related error,
	// it would be wrapped in ServiceError (DbError). If queryMapper return some error - it would be passed through as is.
	Query(sql string, sqlParams []interface{}, mapper QueryMapper) error

	// Executes sql query that is not expected to return any data from database
	// sqlParams   - sql parameters to be used with sql query
	// Returns number of rows affected by query
	Execute(sql string, sqlParams ...interface{}) (int64, error)
}
