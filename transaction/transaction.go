package transaction

import (
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
)

type Transaction struct {
	Input     string `json:"input"`
	Output    string `json:"output"`
	Amount    uint64 `json:"amount"`
	Signature string `json:"signature"`
}

func (t *Transaction) Hash() []byte {
	h := sha512.New()
	inputBytes, _ := hex.DecodeString(string(t.Input))
	outputBytes, _ := hex.DecodeString(string(t.Output))
	amountBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(amountBytes, t.Amount)

	h.Write(inputBytes)
	h.Write(outputBytes)
	h.Write(amountBytes)

	return h.Sum(nil)
}
