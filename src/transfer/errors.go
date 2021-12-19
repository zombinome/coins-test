package transfer

import (
	"fmt"
	"test/coins/account"
	servErr "test/coins/errors"
)

const (
	ErrKindInvalidAccount int = 10 + iota
	ErrKindNotEnoughMoney
	ErrKindTransferAlreadyComplete
)

// Creates new "Invalid account number" error
// 	accountNum - account number
// Returns created error
func ErrInvalidAccount(accountNum account.AccountNumber) error {
	var msg = fmt.Sprintf("account with number [%d] not found", uint64(accountNum))
	return servErr.NewServiceError(msg, nil, ErrKindInvalidAccount)
}

// Error that is expected when there is not enough money on source account to complete transfer
var ErrNotEnoughMoney = servErr.NewServiceError(
	"source account does not have enough money", nil, ErrKindNotEnoughMoney)

// Error that is expected when money transfer with provided id is already complete
var ErrTransferAlreadyComplete = servErr.NewServiceError(
	"transfer already complete", nil, ErrKindTransferAlreadyComplete)
