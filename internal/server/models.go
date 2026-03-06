package server

type AccountResponse struct {
	Address                     string `json:"address"`
	Amount                      int64  `json:"amount"`
	AmountWithoutPendingRewards int64  `json:"amount-without-pending-rewards"`
	AppsTotalSchema             struct {
		NumByteSlice int `json:"num-byte-slice"`
		NumUint      int `json:"num-uint"`
	} `json:"apps-total-schema"`
	MinBalance         int64  `json:"min-balance"`
	PendingRewards     int64  `json:"pending-rewards"`
	RewardBase         int64  `json:"reward-base"`
	Rewards            int64  `json:"rewards"`
	Round              int64  `json:"round"`
	Status             string `json:"status"`
	TotalAppsOptedIn   int    `json:"total-apps-opted-in"`
	TotalAssetsOptedIn int    `json:"total-assets-opted-in"`
	TotalCreatedApps   int    `json:"total-created-apps"`
	TotalCreatedAssets int    `json:"total-created-assets"`
}

type Transaction struct {
	Address       string `json:"address"`
	TxHash        string `json:"txHash"`
	Index         int    `json:"index"`
	Value         string `json:"value"`
	Block         string `json:"block"`
	Height        int    `json:"height"`
	IncludedAt    int64  `json:"includedAt"`
	Confirmations int    `json:"confirmations"`
}

type GetBalanceRequest struct {
	Address string `json:"address"`
	ViewKey string `json:"view_key"`
}

type GetBalanceResponse struct {
	Balance         uint64 `json:"balance"`
	UnlockedBalance uint64 `json:"unlocked_balance"`
}

type SendFundsStruct struct {
	Coin       string  `json:"coin"`
	PrivateKey string  `json:"private_key"`
	Address    string  `json:"address"`
	Amount     float64 `json:"amount"`
	Fee        float64 `json:"fee"`
}

const (
	ETH_RPC_URL  = "https://ethereum-rpc.publicnode.com"
	SOL_RPC_URL  = "https://solana-rpc.publicnode.com"
	TRX_GRPC_URL = "grpc.trongrid.io:50051"
	AVAX_RPC_URL = "https://avalanche-c-chain-rpc.publicnode.com"
	ARB_RPC_URL  = "https://arbitrum-one-rpc.publicnode.com"
	BNB_RPC_URL  = "https://bsc-rpc.publicnode.com"
	ETC_RPC_URL  = "https://besu-at.etc-network.info"
)

const (
	ALGO_HTTP_URL = "https://mainnet-api.4160.nodely.dev/v2/accounts/%s?format=json&exclude=all"
	ADA_HTTP_URL  = "https://ada.coin.space/api/v1/addresses/%s/utxos"
)
