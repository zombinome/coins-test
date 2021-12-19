package transfer

import (
	"test/coins/account"
	"time"

	"github.com/google/uuid"
)

type TransferId uuid.UUID

// Needed to support proper serialization to JSON
func (id TransferId) MarshalJSON() ([]byte, error) {
	var guid = uuid.UUID(id)
	var str = guid.String()

	return []byte("\"" + str + "\""), nil
}

// Needed to support proper deserialization from JSON
func (id *TransferId) UnmarshalJSON(data []byte) error {
	guid, err := uuid.ParseBytes(data)
	if err != nil {
		return err
	}

	*id = TransferId(guid)
	return nil
}

const DirectionIncoming = "incoming"
const DirectionOutgoing = "outgoing"

type Transfer struct {
	// Transfer id
	Id TransferId `json:"id"`

	// Current account number
	Account account.AccountNumber `json:"account"`

	// Account from where money was trasnferred (if direction is "incoming", nil otherwise)
	FromAccount *account.AccountNumber `json:"fromAccount,omitempty"`

	// Account to where money was trasferred (if direction is "outgoing", nil therwise)
	ToAccount *account.AccountNumber `json:"toAccount,omitempty"`

	// Transfer amount
	Amount int64 `json:"amount"`

	// Account direction. Can have values "outgoing" or "incoming"
	Direction string `json:"direction"`

	// Account created timestamp
	CreatedAt time.Time `json:"createdAt"`
}
