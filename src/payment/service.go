package payment

import (
	"errors"

	"github.com/zombinome/coins-test/src/account"
)

type PaymentService interface {
	ListPayments(account account.AccountNumber) ([]Payment, error)

	SendPayment(id PaymentId, source, dest account.AccountNumber, amount uint64) error
}

type paymentService struct {
}

func NewPaymentService() PaymentService {
	return paymentService{}
}

func (svc paymentService) ListPayments(account account.AccountNumber) ([]Payment, error) {
	return nil, errors.New("not implemented")
}

func (svc paymentService) SendPayment(id PaymentId, source, dest account.AccountNumber, amount int64) error {
	return errors.New("not implemented")
}
