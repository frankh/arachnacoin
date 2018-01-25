package store

import (
	"database/sql"
	"github.com/frankh/arachnacoin/block"
	"github.com/frankh/arachnacoin/transaction"
	"github.com/frankh/arachnacoin/work"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strings"
)

var Conn *sql.DB

func Init(path string) {
	var err error
	Conn, err = sql.Open("sqlite3", path)
	if err != nil {
		panic(err)
	}
	Conn.SetMaxOpenConns(1)

	table_check, err := Conn.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name='arach_block'`)
	if err != nil {
		panic(err)
	}

	if !table_check.Next() {
		log.Printf("Database was empty, creating schema...")
		prep, err := Conn.Prepare(`
      CREATE TABLE 'arach_block' (
        'hash' TEXT PRIMARY KEY,
        'height' INT NOT NULL,
        'previous' TEXT NOT NULL,
        'work' INT NOT NULL,
        'created' DATE DEFAULT CURRENT_TIMESTAMP NOT NULL,
        FOREIGN KEY(previous) REFERENCES block(hash)
      );
    `)
		if err != nil {
			panic(err)
		}
		_, err = prep.Exec()
		if err != nil {
			panic(err)
		}
		prep, err = Conn.Prepare(`
      CREATE TABLE 'arach_transaction' (
        'hash' TEXT PRIMARY KEY,
        'input' TEXT NOT NULL,
        'output' TEXT NOT NULL,
        'amount' INT NOT NULL,
        'signature' TEXT NOT NULL,
        'unique_string' TEXT NOT NULL,
        'order' INT NOT NULL,
        'block' TEXT NOT NULL,
        'block_height' INT NOT NULL,
        'created' DATE DEFAULT CURRENT_TIMESTAMP NOT NULL
      );
      FOREIGN KEY(block) REFERENCES block(hash)
    `)
		if err != nil {
			panic(err)
		}
		_, err = prep.Exec()
		if err != nil {
			panic(err)
		}
		prep, err = Conn.Prepare(`
      CREATE TABLE 'arach_wallet' (
        'public_key' TEXT PRIMARY KEY,
        'private_key' TEXT NOT NULL,
        'created' DATE DEFAULT CURRENT_TIMESTAMP NOT NULL
      );
    `)
		if err != nil {
			panic(err)
		}
		_, err = prep.Exec()
		if err != nil {
			panic(err)
		}
	}
	table_check.Close()
	rows, err := Conn.Query(`SELECT hash FROM arach_block WHERE hash=?`, block.GenesisBlock.HashString())
	if err != nil {
		panic(err)
	}
	if !rows.Next() {
		log.Printf("Creating genesis block...")
		StoreBlock(block.GenesisBlock)
	}
	rows.Close()

	w := FetchWallet()
	MyWallet = &w
}

func StoreWallet(w Wallet) {
	prep, err := Conn.Prepare(`
    INSERT INTO arach_wallet (
      public_key,
      private_key
    ) values (
      ?,?
    )
  `)

	if err != nil {
		panic(err)
	}

	pub, priv := w.KeyStrings()
	_, err = prep.Exec(
		pub,
		priv,
	)

	if err != nil {
		panic(err)
	}

}

func FetchWallet() Wallet {
	if Conn == nil {
		panic("Database connection not initialised")
	}

	rows, err := Conn.Query(`SELECT
    public_key,
    private_key
  FROM arach_wallet`)
	defer rows.Close()
	if err != nil {
		panic(err)
	}

	if !rows.Next() {
		w := GenerateWallet()
		rows.Close()
		StoreWallet(w)
		return w
	}

	var private_key string
	var public_key string

	err = rows.Scan(
		&public_key,
		&private_key,
	)
	if err != nil {
		panic("Couldn't get wallet")
	}

	return FromKeyStrings(public_key, private_key)
}

func FetchHighestBlock() block.Block {
	if Conn == nil {
		panic("Database connection not initialised")
	}

	rows, err := Conn.Query(`SELECT
    hash,
    height,
    previous,
    work
  FROM arach_block ORDER BY height DESC, hash`)

	if err != nil {
		panic(err)
	}

	if !rows.Next() {
		panic("Missing highest block")
	}

	block := blockFromRows(rows)
	return block

}

// Fetch a block from the database from the hash
// Returns nil if not found
func FetchBlock(hash string) *block.Block {
	if Conn == nil {
		panic("Database connection not initialised")
	}

	rows, err := Conn.Query(`SELECT
    hash,
    height,
    previous,
    work
  FROM arach_block WHERE hash=?`, hash)

	if err != nil {
		panic(err)
	}

	if !rows.Next() {
		return nil
	}

	block := blockFromRows(rows)
	return &block
}

func blockFromRows(rows *sql.Rows) block.Block {
	var hash string
	var height uint32
	var previous string
	var work uint32

	err := rows.Scan(
		&hash,
		&height,
		&previous,
		&work,
	)

	if err != nil {
		panic(err)
	}
	rows.Close()

	block := block.Block{
		previous,
		work,
		height,
		FetchBlockTransactions(hash),
	}

	if block.HashString() != hash {
		panic("Fetched hash is wrong! DB corrupt!")
	}

	return block
}

func StoreBlock(b block.Block) {
	// Ignore blocks we already have
	if FetchBlock(b.HashString()) != nil {
		return
	}

	prep, err := Conn.Prepare(`
    INSERT INTO arach_block (
      hash,
      height,
      previous,
      work
    ) values (
      ?,?,?,?
    )
  `)

	if err != nil {
		panic(err)
	}

	_, err = prep.Exec(
		b.HashString(),
		b.Height,
		b.Previous,
		b.Work,
	)

	if err != nil {
		panic(err)
	}

	StoreTransactions(b, b.Transactions)
}

func StoreTransactions(b block.Block, ts []transaction.Transaction) {
	for n, t := range ts {
		StoreTransaction(b, n, t)
	}
}

func StoreTransaction(b block.Block, order int, t transaction.Transaction) {
	prep, err := Conn.Prepare(`
    INSERT INTO arach_transaction (
      hash,
      input,
      output,
      amount,
      signature,
      'unique_string',
      'order',
      block,
      block_height
    ) values (
      ?,?,?,?,?,?,?,?,?
    )
  `)

	if err != nil {
		panic(err)
	}

	_, err = prep.Exec(
		t.HashString(),
		t.Input,
		t.Output,
		t.Amount,
		t.Signature,
		t.Unique,
		order,
		b.HashString(),
		b.Height,
	)

	if err != nil {
		panic(err)
	}
}

func FetchTransactionsForAccount(account string) []transaction.Transaction {
	results := make([]transaction.Transaction, 0)
	if Conn == nil {
		panic("Database connection not initialised")
	}

	latest := FetchHighestBlock()
	queryArgs := []interface{}{account, account}

	for _, hash := range GetBlockHashChain(&latest) {
		queryArgs = append(queryArgs, hash)
	}

	rows, err := Conn.Query(`SELECT
      input,
      output,
      amount,
      signature,
      unique_string
    FROM 'arach_transaction' WHERE (input=? OR output=?) AND block in (`+strings.Join(strings.Split(strings.Repeat("?", len(queryArgs)-2), ""), ",")+`) ORDER BY 'block_height' asc, 'order' asc`, queryArgs...)
	defer rows.Close()

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var input string
		var output string
		var amount uint32
		var signature string
		var unique string

		err = rows.Scan(
			&input,
			&output,
			&amount,
			&signature,
			&unique,
		)
		results = append(results, transaction.Transaction{
			input,
			output,
			amount,
			signature,
			unique,
		})
	}

	return results
}

func ValidateBlock(b block.Block) bool {
	// Always trust the genesis block
	if b.HashString() == block.GenesisBlock.HashString() {
		return true
	} else {
		// Genesis block doesn't have correct blockhash
		if b.Height < 1 {
			return false
		}

		if !work.ValidateBlockWork(b) {
			log.Printf("Bad work")
			return false
		}
		// Ensure there's only 1 blockreward issued per block and that it's
		// for the correct blockreward amount.
		hasReward := false
		for _, t := range b.Transactions {
			if t.Input == "blockReward" && (hasReward || t.Amount != block.BlockReward) {
				log.Printf("Bad reward")
				return false
			}
		}

		// Get list of block hashes back to genesis
		hashChain := GetBlockHashChain(&b)
		// Missing link in the chain - this is an invalid block
		// until we get the missing links.
		if hashChain == nil {
			log.Printf("Bad hashchain")
			return false
		}

		// Finally, check the transactions from genesis to
		// now all make sense.
		return VerifyTransactionsInChain(hashChain)
	}
}

func GetBlockHashChain(b *block.Block) []string {
	results := []string{b.HashString()}

	for b.Previous != block.GenesisBlock.Previous {
		b = FetchBlock(b.Previous)
		if b == nil {
			return nil
		}
		results = append(results, b.HashString())
	}
	return results
}

func VerifyTransactionsInChain(blockHashes []string) bool {
	ts := GetTransactionsForHashes(blockHashes)
	balances := make(map[string]uint32)

	for _, t := range ts {
		// Assume Blockrewards are valid. These should be checked
		// in the block itself.
		if t.Input == "blockReward" {
			continue
		}

		if t.Amount > balances[t.Input] {
			log.Printf("Bad amount")
			return false
		}
		balances[t.Input] -= t.Amount
		balances[t.Output] += t.Amount
	}
	return true
}

func GetTransactionsForHashes(blockHashes []string) []transaction.Transaction {
	results := make([]transaction.Transaction, 0)
	if len(blockHashes) == 0 {
		return results
	}
	if Conn == nil {
		panic("Database connection not initialised")
	}

	var queryArgs []interface{}
	for _, hash := range blockHashes {
		queryArgs = append(queryArgs, hash)
	}

	rows, err := Conn.Query(`SELECT
      input,
      output,
      amount,
      signature,
      unique_string
    FROM 'arach_transaction' WHERE block in (`+strings.Join(strings.Split(strings.Repeat("?", len(queryArgs)), ""), ",")+`) ORDER BY 'block_height' asc, 'order' asc`, queryArgs...)
	defer rows.Close()

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var input string
		var output string
		var amount uint32
		var signature string
		var unique string

		err = rows.Scan(
			&input,
			&output,
			&amount,
			&signature,
			&unique,
		)
		results = append(results, transaction.Transaction{
			input,
			output,
			amount,
			signature,
			unique,
		})
	}

	return results
}

func FetchBlockTransactions(blockHash string) []transaction.Transaction {
	results := make([]transaction.Transaction, 0)
	if Conn == nil {
		panic("Database connection not initialised")
	}

	rows, err := Conn.Query(`SELECT
      input,
      output,
      amount,
      signature,
      unique_string
    FROM 'arach_transaction' WHERE block=?`, blockHash)
	defer rows.Close()

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var input string
		var output string
		var amount uint32
		var signature string
		var unique string

		err = rows.Scan(
			&input,
			&output,
			&amount,
			&signature,
			&unique,
		)
		results = append(results, transaction.Transaction{
			input,
			output,
			amount,
			signature,
			unique,
		})
	}

	return results
}
