package transfer

import (
	"errors"
	"test/coins/account"
	"test/coins/db"
	"time"

	"github.com/google/uuid"

	servErr "test/coins/errors"
)

var errQueryReturnedNoData = errors.New("no data returned from database request")

// Type alias for sql parameters array
type sqlParams = []interface{}

// Transfer service. Incapsulates operations with money transfers
type TransferService interface {
	// Returns list of transfers for specific account
	//	accountNum - account number
	// Returns list of transfers for specified account
	ListTransfers(accountNum account.AccountNumber) ([]Transfer, error)

	// Transfers money between accounts
	//	id - unique transfer id
	// 	source - source account number
	// 	dest   - dest account number
	//	amount - amount to trangfer
	TransferMoney(id TransferId, source, dest account.AccountNumber, amount uint64) error
}

// Transfer service implementation
type transferService struct {
	dbContextFactory func() (db.DbContext, error)
}

// Creates new transfer service
//	dbContextFactory - factory function used to create new db context
func NewTransferService(dbContextFactory func() (db.DbContext, error)) TransferService {
	return transferService{dbContextFactory}
}

func (svc transferService) ListTransfers(accountNumber account.AccountNumber) ([]Transfer, error) {
	dbContext, err := svc.dbContextFactory()
	if err != nil {
		return nil, err
	}

	defer dbContext.Release()

	var accountNum = int64(uint64(accountNumber))
	var count int64 = -1
	err = dbContext.Query(
		"SELECT COUNT(*) FROM public.accounts WHERE account_number = $1",
		sqlParams{accountNum},
		func(rows db.QueryResultRows) error {
			if !rows.Next() {
				return servErr.ErrDatabaseError(errQueryReturnedNoData)
			}

			rows.Scan(&count)
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, ErrInvalidAccount(accountNumber)
	}

	var result = []Transfer{}
	err = dbContext.Query(
		"SELECT transfer_id, amount, source_account, dest_account, created_at FROM public.transfers "+
			"WHERE source_account = $1 or dest_account = $1 ORDER BY created_at DESC",
		sqlParams{accountNum},
		func(rows db.QueryResultRows) error {
			for rows.Next() {
				var (
					id        uuid.UUID
					amount    int64
					createdAt time.Time
					sourceAcc int64
					destAcc   int64
				)

				err = rows.Scan(&id, &amount, &sourceAcc, &destAcc, &createdAt)
				if err != nil {
					return servErr.ErrDatabaseError(err)
				}

				var direction string
				var fromAccount *account.AccountNumber = nil
				var toAccount *account.AccountNumber = nil

				if account.AccountNumber(sourceAcc) == accountNumber {
					var val = account.AccountNumber(destAcc)
					toAccount = &val
					direction = DirectionOutgoing
				}
				if account.AccountNumber(destAcc) == accountNumber {
					var val = account.AccountNumber(sourceAcc)
					fromAccount = &val
					direction = DirectionIncoming
				}

				result = append(result, Transfer{
					Id:          TransferId(id),
					Account:     accountNumber,
					Amount:      amount,
					CreatedAt:   createdAt,
					Direction:   direction,
					FromAccount: fromAccount,
					ToAccount:   toAccount,
				})
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (svc transferService) TransferMoney(id TransferId, source, dest account.AccountNumber, amount uint64) error {
	dbContext, err := svc.dbContextFactory()
	if err != nil {
		return err
	}

	defer dbContext.Release()

	// Reading existing accounts
	// Rows for accounts would be blocked until transaction is finished
	sourceAccount, destAccount, err := readPaymentAccounts(dbContext, source, dest)
	if err != nil {
		return err
	}

	if sourceAccount == nil {
		return ErrInvalidAccount(source)
	}

	if destAccount == nil {
		return ErrInvalidAccount(dest)
	}

	// Checking if money thransfer with the same ID already exists (to avoid revolut-like fuckup)
	isDuplicate, err := checkForTransferDuplicate(dbContext, id)
	if err != nil {
		return err
	}

	if isDuplicate {
		return ErrTransferAlreadyComplete
	}

	// checking for balance
	if sourceAccount.Balance < int64(amount) {
		return ErrNotEnoughMoney
	}

	// updating balance
	err = updateAccountBalancesForTransfer(dbContext, source, dest, amount)
	if err != nil {
		return err
	}

	// adding payment history records for both accounts
	err = addPaymentHistory(dbContext, uuid.UUID(id), source, dest, amount)
	if err != nil {
		return err
	}

	return dbContext.Save()
}

func readPaymentAccounts(dbContext db.DbContext, sourceNumber, destNumber account.AccountNumber) (sourceAccount, destAccount *account.Account, err error) {
	sourceAccount = nil
	destAccount = nil

	err = dbContext.Query(
		"SELECT account_number, balance FROM public.accounts WHERE account_number = $1 or account_number = $2 FOR UPDATE",
		sqlParams{sourceNumber, destNumber},
		func(rows db.QueryResultRows) error {
			for rows.Next() {
				var (
					accNum  int64
					balance int64
				)

				err := rows.Scan(&accNum, &balance)
				if err != nil {
					return servErr.ErrDatabaseError(err)
				}

				var acc = account.Account{Number: account.AccountNumber(uint64(accNum)), Balance: balance}
				if acc.Number == sourceNumber {
					sourceAccount = &acc
				}

				if acc.Number == destNumber {
					destAccount = &acc
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return sourceAccount, destAccount, nil
}

func checkForTransferDuplicate(dbContext db.DbContext, transferId TransferId) (bool, error) {
	var id = uuid.UUID(transferId)
	var result = false
	var err = dbContext.Query(
		"SELECT COUNT(*) FROM public.transfers WHERE transfer_id = $1",
		sqlParams{id},
		func(value db.QueryResultRows) error {
			if !value.Next() {
				return servErr.ErrDatabaseError(errQueryReturnedNoData)
			}

			var val int64
			err := value.Scan(&val)
			if err != nil {
				return servErr.ErrDatabaseError(err)
			}

			result = val > 0
			return nil
		})

	return result, err
}

func updateAccountBalancesForTransfer(dbContext db.DbContext, sourceNumber, destNumber account.AccountNumber, amount uint64) error {
	rowsAffected, err := dbContext.Execute(
		"UPDATE public.accounts SET balance = balance - $1 WHERE account_number = $2",
		int64(amount), int64(uint64(sourceNumber)),
	)

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrInvalidAccount(sourceNumber)
	}

	rowsAffected, err = dbContext.Execute(
		"UPDATE public.accounts SET balance = balance + $1 WHERE account_number = $2",
		int64(amount), destNumber,
	)

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrInvalidAccount(destNumber)
	}

	return nil
}

func addPaymentHistory(dbContext db.DbContext, transferId uuid.UUID, sourceNumber, destNumber account.AccountNumber, amount uint64) error {
	rowsAffected, err := dbContext.Execute(
		"INSERT INTO public.transfers (transfer_id, amount, source_account, dest_account)"+
			"VALUES ($1, $2, $3, $4)",
		transferId, amount, sourceNumber, destNumber,
	)
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return ErrTransferAlreadyComplete
	}

	return nil
}
