package main

import (
	"OmniSwapParser/config"
	"OmniSwapParser/core"
	"flag"
	"fmt"
)

func main() {
	config.GetChainByStargateChainId(1)

	h := flag.String("h", "", "tx hash")
	c := flag.String("c", "", "chain name, eg: bsc,ethereum,eth,op,avax")
	flag.Parse()

	if nil == h {
		panic("please input tx hash, -h txhash")
	}

	if nil == c {
		// 从所有链上进行查询
		allChain := config.GetAllChains()
		for _, chain := range allChain {
			go func(tmpChain config.ChainInfo) {
				core.ParseTxOnChain(&tmpChain, *h)
			}(chain)
		}
	} else {
		chain := config.GetChainByName(*c)
		if nil == chain {
			fmt.Printf("unsupport chain: %s\n", *c)
			return
		}
		core.ParseTxOnChain(chain, *h)
	}
}
