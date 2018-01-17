package wallet

import (
	"github.com/frankh/arachnacoin/block"
	"github.com/frankh/arachnacoin/store"
	"github.com/frankh/arachnacoin/transaction"
	"github.com/frankh/arachnacoin/work"
	"testing"
)

func TestGetBalance(t *testing.T) {
	work.Difficulty = 0xff000000
	store.Init(":memory:")

	if GetBalance("zero") != 0 {
		t.Errorf("Empty account should have zero balance")
	}

	b := work.Mine(store.FetchHighestBlock(), make([]transaction.Transaction, 0), "rewardAccount")
	store.StoreBlock(b)
	if GetBalance("rewardAccount") != block.BlockReward {
		t.Errorf("Blockreward not added")
	}
}
