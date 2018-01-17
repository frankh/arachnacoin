package main

import (
	"github.com/frankh/arachnacoin/store"
	"github.com/frankh/arachnacoin/transaction"
	"github.com/frankh/arachnacoin/wallet"
	"github.com/frankh/arachnacoin/work"
	"log"
)

var memPool []*transaction.Transaction

func main() {
	log.Printf("Arachnacoin starting up...")
	store.Init("db.sqlite")
	head := store.FetchHighestBlock()
	log.Printf("Initialised... Longest chain is height %d", head.Height)

	for {
		log.Printf("%d transactions in mempool", len(memPool))
		newBlock := work.Mine(head, make([]transaction.Transaction, 0), "account")
		log.Printf("Mined new block, new height %d", newBlock.Height)
		store.StoreBlock(newBlock)
		log.Printf("Balance: %d", wallet.GetBalance("account"))
		head = newBlock
	}
}
