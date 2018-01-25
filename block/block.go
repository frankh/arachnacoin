package block

import (
	"crypto/sha512"
	"encoding/hex"
	"github.com/frankh/arachnacoin/transaction"
)

const BlockReward = uint32(5000)

var GenesisBlock = Block{
	"00000000000000000000000000000000",
	0x01f17e51,
	0,
	make([]transaction.Transaction, 0),
}

type Block struct {
	Previous     string                    `json:"previous"`
	Work         uint32                    `json:"work"`
	Height       uint32                    `json:"height"`
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

func (b *Block) HashString() string {
	return hex.EncodeToString(b.Hash())
}
