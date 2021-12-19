package transfer_test

import (
	"errors"
	"fmt"
	"test/coins/account"
	"test/coins/db"
	"test/coins/transfer"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	servErr "test/coins/errors"
)

const (
	dbAccountNumber1 int64 = 1
	dbAccountNumber2 int64 = 2
)

func setupService(setupMock func(mock sqlmock.Sqlmock)) transfer.TransferService {
	return transfer.NewTransferService(func() (db.DbContext, error) {
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

func Test_ListTransfers_SqlErrorHandled(t *testing.T) {
	// Arrange
	var expectedErr = errors.New("database related error")
	var dbMock sqlmock.Sqlmock = nil
	var service = setupService(func(mock sqlmock.Sqlmock) {
		dbMock = mock
		mock.ExpectBegin()

		var accCountRows = sqlmock.NewRows([]string{""}).AddRow(1)
		mock.ExpectQuery("SELECT COUNT").WillReturnRows(accCountRows)

		mock.ExpectQuery("SELECT transfer_id, amount, source_account, dest_account, created_at FROM public.transfers").WillReturnError(expectedErr)

		mock.ExpectRollback()
	})

	// Act
	transfers, err := service.ListTransfers(1)

	// Assert
	isValid, msg := valdiateServiceError(servErr.ErrorKindDB, expectedErr, err, "ListTransfers()")
	if !isValid {
		t.Fatalf(msg)
	}

	if transfers != nil {
		t.Fatalf("in case of any error, ListTransfers() should return (nil, error) as result")
	}

	err = dbMock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("db methods call expectations were not met: %s", err.Error())
	}
}

func Test_ListTransfer_InvalidAccount(t *testing.T) {
	// Arrange
	var dbMock sqlmock.Sqlmock = nil
	var service = setupService(func(mock sqlmock.Sqlmock) {
		dbMock = mock
		mock.ExpectBegin()

		var accCountRows = sqlmock.NewRows([]string{""}).AddRow(0)
		mock.ExpectQuery("SELECT COUNT").WillReturnRows(accCountRows)

		mock.ExpectRollback()
	})

	// Act
	transfers, err := service.ListTransfers(1)

	// Assert
	if err == nil {
		t.Fatalf("expected error t obe returned by ListTransfers method")
	}

	if transfers != nil {
		t.Fatalf("in case of any error, ListTransfers() should return (nil, error) as result")
	}

	serviceErr, ok := err.(servErr.ServiceError)
	if !ok {
		t.Fatalf("error of type ServiceError expected")
	}

	if serviceErr.Kind() != transfer.ErrKindInvalidAccount {
		t.Fatalf("invalid error kind returned: %d, expected: %d", serviceErr.Kind(), transfer.ErrKindInvalidAccount)
	}

	err = dbMock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("db methods call expectations were not met: %s", err.Error())
	}
}

func Test_ListTransfers_NoTransfers(t *testing.T) {
	// Arrange
	var dbMock sqlmock.Sqlmock = nil
	var service = setupService(func(mock sqlmock.Sqlmock) {
		dbMock = mock
		mock.ExpectBegin()

		var accCountRows = sqlmock.NewRows([]string{""}).AddRow(1)
		mock.ExpectQuery("SELECT COUNT").WillReturnRows(accCountRows)

		var rows = sqlmock.NewRows([]string{"transfer_id", "amount", "source_account", "dest_account", "created_at"})
		mock.ExpectQuery("SELECT transfer_id, amount, source_account, dest_account, created_at FROM public.transfers").WillReturnRows(rows)

		mock.ExpectRollback()
	})

	// Act
	transfers, err := service.ListTransfers(1)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error occured when ListTransfers() was called: %s", err.Error())
	}

	if transfers == nil || len(transfers) != 0 {
		t.Fatalf("expected empty list of money transfers")
	}

	err = dbMock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("db methods call expectations were not met: %s", err.Error())
	}
}

func Test_TransferMoney_SqlErrorHandled(t *testing.T) {
	// Arrange
	var (
		transferId        = transfer.TransferId(uuid.New())
		sourceAcc         = account.AccountNumber(1)
		descAcc           = account.AccountNumber(2)
		amount     uint64 = 250
	)

	var expectedErr = errors.New("database related error")
	var dbMock sqlmock.Sqlmock = nil
	var service = setupService(func(mock sqlmock.Sqlmock) {
		dbMock = mock
		mock.ExpectBegin()

		mock.ExpectQuery("SELECT account_number, balance FROM public.accounts").WillReturnError(expectedErr)

		mock.ExpectRollback()
	})

	// Act
	err := service.TransferMoney(transferId, sourceAcc, descAcc, amount)

	// Assert
	isValid, msg := valdiateServiceError(servErr.ErrorKindDB, expectedErr, err, "SendMoney()")
	if !isValid {
		t.Fatalf(msg)
	}

	err = dbMock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("db methods call expectations were not met: %s", err.Error())
	}
}

func Test_TransferMoney_CheckForValidSourceAccount(t *testing.T) {
	// Arrange
	var (
		transferId        = transfer.TransferId(uuid.New())
		sourceAcc         = account.AccountNumber(dbAccountNumber1)
		destAcc           = account.AccountNumber(dbAccountNumber2)
		amount     uint64 = 250
	)

	var dbMock sqlmock.Sqlmock = nil
	var service = setupService(func(mock sqlmock.Sqlmock) {
		dbMock = mock
		mock.ExpectBegin()

		var accountsListRows = sqlmock.NewRows([]string{"account_number", "balance"})
		mock.ExpectQuery("SELECT account_number, balance FROM public.accounts").WillReturnRows(accountsListRows)

		mock.ExpectRollback()
	})

	// Act
	var err = service.TransferMoney(transferId, sourceAcc, destAcc, amount)

	// Assert
	isValid, msg := valdiateServiceError(transfer.ErrKindInvalidAccount, nil, err, "TransferMoney(...)")
	if !isValid {
		t.Fatalf(msg)
	}

	var expectedErrorMsg = transfer.ErrInvalidAccount(sourceAcc).Error()
	if err.Error() != expectedErrorMsg {
		t.Fatalf("error message contains invalid wrong account number. Expected message: %s, acutal: %s", err.Error(), expectedErrorMsg)
	}

	err = dbMock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("db methods call expectations were not met: %s", err.Error())
	}
}

func Test_TransferMoney_CheckForValidDestAccount(t *testing.T) {
	// Arrange
	var (
		transferId        = transfer.TransferId(uuid.New())
		sourceAcc         = account.AccountNumber(dbAccountNumber1)
		destAcc           = account.AccountNumber(dbAccountNumber2)
		amount     uint64 = 250
	)

	var dbMock sqlmock.Sqlmock = nil
	var service = setupService(func(mock sqlmock.Sqlmock) {
		dbMock = mock
		mock.ExpectBegin()

		var accountsListRows = sqlmock.
			NewRows([]string{"account_number", "balance"}).
			AddRow(dbAccountNumber1, 1000)
		mock.ExpectQuery("SELECT account_number, balance FROM public.accounts").WillReturnRows(accountsListRows)

		mock.ExpectRollback()
	})

	// Act
	var err = service.TransferMoney(transferId, sourceAcc, destAcc, amount)

	// Assert
	isValid, msg := valdiateServiceError(transfer.ErrKindInvalidAccount, nil, err, "TransferMoney(...)")
	if !isValid {
		t.Fatalf(msg)
	}

	var expectedErrorMsg = transfer.ErrInvalidAccount(destAcc).Error()
	if err.Error() != expectedErrorMsg {
		t.Fatalf("error message contains invalid wrong account number. Expected message: %s, acutal: %s", err.Error(), expectedErrorMsg)
	}

	err = dbMock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("db methods call expectations were not met: %s", err.Error())
	}
}

func Test_TransferMoney_DoesNotAllowToCreateTwoTransfersWithSameId(t *testing.T) {
	// Arrange
	var (
		transferId        = transfer.TransferId(uuid.New())
		sourceAcc         = account.AccountNumber(dbAccountNumber1)
		descAcc           = account.AccountNumber(dbAccountNumber2)
		amount     uint64 = 250
	)

	var dbMock sqlmock.Sqlmock = nil
	var service = setupService(func(mock sqlmock.Sqlmock) {
		dbMock = mock
		mock.ExpectBegin()

		var accountsListRows = sqlmock.
			NewRows([]string{"account_number", "balance"}).
			AddRow(dbAccountNumber1, 1000).
			AddRow(dbAccountNumber2, 2000)
		mock.ExpectQuery("SELECT account_number, balance FROM public.accounts").WillReturnRows(accountsListRows)

		var duplicateCheckRows = sqlmock.NewRows([]string{""}).AddRow(2)
		mock.ExpectQuery("SELECT COUNT").WillReturnRows(duplicateCheckRows)

		mock.ExpectRollback()
	})

	// Act
	var err = service.TransferMoney(transferId, sourceAcc, descAcc, amount)

	// Assert
	isValid, msg := valdiateServiceError(transfer.ErrKindTransferAlreadyComplete, nil, err, "TransferMoney(...)")
	if !isValid {
		t.Fatalf(msg)
	}

	err = dbMock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("db methods call expectations were not met: %s", err.Error())
	}
}

func Test_TransferMoney_CheckForEnoughMoney(t *testing.T) {
	// Arrange
	var (
		transferId        = transfer.TransferId(uuid.New())
		sourceAcc         = account.AccountNumber(dbAccountNumber1)
		descAcc           = account.AccountNumber(dbAccountNumber2)
		amount     uint64 = 1250
	)

	var dbMock sqlmock.Sqlmock = nil
	var service = setupService(func(mock sqlmock.Sqlmock) {
		dbMock = mock
		mock.ExpectBegin()

		var accountsListRows = sqlmock.
			NewRows([]string{"account_number", "balance"}).
			AddRow(dbAccountNumber1, 1000).
			AddRow(dbAccountNumber2, 2000)
		mock.ExpectQuery("SELECT account_number, balance FROM public.accounts").WillReturnRows(accountsListRows)

		var duplicateCheckRows = sqlmock.NewRows([]string{""}).AddRow(0)
		mock.ExpectQuery("SELECT COUNT").WillReturnRows(duplicateCheckRows)

		mock.ExpectRollback()
	})

	// Act
	var err = service.TransferMoney(transferId, sourceAcc, descAcc, amount)

	// Assert

	isValid, msg := valdiateServiceError(transfer.ErrKindNotEnoughMoney, nil, err, "TransferMoney(...)")
	if !isValid {
		t.Fatalf(msg)
	}

	err = dbMock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("db methods call expectations were not met: %s", err.Error())
	}
}

func Test_TransferMoney_SuccessOnValidDataProvided(t *testing.T) {
	// Arrange
	var (
		transferUuid        = uuid.New()
		transferId          = transfer.TransferId(transferUuid)
		sourceAcc           = account.AccountNumber(dbAccountNumber1)
		descAcc             = account.AccountNumber(dbAccountNumber2)
		amount       uint64 = 250
	)

	var dbMock sqlmock.Sqlmock = nil
	var service = setupService(func(mock sqlmock.Sqlmock) {
		dbMock = mock
		mock.ExpectBegin()

		var accountsListRows = sqlmock.
			NewRows([]string{"account_number", "balance"}).
			AddRow(dbAccountNumber1, 1000).
			AddRow(dbAccountNumber2, 2000)
		mock.ExpectQuery("SELECT account_number, balance FROM public.accounts").WillReturnRows(accountsListRows)

		var duplicateCheckRows = sqlmock.NewRows([]string{""}).AddRow(0)
		mock.ExpectQuery("SELECT COUNT").WillReturnRows(duplicateCheckRows)

		var updateCountResult = sqlmock.NewResult(0, 1)
		mock.ExpectExec(
			"UPDATE public.accounts SET balance = balance",
		).WithArgs(int64(amount), dbAccountNumber1).WillReturnResult(updateCountResult)

		mock.ExpectExec(
			"UPDATE public.accounts SET balance = balance",
		).WithArgs(int64(amount), dbAccountNumber2).WillReturnResult(updateCountResult)

		var insertCountResult = sqlmock.NewResult(0, 1)
		mock.ExpectExec(
			"INSERT INTO public.transfers",
		).WithArgs(transferUuid, int64(amount), dbAccountNumber1, dbAccountNumber2).WillReturnResult(insertCountResult)

		mock.ExpectCommit()
	})

	// Act
	var err = service.TransferMoney(transferId, sourceAcc, descAcc, amount)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error returned when called for TransferMoney(...): %s", err.Error())
	}

	err = dbMock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("db methods call expectations were not met: %s", err.Error())
	}
}
