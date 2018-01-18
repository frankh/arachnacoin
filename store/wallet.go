package store

import (
	"encoding/hex"
	"golang.org/x/crypto/ed25519"
)

type Wallet struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

var MyWallet *Wallet

func (w *Wallet) Address() string {
	return hex.EncodeToString(w.PublicKey)
}

func (w *Wallet) KeyStrings() (pub string, priv string) {
	return hex.EncodeToString(w.PublicKey), hex.EncodeToString(w.PrivateKey)
}

func FromKeyStrings(pub string, priv string) Wallet {
	pubKey, err := hex.DecodeString(pub)
	if err != nil {
		panic("Invalid public key")
	}
	privKey, err := hex.DecodeString(priv)
	if err != nil {
		panic("Invalid private key")
	}
	return Wallet{
		pubKey,
		privKey,
	}
}

func GenerateWallet() Wallet {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic("Couldn't generate private key")
	}
	return Wallet{
		pub,
		priv,
	}
}

func GetBalance(address string) uint32 {
	transactions := FetchTransactionsForAccount(address)
	balance := uint32(0)
	for _, t := range transactions {
		if t.Output == address {
			balance += t.Amount
		}
		if t.Input == address {
			balance -= t.Amount
		}
		if balance < 0 {
			panic("invalid transaction, caused negative balance")
		}
	}

	return balance
}
