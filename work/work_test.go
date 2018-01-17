package work

import (
	"github.com/frankh/arachnacoin/block"
	"testing"
)

func TestGenerateWork(t *testing.T) {
	GenerateWork(block.GenesisBlock)
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
