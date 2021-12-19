package transfer

import (
	"context"
	"test/coins/account"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
)

type listTransfersRequest struct {
	AccountNumber uint64
}

type listTransfersResponse struct {
	Tranfers []Transfer `json:"transfers,omitempty"`
	Error    error      `json:"error,omitempty"`
}

func (r listTransfersResponse) error() error { return r.Error }

func makeListTransfersEndpoint(svc TransferService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listTransfersRequest)
		accountNumber := account.AccountNumber(req.AccountNumber)
		transfers, err := svc.ListTransfers(accountNumber)
		return listTransfersResponse{transfers, err}, nil
	}
}

type sendPaymentRequest struct {
	Id     uuid.UUID `json:"id"`
	Source uint64    `json:"source"`
	Dest   uint64    `json:"dest"`
	Amount uint64    `json:"amount"`
}

type sendPaymentResponse struct {
	Error error `json:"error,omitempty"`
}

func (r sendPaymentResponse) error() error { return r.Error }

func makeSendPaymentEnpoint(svc TransferService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(sendPaymentRequest)
		sourceAcc := account.AccountNumber(req.Source)
		destAcc := account.AccountNumber(req.Dest)
		err := svc.TransferMoney(TransferId(req.Id), sourceAcc, destAcc, req.Amount)
		return sendPaymentResponse{err}, nil
	}
}
