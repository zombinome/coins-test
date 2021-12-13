package payment

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
)

type listPaymentsRequest struct {
	Id uuid.UUID
}

type listPaymentsResponse struct {
	Payments []Payment `json:"payments,omitempty"`
	Error    error     `json:"error,omitempty"`
}

func (r listPaymentsResponse) error() error { return r.Error }

func makeListPaymentsEndpoint(svc PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listPaymentsRequest)
		payments, err := svc.ListPayments(req.Id)
		return listPaymentsResponse{payments, err}, nil
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

func makeSendPaymentEnpoint(svc PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(sendPaymentRequest)
		err := svc.SendPayment(PaymentId(req.Id), req.Source, req.Dest, req.Amount)
		return sendPaymentResponse{err}, nil
	}
}
