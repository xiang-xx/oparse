package core

import (
	"github.com/xiang-xx/oparse/config"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Token struct {
	Name     string
	Symbol   string
	Decimals int
	Address  string
}

var tokenCache map[string]Token

func init() {
	tokenCache = make(map[string]Token, 0)
}

func getTokenInfo(client *ethclient.Client, chain *config.ChainInfo, tokenAddress common.Address) (token Token, err error) {
	cacheKey := chain.ChainName + tokenAddress.Hex()
	if token, ok := tokenCache[cacheKey]; ok {
		return token, nil
	}
	if isZeroAddress(tokenAddress) {
		return Token{
			Name:     chain.CurrancySymbol,
			Symbol:   chain.CurrancySymbol,
			Decimals: 18,
			Address:  tokenAddress.String(),
		}, nil
	}
	instance, err := NewStore(tokenAddress, client)
	if err != nil {
		return
	}
	symbol, err := instance.Symbol(&bind.CallOpts{})
	if err != nil {
		return
	}
	name, err := instance.Name(&bind.CallOpts{})
	if err != nil {
		return
	}
	decimals, err := instance.Decimals(&bind.CallOpts{})
	if err != nil {
		return
	}
	token.Name = name
	token.Symbol = symbol
	token.Decimals = int(decimals)
	token.Address = tokenAddress.String()
	tokenCache[cacheKey] = token
	return
}

func isZeroAddress(address common.Address) bool {
	bs := address.Bytes()
	for _, b := range bs {
		if b != 0 {
			return false
		}
	}
	return true
}
