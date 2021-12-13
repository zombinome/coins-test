package account

type AccountNumber uint64

type Account struct {
	Number AccountNumber

	Amount int64

	LockedAmount int64
}
