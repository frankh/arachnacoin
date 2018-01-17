package block

import (
	"crypto/sha512"
	"encoding/hex"
	"github.com/frankh/arachnacoin/transaction"
)

var GenesisBlock = Block{
	"00000000000000000000000000000000",
	0x000000000012f6de,
	make([]transaction.Transaction, 0),
}

type Block struct {
	Previous     string                    `json:"previous"`
	Work         uint64                    `json:"work"`
	Transactions []transaction.Transaction `json:"transactions"`
}

func (b *Block) Hash() []byte {
	h := sha512.New()
	prevBytes, _ := hex.DecodeString(b.Previous)

	h.Write(prevBytes)

	for _, t := range b.Transactions {
		h.Write(t.Hash())
	}

	return h.Sum(nil)
}
