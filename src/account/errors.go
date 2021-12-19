package account

import (
	"fmt"
	servErr "test/coins/errors"
)

const (
	errCodeInvalidAccount int = 20 + iota
)

// Creates new "Invalid account number" error
// 	accountNum - account number
// Returns created error
func ErrInvalidAccount(accountNum AccountNumber) error {
	var msg = fmt.Sprintf("account with number [%d] not found", uint64(accountNum))
	return servErr.NewServiceError(msg, nil, errCodeInvalidAccount)
}
