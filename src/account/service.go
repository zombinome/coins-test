package account

import "errors"

type AccountService interface {
	ListAccounts() ([]Account, error)
}

type accountService struct {
}

func NewAccountService() AccountService {
	return accountService{}
}

func (svc accountService) ListAccounts() ([]Account, error) {
	return nil, errors.New("not implemented")
}
