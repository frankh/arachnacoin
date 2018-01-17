package store

import (
	"database/sql"
	"github.com/frankh/arachnacoin/block"
	"github.com/frankh/arachnacoin/transaction"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var Conn *sql.DB

func Init(path string) {
	var err error
	Conn, err = sql.Open("sqlite3", path)
	if err != nil {
		panic(err)
	}
	Conn.SetMaxOpenConns(1)

	table_check, err := Conn.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name='block'`)
	if err != nil {
		panic(err)
	}

	if !table_check.Next() {
		log.Printf("Database was empty, creating schema...")
		prep, err := Conn.Prepare(`
      CREATE TABLE 'block' (
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
      CREATE TABLE 'transactions' (
        'hash' TEXT PRIMARY KEY,
        'input' TEXT NOT NULL,
        'output' TEXT NOT NULL,
        'amount' INT NOT NULL,
        'signature' TEXT NOT NULL,
        'unique' TEXT NOT NULL,
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
	}
	table_check.Close()
	rows, err := Conn.Query(`SELECT hash FROM block WHERE hash=?`, block.GenesisBlock.HashString())
	defer rows.Close()
	if err != nil {
		panic(err)
	}
	if !rows.Next() {
		log.Printf("Creating genesis block...")
		StoreBlock(block.GenesisBlock)
	}
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
  FROM block ORDER BY height DESC, hash`)

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
  FROM block WHERE hash=?`, hash)

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
	prep, err := Conn.Prepare(`
    INSERT INTO block (
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
    INSERT INTO transactions (
      hash,
      input,
      output,
      amount,
      signature,
      'unique',
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

	rows, err := Conn.Query(`SELECT
      input,
      output,
      amount,
      signature,
      'unique'
    FROM transactions WHERE input=? OR output=? ORDER BY 'block_height' asc, 'order' asc`, account, account)
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
      'unique'
    FROM transactions WHERE block=?`, blockHash)
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
