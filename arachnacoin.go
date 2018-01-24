package main

import (
	"github.com/frankh/arachnacoin/node"
	"github.com/frankh/arachnacoin/store"
	"github.com/frankh/arachnacoin/transaction"
	"github.com/frankh/arachnacoin/work"
	"log"
)

var memPool []*transaction.Transaction

func main() {
	log.Printf("Arachnacoin starting up...")
	go node.PeerServer()
	go node.ListenForPeers()
	go node.BroadcastForPeers()
	store.Init("db.sqlite")
	head := store.FetchHighestBlock()
	log.Printf("Initialised... Longest chain is height %d", head.Height)

	for {
		log.Printf("%d transactions in mempool", len(memPool))
		newBlock := work.Mine(head, make([]transaction.Transaction, 0), store.MyWallet.Address())
		log.Printf("Mined new block, new height %d", newBlock.Height)
		store.StoreBlock(newBlock)
		node.BroadcastLatestBlock()
		head = store.FetchHighestBlock()
		if head.Height > newBlock.Height {
			log.Printf("Block was orphaned :(")
		}
		log.Printf("Balance: %d", store.GetBalance(store.MyWallet.Address()))
	}
}
