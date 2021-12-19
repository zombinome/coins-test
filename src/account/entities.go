package account

type AccountNumber uint64

type Account struct {
	// Account number
	Number AccountNumber `json:"number"`

	// Current account balance
	Balance int64 `json:"balance"`
}
