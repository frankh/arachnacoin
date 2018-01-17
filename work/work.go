package work

import (
	"crypto/sha512"
	"encoding/binary"
	"github.com/frankh/arachnacoin/block"
)

var difficulty uint64 = 0xfffff00000000000

func GenerateWork(b block.Block) uint64 {
	hash := b.Hash()
	work := uint64(0)
	for !ValidateWork(hash, work) {
		work++
	}

	return work
}

func ValidateWork(blockHash []byte, work uint64) bool {
	h := sha512.New()
	workBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(workBytes, work)
	h.Write(workBytes)
	h.Write(blockHash)

	value := h.Sum(nil)
	valueInt := binary.BigEndian.Uint64(value)

	return valueInt > difficulty
}

func ValidateBlockWork(b block.Block) bool {
	return ValidateWork(b.Hash(), b.Work)
}
