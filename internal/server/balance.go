package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	tron "github.com/0xAF4/lib_tron"
	solcli "github.com/blocto/solana-go-sdk/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)



func getETHBalance(addr string) (interface{}, error) {
	client, err := ethclient.Dial(ETH_RPC_URL)
	if err != nil {
		return nil, errors.New("DIAL_ERROR")
	}
	defer client.Close()

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(addr), nil)
	if err != nil {
		return nil, err
	}

	f, _ := balance.Float64()
	return f * 0.000000000000000001, nil
}

func getSOLBalance(addr string) (interface{}, error) {
	client := solcli.NewClient(SOL_RPC_URL)
	balance, err := client.GetBalance(context.Background(), addr)
	if err != nil {
		return nil, err
	}

	f := float64(balance)
	return f * 0.000000001, nil
}

func getTRXBalance(addr string) (interface{}, error) {
	node := tron.NewGrpcClient(TRX_GRPC_URL)
	if err := node.Start(); err != nil {
		return nil, err
	}
	defer node.Conn.Close()

	account, err := node.GetAccount(addr)
	if err != nil {
		return nil, err
	}

	f := float64(account.Balance)
	return f * 0.000001, nil
}

func getAVAXBalance(addr string) (interface{}, error) {
	client, err := ethclient.Dial(AVAX_RPC_URL)
	if err != nil {
		return nil, errors.New("DIAL_ERROR")
	}
	defer client.Close()

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(addr), nil)
	if err != nil {
		return nil, err
	}

	f, _ := balance.Float64()
	return f * 0.000000000000000001, nil
}

func getARBBalance(addr string) (interface{}, error) {
	client, err := ethclient.Dial(ARB_RPC_URL)
	if err != nil {
		return nil, errors.New("DIAL_ERROR")
	}
	defer client.Close()

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(addr), nil)
	if err != nil {
		return nil, err
	}

	f, _ := balance.Float64()
	return f * 0.000000000000000001, nil
}

func getBNBBalance(addr string) (interface{}, error) {
	client, err := ethclient.Dial(BNB_RPC_URL)
	if err != nil {
		return nil, errors.New("DIAL_ERROR")
	}
	defer client.Close()

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(addr), nil)
	if err != nil {
		return nil, err
	}

	f, _ := balance.Float64()
	return f * 0.000000000000000001, nil
}

func getETCBalance(addr string) (interface{}, error) {
	client, err := ethclient.Dial(ETC_RPC_URL)
	if err != nil {
		return nil, errors.New("DIAL_ERROR")
	}
	defer client.Close()

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(addr), nil)
	if err != nil {
		return nil, err
	}

	f, _ := balance.Float64()
	return f * 0.000000000000000001, nil
}

func getALGOBalance(addr string) (interface{}, error) {
	url := fmt.Sprintf(ALGO_HTTP_URL, addr)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Standart Coin Processor")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Response code error: %d", resp.StatusCode)
	}

	var account AccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, err
	}

	return float64(account.Amount) * 0.000001, nil
}

func getADABalance(addr string) (interface{}, error) {
	url := fmt.Sprintf(ADA_HTTP_URL, addr)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Standart Coin Processor")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Response code error: %d", resp.StatusCode)
	}

	var txs []Transaction
	if err := json.NewDecoder(resp.Body).Decode(&txs); err != nil {
		return nil, err
	}

	if len(txs) == 0 {
		return 0, nil
	}

	balance := 0
	for _, tx := range txs {
		intValue, err := strconv.ParseInt(tx.Value, 10, 64)
		if err != nil {
			return nil, err
		}
		balance += int(intValue)
	}

	return float64(balance) * 0.000001, nil
}
