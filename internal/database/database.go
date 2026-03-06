package database

import (
	"errors"
	"os"

	l_db "github.com/0xAF4/lib_db"
)

type Database struct {
	db *l_db.DB
}

type Node struct {
	Address string
	Port    int32
}

type Blocker interface {
	GetChainName() string
	ConvertToDBBlock() *Block
}

type Block struct {
	Hash         string
	PreviousHash string
	Tx_outs      []Tx
}

type Tx struct {
	Address string
	Balance float64
}

const (
	GetChainNodes = `
		SELECT CP.ADDRESS, CP.PORT FROM CHAIN."CURRENCY" C
		INNER JOIN CHAIN."CURRENCY$PEER" CP ON C.ID = CP.ID
		WHERE C.NAME = $1 AND COALESCE(CP.ARH, 0) = 0
	`

	GetCoinHeight = `
		SELECT c.last_height, c.last_hash 
		FROM chain."CURRENCY" c
		WHERE c.name = $1
	`
)

func New(driver string, connStr string, fl *os.File) (*Database, error) {
	db, err := l_db.New(&l_db.DBConfig{
		"driver":           l_db.StrToDBType(driver),
		"connectionString": connStr,
		"logFile":          fl,
	})

	if err != nil {
		return nil, err
	}

	d := &Database{
		db: db,
	}
	return d, nil
}

func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) GetNodeAddrs(coin string) (*[]Node, error) {
	res, err := d.db.QueryRow(l_db.TxRead, GetChainNodes, coin)
	if err != nil {
		return nil, err
	}

	ns := []Node{}

	for _, r := range *res {
		v, _ := r["port"].(int32)
		ns = append(ns, Node{
			Address: r["address"].(string),
			Port:    v,
		})
	}

	return &ns, nil
}

func (d *Database) GetChainHeight(coin string) (int32, string, error) {
	res, err := d.db.QueryRow(l_db.TxRead, GetCoinHeight, coin)
	if err != nil {
		return 0, "", err
	}

	if res.Count() != 1 {
		return 0, "", errors.New("NOT_ACCEPTABLE_RES_COUNT")
	}

	height := (*res)[0]["last_height"].(int32)
	hash := (*res)[0]["last_hash"].(string)
	return height, hash, nil
}

func (d *Database) InsertBlock(call Blocker) error {
	_, err := d.db.QueryRow(l_db.TxWrite, `SELECT "chain"."BLOCK_INSERT"($1, $2)`, call.GetChainName(), call.ConvertToDBBlock())
	return err
}

func (d *Database) GetBalance(coin string, addr string, view_key *string) (interface{}, error) {
	res, err := d.db.QueryRow(l_db.TxRead, `SELECT chain."GET_BALANCE"($1, $2, $3) balance`, coin, addr, view_key)
	if err != nil {
		return 0, err
	}
	return (*res)[0]["balance"], nil
}
