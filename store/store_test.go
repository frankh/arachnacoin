package store

import (
	"github.com/frankh/arachnacoin/block"
	"testing"
)

func TestStoreFetchGenesis(t *testing.T) {
	Init(":memory:")

	b := FetchBlock(block.GenesisBlock.HashString())
	if b.HashString() != block.GenesisBlock.HashString() {
		t.Errorf("Failed to store and fetch genesis block")
	}
}
