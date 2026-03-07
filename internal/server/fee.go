package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// ─── EVM-based (ETH, BNB, ARB, AVAX, ETC) ───────────────────────────────────

func getEVMGasPrice(rpcURL string) (float64, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return 0, errors.New("DIAL_ERROR")
	}
	defer client.Close()

	price, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return 0, err
	}

	// Возвращаем в Gwei
	gwei := float64(price.Int64()) / 1e9
	return gwei, nil
}

// getETHFee — рекомендуемая цена газа в Gwei для Ethereum Mainnet.
func getETHFee() (interface{}, error) {
	gwei, err := getEVMGasPrice(ETH_RPC_URL)
	if err != nil {
		return nil, err
	}
	// Стандартный перевод: 21000 gas * price
	return map[string]interface{}{
		"gas_price_gwei": gwei,
		"transfer_eth":   fmt.Sprintf("%.8f", gwei*21000/1e9),
		"unit":           "gwei",
	}, nil
}

func getBNBFee() (interface{}, error) {
	gwei, err := getEVMGasPrice(BNB_RPC_URL)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"gas_price_gwei": gwei,
		"transfer_eth":   fmt.Sprintf("%.8f", gwei*21000/1e9),
		"unit":           "gwei",
	}, nil
}

func getARBFee() (interface{}, error) {
	gwei, err := getEVMGasPrice(ARB_RPC_URL)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"gas_price_gwei": gwei,
		"transfer_eth":   fmt.Sprintf("%.8f", gwei*21000/1e9),
		"unit":           "gwei",
	}, nil
}

func getAVAXFee() (interface{}, error) {
	gwei, err := getEVMGasPrice(AVAX_RPC_URL)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"gas_price_gwei": gwei,
		"transfer_eth":   fmt.Sprintf("%.8f", gwei*21000/1e9),
		"unit":           "gwei",
	}, nil
}

func getETCFee() (interface{}, error) {
	gwei, err := getEVMGasPrice(ETC_RPC_URL)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"gas_price_gwei": gwei,
		"transfer_eth":   fmt.Sprintf("%.8f", gwei*21000/1e9),
		"unit":           "gwei",
	}, nil
}

// ─── SOL ─────────────────────────────────────────────────────────────────────

type solFeeResp struct {
	Result struct {
		Value uint64 `json:"value"`
	} `json:"result"`
}

func getSOLFee() (interface{}, error) {
	body := `{"jsonrpc":"2.0","id":1,"method":"getFeeForMessage","params":["AQABAgIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEBAgAA","latest"]}`

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Post(SOL_RPC_URL, "application/json", jsonReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result solFeeResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	lamports := result.Result.Value
	sol := float64(lamports) * 0.000000001
	return map[string]interface{}{
		"lamports":     lamports,
		"sol":          sol,
		"typical_tx":   "5000 lamports",
		"unit":         "lamports",
	}, nil
}

// ─── TRX ─────────────────────────────────────────────────────────────────────

type trxFeeResp struct {
	BandwidthFee  int64 `json:"bandwidth_fee"`
	EnergyFee     int64 `json:"energy_fee"`
}

func getTRXFee() (interface{}, error) {
	// TRX: стандартный перевод TRX стоит ~0.1-1 TRX bandwidth
	// Используем Trongrid REST API
	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequest("GET", "https://api.trongrid.io/wallet/getchainparameters", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		// Возвращаем дефолтные значения если API недоступен
		return map[string]interface{}{
			"bandwidth_price": 1000,   // SUN per byte
			"energy_price":    420,    // SUN per energy
			"typical_transfer_trx": 1.0,
			"unit": "SUN",
			"note": "default values",
		}, nil
	}
	defer resp.Body.Close()

	var params struct {
		ChainParameter []struct {
			Key   string `json:"key"`
			Value int64  `json:"value"`
		} `json:"chainParameter"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&params); err != nil {
		return nil, err
	}

	result := map[string]interface{}{"unit": "SUN"}
	for _, p := range params.ChainParameter {
		switch p.Key {
		case "getTransactionFee":
			result["bandwidth_price"] = p.Value
		case "getEnergyFee":
			result["energy_price"] = p.Value
		}
	}
	return result, nil
}

// ─── BTC (через Electrum estimatefee) ────────────────────────────────────────

func getBTCFee() (interface{}, error) {
	// Используем mempool.space API как fallback
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get("https://mempool.space/api/v1/fees/recommended")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var fees map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&fees); err != nil {
		return nil, err
	}
	return fees, nil
}

// ─── XMR ─────────────────────────────────────────────────────────────────────

type xmrFeeResp struct {
	Result struct {
		Fee      uint64 `json:"fee"`
		Fees     []uint64 `json:"fees"`
		Credits  uint64 `json:"credits"`
	} `json:"result"`
}

func getXMRFee() (interface{}, error) {
	body := `{"jsonrpc":"2.0","id":"0","method":"get_fee_estimate"}`
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Post("https://xmrnode.moneroworld.com:18089/json_rpc", "application/json", jsonReader(body))
	if err != nil {
		return map[string]interface{}{
			"fee_per_byte":   20000,
			"unit":           "piconero",
			"note":           "default values",
		}, nil
	}
	defer resp.Body.Close()

	var result xmrFeeResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"fee_per_byte": result.Result.Fee,
		"unit":         "piconero",
	}, nil
}

// ─── ALGO ─────────────────────────────────────────────────────────────────────

type algoParamsResp struct {
	MinFee       uint64 `json:"min-fee"`
	FlatFee      bool   `json:"flat-fee"`
}

func getALGOFee() (interface{}, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get("https://mainnet-api.4160.nodely.dev/v2/transactions/params")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var params map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&params); err != nil {
		return nil, err
	}

	minFee := float64(0)
	if mf, ok := params["min-fee"].(float64); ok {
		minFee = mf
	}
	return map[string]interface{}{
		"min_fee_microalgo": minFee,
		"algo":              minFee * 0.000001,
		"unit":              "microAlgo",
	}, nil
}

// ─── ADA ─────────────────────────────────────────────────────────────────────

func getADAFee() (interface{}, error) {
	// Cardano: минимальная комиссия = a + b * tx_size
	// a = 155381 lovelace, b = 43.946 lovelace/byte (protocol params)
	// Типичный tx ~350 bytes → ~155381 + 43.946*350 ≈ 170,762 lovelace ≈ 0.17 ADA
	return map[string]interface{}{
		"min_fee_lovelace": 170762,
		"ada":              0.170762,
		"formula":          "155381 + 43.946 * tx_bytes",
		"typical_tx_bytes": 350,
		"unit":             "lovelace",
	}, nil
}
