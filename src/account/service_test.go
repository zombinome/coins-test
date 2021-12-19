package account_test

import (
	"errors"
	"fmt"
	"test/coins/account"
	"test/coins/db"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	servErr "test/coins/errors"
)

func setupService(setupMock func(mock sqlmock.Sqlmock)) account.AccountService {
	return account.NewAccountService(func() (db.DbContext, error) {
		return db.CreateMockDbContext(setupMock)
	})
}

func valdiateServiceError(expectedKind int, expectedInnerErr error, actual error, method string) (bool, string) {
	if actual == nil {
		return false, fmt.Sprintf("error expected to be returned by method %s", method)
	}

	err, ok := actual.(servErr.ServiceError)
	if !ok {
		return false, "expected error to be of type ServiceError"
	}

	if err.Kind() != expectedKind {
		return false, fmt.Sprintf("expected error with kind %d, got %d", expectedKind, err.Kind())
	}

	if err.Unwrap() != expectedInnerErr {
		return false, "inner error differs from expected"
	}

	return true, ""
}

func TestP_ListAccounts_SqlErrorHandled(t *testing.T) {
	// Arrange
	var expectedErr = errors.New("database related error")
	var dbMock sqlmock.Sqlmock = nil
	var service = setupService(func(mock sqlmock.Sqlmock) {
		dbMock = mock
		mock.ExpectBegin()

		mock.ExpectQuery("SELECT account_number, balance FROM public.accounts").WillReturnError(expectedErr)

		mock.ExpectRollback()
	})

	// Act
	accounts, err := service.ListAccounts()

	// Assert
	isValid, msg := valdiateServiceError(servErr.ErrorKindDB, expectedErr, err, "ListAccounts()")
	if !isValid {
		t.Fatalf(msg)
	}

	if accounts != nil {
		t.Fatalf("in case of any error, ListAccount() should return (nil, error) as result")
	}

	err = dbMock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("db methods call expectations were not met: %s", err.Error())
	}
}

func Test_ListAccounts_NoAccounts(t *testing.T) {
	// Arrange
	var dbMock sqlmock.Sqlmock = nil
	var service = setupService(func(mock sqlmock.Sqlmock) {
		dbMock = mock
		mock.ExpectBegin()

		var rows = sqlmock.NewRows([]string{"account_number", "balance"})
		mock.ExpectQuery("SELECT account_number, balance FROM public.accounts").WillReturnRows(rows)

		mock.ExpectRollback()
	})

	// Act
	accounts, err := service.ListAccounts()

	// Assert
	if err != nil {
		t.Fatalf("unexpected error occured when ListAccounts() was called: %s", err.Error())
	}

	if accounts == nil || len(accounts) != 0 {
		t.Fatalf("expected empty list of accounts")
	}

	err = dbMock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("db methods call expectations were not met: %s", err.Error())
	}
}

func Test_ListAccounts_AccountsRetrievedSuccessfully(t *testing.T) {
	// Arrange
	var dbMock sqlmock.Sqlmock = nil

	var (
		an1 int64 = 1
		b1  int64 = 1000
		an2 int64 = 2
		b2  int64 = 2000
	)
	var service = setupService(func(mock sqlmock.Sqlmock) {
		dbMock = mock
		mock.ExpectBegin()

		var rows = sqlmock.
			NewRows([]string{"account_number", "balance"}).
			AddRow(an1, b1).
			AddRow(an2, b2)
		mock.ExpectQuery("SELECT account_number, balance FROM public.accounts").WillReturnRows(rows)

		mock.ExpectRollback()
	})

	// Act
	accounts, err := service.ListAccounts()

	// Assert
	if err != nil {
		t.Fatalf("unexpected error occured when ListAccounts() was called: %s", err.Error())
	}

	if accounts == nil || len(accounts) != 2 {
		t.Fatalf("expected list of 2 accounts")
	}

	err = dbMock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("db methods call expectations were not met: %s", err.Error())
	}

	var (
		acc1 account.Account
		acc2 account.Account
	)
	for _, acc := range accounts {
		if acc.Number == account.AccountNumber(uint64(an1)) {
			acc1 = acc
		}
		if acc.Number == account.AccountNumber(uint64(an2)) {
			acc2 = acc
		}
	}

	if acc1.Number != account.AccountNumber(uint64(an1)) {
		t.Fatalf("account 1 was not read from databse")
	}

	if acc1.Balance != b1 {
		t.Fatalf("invalid account 1 balance")
	}

	if acc2.Number != account.AccountNumber(uint64(an2)) {
		t.Fatalf("account 2 was not read from database")
	}

	if acc2.Balance != b2 {
		t.Fatalf("invalid account 2 balance")
	}
}
