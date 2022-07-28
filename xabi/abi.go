package xabi

import (
	"bytes"
	_ "embed"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

//go:embed SoDiamond.json
var soDiamond []byte

//go:embed ISwapRouter.json
var iSwapRouter []byte

//go:embed IUniswapV2Router02.json
var iUniswapV2Router02 []byte

//go:embed IUniswapV2Router02AVAX.json
var iUniswapV2Router02AVAX []byte

//go:embed ERC20.json
var erc20 []byte

var (
	SoDiamond              abi.ABI
	ISwapRouter            abi.ABI
	IUniswapV2Router02     abi.ABI
	IUniswapV2Router02AVAX abi.ABI
	ERC20                  abi.ABI
)

func init() {
	var err error
	SoDiamond, err = abi.JSON(bytes.NewReader(soDiamond))
	if err != nil {
		panic(err)
	}
	ISwapRouter, err = abi.JSON(bytes.NewReader(iSwapRouter))
	if err != nil {
		panic(err)
	}
	IUniswapV2Router02, err = abi.JSON(bytes.NewReader(iUniswapV2Router02))
	if err != nil {
		panic(err)
	}
	IUniswapV2Router02AVAX, err = abi.JSON(bytes.NewReader(iUniswapV2Router02AVAX))
	if err != nil {
		panic(err)
	}
	ERC20, err = abi.JSON(bytes.NewReader(erc20))
	if err != nil {
		panic(err)
	}
}
