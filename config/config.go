package config

import (
	_ "embed"
	"encoding/json"
)

//go:embed OmniSwapInfo.json
var swapInfo []byte

var chains map[string]ChainInfo

type ChainInfo struct {
	ChainName       string
	Rpc             string
	CurrancySymbol  string
	ChainId         int             `json:"ChainId"`
	SoDiamond       string          `json:"SoDiamond"`
	StargateChainId int             `json:"StargateChainId"`
	UniswapRouter   []UniswapRouter `json:"UniswapRouter"`
	StargatePool    []Pool          `json:"StargatePool"`
}

type UniswapRouter struct {
	Name          string `json:"Name"`
	RouterAddress string `json:"RouterAddress"`
	Type          string `json:"Type"`
}

type Pool struct {
	PoolId       int
	Decimal      int
	TokenAddress string
	TokenName    string
}

func init() {
	chains = make(map[string]ChainInfo, 0)
	err := json.Unmarshal(swapInfo, &chains)
	if err != nil {
		panic(err)
	}

	rpcs := map[int]string{
		1:     "https://rpc.ankr.com/eth",              // eth
		56:    "https://bsc-dataseed3.ninicoin.io",     // bsc
		43114: "https://api.avax.network/ext/bc/C/rpc", // avax-c
		137:   "https://polygon-rpc.com",               // polygon
		42161: "https://rpc.ankr.com/arbitrum",         // arbitrum
		10:    "https://mainnet.optimism.io",           // op
	}
	currancySymbol := map[int]string{
		1:     "ETH",         // eth
		56:    "BNB",         // bsc
		43114: "AVAX",        // avax-c
		137:   "MATIC",       // polygon
		42161: "ARBITRUMETH", // arbitrum
		10:    "OPETH",       // op
	}

	for k, v := range chains {
		v.ChainName = k
		v.Rpc = rpcs[v.ChainId]
		v.CurrancySymbol = currancySymbol[v.ChainId]
		chains[k] = v
	}
}

func GetChainByStargateChainId(stargateChainId int) *ChainInfo {
	for _, c := range chains {
		if c.StargateChainId == stargateChainId {
			return &c
		}
	}
	return nil
}

func GetChainByChainId(chainId int) *ChainInfo {
	for _, c := range chains {
		if c.ChainId == chainId {
			return &c
		}
	}
	return nil
}

func GetAllChains() map[string]ChainInfo {
	return chains
}

func GetChainByName(name string) *ChainInfo {
	var chainId int
	switch name {
	case "eth", "ethereum", "mainnet", "evm":
		chainId = 1
	case "bsc", "binance":
		chainId = 56
	case "avax", "avax-c", "avalanche":
		chainId = 43114
	case "polygon":
		chainId = 137
	case "arbitrum":
		chainId = 42161
	case "op", "optimism":
		chainId = 10
	}
	return GetChainByChainId(chainId)
}
