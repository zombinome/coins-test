package account

import (
	"test/coins/db"
	servErr "test/coins/errors"
)

// Type alias for sql parameters array
type sqlParams = []interface{}

// Account service. Incapsulates operations with accounts
type AccountService interface {
	// Returns list of accounts that is existing in database
	ListAccounts() ([]Account, error)
}

// Account service implementation
type accountService struct {
	dbContextFactory func() (db.DbContext, error)
}

// Creates new account service
//	dbContextFactory - factory function used to create new db context
func NewAccountService(dbContextFactory func() (db.DbContext, error)) AccountService {
	return accountService{dbContextFactory}
}

func (svc accountService) ListAccounts() ([]Account, error) {
	dbContext, err := svc.dbContextFactory()
	if err != nil {
		return nil, err
	}
	defer dbContext.Release()

	var result = []Account{}
	err = dbContext.Query(
		"SELECT account_number, balance FROM public.accounts",
		sqlParams{},
		func(rows db.QueryResultRows) error {
			for rows.Next() {
				var (
					accountNumber int64
					balance       int64
				)
				err := rows.Scan(&accountNumber, &balance)
				if err != nil {
					return servErr.ErrDatabaseError(err)
				}

				var account = Account{
					Number:  AccountNumber(uint64(accountNumber)),
					Balance: balance,
				}

				result = append(result, account)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}
