package main

import (
	"OmniSwapParser/config"
	"OmniSwapParser/core"
	"flag"
	"fmt"
	"sync"
)

func main() {
	config.GetChainByStargateChainId(1)

	h := flag.String("h", "", "tx hash")
	c := flag.String("c", "", "chain name, eg: bsc,ethereum,eth,op,avax")
	d := flag.Bool("d", true, "with detail info")
	flag.Parse()

	if nil == h || *h == "" {
		panic("please input tx hash, -h txhash")
	}

	if nil == c || *c == "" {
		// 从所有链上进行查询
		allChain := config.GetAllChains()
		wg := sync.WaitGroup{}
		for _, chain := range allChain {
			if chain.Rpc == "" {
				continue
			}
			wg.Add(1)
			go func(tmpChain config.ChainInfo) {
				defer func() {
					wg.Done()
				}()
				core.ParseTxOnChain(&tmpChain, *h, *d)
			}(chain)
		}
		wg.Wait()
	} else {
		chain := config.GetChainByName(*c)
		if nil == chain {
			fmt.Printf("unsupport chain: %s\n", *c)
			return
		}
		core.ParseTxOnChain(chain, *h, *d)
	}
}
