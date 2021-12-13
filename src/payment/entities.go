package payment

import (
	"github.com/google/uuid"
	"github.com/zombinome/coins-test/src/account"
)

type PaymentId uuid.UUID

type Payment struct {
	Id PaymentId

	Source account.AccountNumber

	Dest account.AccountNumber

	Amount int64

	Status PaymentStatus
}

type PaymentStatus byte

const (
	PaymentNew        PaymentStatus = 0
	PaymentInProgress PaymentStatus = 1
	PaymentComplete   PaymentStatus = 2
	PaymentRejected   PaymentStatus = 3
)

type RejectReason byte

const (
	RejectReasonOther                RejectReason = 0
	RejectReasonNoMoney              RejectReason = 1
	RejectReasonInvalidSOurceAccount RejectReason = 2
	RejectReasonInvalidTargetAccount RejectReason = 3
)
