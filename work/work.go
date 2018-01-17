package work

import (
	"crypto/sha512"
	"encoding/binary"
	"github.com/frankh/arachnacoin/block"
	"github.com/frankh/arachnacoin/transaction"
)

var difficulty uint32 = 0xffffff00

func GenerateWork(b block.Block) uint32 {
	hash := b.Hash()
	work := uint32(0)
	for !ValidateWork(hash, work) {
		work++
	}

	return work
}

func ValidateWork(blockHash []byte, work uint32) bool {
	h := sha512.New()
	workBytes := make([]byte, 8)
	binary.BigEndian.PutUint32(workBytes, uint32(work))
	h.Write(workBytes)
	h.Write(blockHash)

	value := h.Sum(nil)
	valueInt := binary.BigEndian.Uint32(value)

	return valueInt > difficulty
}

func ValidateBlockWork(b block.Block) bool {
	return ValidateWork(b.Hash(), b.Work)
}

func Mine(previous block.Block, transactions []transaction.Transaction, rewardAccount string) block.Block {
	b := block.Block{
		previous.HashString(),
		0x0, //empty work to start with
		previous.Height + 1,
		transactions,
	}

	b.Work = GenerateWork(b)

	return b
}
