package electrum

import (
	"errors"
	"log"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

type SockInterface interface {
	Write(b []byte) (int, error)
	Read(b []byte) (int, error)
	Close() error
}

type ElectrumNode struct {
	Nodes *[]Node
	log   *log.Logger
	Conn  SockInterface
	pull  *Pull
	mus   sync.Mutex

	strBuffer string

	connected bool
	destroy   bool
	net       string
}

const (
	GET_BALANCE = "blockchain.scripthash.get_balance"
	GET_FEE     = "blockchain.estimatefee"
	GET_TXCOUNT = "blockchain.scripthash.get_history"
)

const (
	SSL = iota
	TCP
)

func New(coin string, lg *os.File) (*ElectrumNode, error) {
	var err error
	electrum := ElectrumNode{
		pull: NewPull(),
		net:  coin,
	}

	if electrum.Nodes, err = getNodes(coin); err != nil {
		return nil, err
	}

	if lg != nil {
		electrum.log = log.New(lg, "[Electrum("+coin+")] ", log.Ldate|log.Lmicroseconds)
	}

	if err = electrum.connect(); err != nil {
		return nil, err
	}

	return &electrum, nil
}

func (e *ElectrumNode) Close() {
	e.destroy = true
}

func (e *ElectrumNode) RunLoops() {
	go e.writer()
	go e.reader()
}

func (e *ElectrumNode) GetBalance(addr string) (interface{}, error) {
	scripthashaddr, err := ScripthashAddr(addr, strings.ToLower(e.net))
	if err != nil {
		return nil, err
	}

	RPC := UniversalRPC{
		ID: GenID(),
		ST: Created,
		Request: Request{
			Method: GET_BALANCE,
			Params: []string{scripthashaddr},
		},
	}
	e.pull.WriteInPull(&RPC)

	for slices.Contains([]Status{Created, Sended}, RPC.ST) {
		time.Sleep(time.Millisecond * 100)
	}

	if RPC.ST == Error {
		return nil, errors.New("RPC Error")
	}

	confirmed, _, err := RPC.ExtractBalances()
	if err != nil {
		return nil, err
	}

	return confirmed, err
}

func (e *ElectrumNode) GetFee() (interface{}, error) {
	RPC := UniversalRPC{
		ID: GenID(),
		ST: Created,
		Request: Request{
			Method: GET_FEE,
			Params: []string{"1"},
		},
	}
	e.pull.WriteInPull(&RPC)

	for slices.Contains([]Status{Created, Sended}, RPC.ST) {
		time.Sleep(time.Millisecond * 100)
	}

	if RPC.ST == Error {
		return nil, errors.New("RPC Error")
	}

	fee, err := RPC.ExtractFee()
	if err != nil {
		return nil, err
	}

	return fee, nil
}

func (e *ElectrumNode) GetTransactions(addr string) (interface{}, error) {
	scripthashaddr, err := ScripthashAddr(addr, strings.ToLower(e.net))
	if err != nil {
		return nil, err
	}

	RPC := UniversalRPC{
		ID: GenID(),
		ST: Created,
		Request: Request{
			Method: GET_TXCOUNT,
			Params: []string{scripthashaddr},
		},
	}
	e.pull.WriteInPull(&RPC)

	for slices.Contains([]Status{Created, Sended}, RPC.ST) {
		time.Sleep(time.Millisecond * 100)
	}

	if RPC.ST == Error {
		return nil, errors.New("RPC Error")
	}

	return RPC, nil
}

func (e *ElectrumNode) ExecuteCommand(method string, params []string) (interface{}, error) {
	RPC := UniversalRPC{
		ID: GenID(),
		ST: Created,
		Request: Request{
			Method: method,
			Params: params,
		},
	}
	e.pull.WriteInPull(&RPC)

	for slices.Contains([]Status{Created, Sended}, RPC.ST) {
		time.Sleep(time.Millisecond * 100)
	}

	if RPC.ST == Error {
		return nil, errors.New("RPC Error")
	}

	return RPC, nil
}
