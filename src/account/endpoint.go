package account

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

type listAccountsResponse struct {
	Accounts []Account `json:"accounts,omitempty"`
	Error    error     `json:"error,omitempty"`
}

func (r listAccountsResponse) error() error { return r.Error }

func makeListAccountsEndpoint(svc AccountService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		accounts, err := svc.ListAccounts()
		return listAccountsResponse{accounts, err}, nil
	}
}
