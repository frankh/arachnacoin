package transaction

import (
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
)

type Transaction struct {
	Input     string `json:"input"`
	Output    string `json:"output"`
	Amount    uint32 `json:"amount"`
	Signature string `json:"signature"`
	Unique    string `json:"unique"`
}

func (t *Transaction) Hash() []byte {
	h := sha512.New()
	inputBytes, _ := hex.DecodeString(string(t.Input))
	outputBytes, _ := hex.DecodeString(string(t.Output))
	uniqueBytes, _ := hex.DecodeString(string(t.Unique))
	amountBytes := make([]byte, 8)
	binary.BigEndian.PutUint32(amountBytes, t.Amount)

	h.Write(inputBytes)
	h.Write(outputBytes)
	h.Write(amountBytes)
	h.Write(uniqueBytes)

	return h.Sum(nil)
}

func (t *Transaction) HashString() string {
	return hex.EncodeToString(t.Hash())
}
