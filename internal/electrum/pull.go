package electrum

import (
	"encoding/json"
	"errors"
	"sync"
)

type Pull struct {
	RPC map[int]*UniversalRPC
	com sync.Mutex
}

type Status int

type Request struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     int      `json:"id"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	ID      int         `json:"id"`
}

type UniversalRPC struct {
	ID int
	ST Status

	Request
	Response
}

const (
	NotFound Status = iota
	Created
	Sended
	Done
	Error
)

func NewPull() *Pull {
	return &Pull{
		RPC: make(map[int]*UniversalRPC),
	}
}

func (p *Pull) WriteInPull(r *UniversalRPC) {
	p.com.Lock()
	defer p.com.Unlock()
	p.RPC[r.ID] = r
}

func (p *Pull) ReadFromPull(id int) *UniversalRPC {
	p.com.Lock()
	defer p.com.Unlock()
	if r, ok := p.RPC[id]; ok {
		return r
	} else {
		return nil
	}
}

func (p *Pull) DeleteFromPull(id int) {
	p.com.Lock()
	defer p.com.Unlock()
	delete(p.RPC, id)
}

func (p *Pull) Iterate(fn func(u *UniversalRPC)) {
	p.com.Lock()
	defer p.com.Unlock()

	for _, rpc := range p.RPC {
		fn(rpc)
	}
}

func (p *Pull) ResetRecord(id int) {
	p.com.Lock()
	defer p.com.Unlock()
	rec, _ := p.RPC[id]
	rec.ID = GenID()
	rec.ST = Created
	p.RPC[rec.ID] = rec
	delete(p.RPC, id)
}

func (u *UniversalRPC) BuildPayload() []byte {
	payload := Request{
		Method: u.Request.Method,
		Params: u.Request.Params,
		ID:     u.ID,
	}
	jsonData, _ := json.Marshal(payload)
	return jsonData
}

func (u *Response) ExtractFee() (fee float64, err error) {
	defer func() {
		if pErr := recover(); pErr != nil {
			err = errors.New("Extract fee errror")
		}
	}()

	fee = u.Result.(float64)
	return
}

func (r *Response) IsOkResult() (ok bool) {
	_, ok = r.Result.(map[string]interface{})
	return
}

func (r *Response) ExtractBalances() (confirmed float64, unconfirmed float64, err error) {
	defer func() {
		if pErr := recover(); pErr != nil {
			err = errors.New("Extract balance errror")
		}
	}()

	firmed := r.Result.(map[string]interface{})
	confirmed = firmed["confirmed"].(float64) * 0.00000001
	unconfirmed = firmed["unconfirmed"].(float64) * 0.00000001
	return
}
