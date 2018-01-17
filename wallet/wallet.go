package wallet

import (
	"github.com/frankh/arachnacoin/store"
)

func GetBalance(address string) uint32 {
	transactions := store.FetchTransactionsForAccount(address)
	balance := uint32(0)
	for _, t := range transactions {
		if t.Output == address {
			balance += t.Amount
		}
		if t.Input == address {
			balance -= t.Amount
		}
		if balance < 0 {
			panic("invalid transaction, caused negative balance")
		}
	}

	return balance
}
