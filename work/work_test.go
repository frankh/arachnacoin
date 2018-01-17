package work

import (
	"github.com/frankh/arachnacoin/block"
	"github.com/frankh/arachnacoin/transaction"
	"testing"
)

func TestGenerateWork(t *testing.T) {
	difficulty = 0xffff0000
	work := GenerateWork(block.GenesisBlock)
	if work < 1000 {
		t.Errorf("Work was too easy")
	}
}

func TestValidateWork(t *testing.T) {
	if !ValidateBlockWork(block.GenesisBlock) {
		t.Errorf("Invalid genesis block work")
	}
	block.GenesisBlock.Work--
	if ValidateBlockWork(block.GenesisBlock) {
		t.Errorf("Invalid work passed validation")
	}
}

func TestMine(t *testing.T) {
	difficulty = 0xffff0000
	b := Mine(block.GenesisBlock, make([]transaction.Transaction, 0), "unspendable")

	if !ValidateBlockWork(b) {
		t.Errorf("Work failed on mined block")
	}
}
