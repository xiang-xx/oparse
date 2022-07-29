# oparse

Parse [OminiSwap](https://github.com/OmniBTC/OmniSwap) tx input data

### usage

install

```sh
go install github.com/xiang-xx/oparse@latest
```

parse tx by tx hash

```sh
oparse -h 0x977d4fb7c5747b66b52b1e86ab808dc895a3454f28ae54857aabed8dee575514
```

example
```
➜  ~ oparse -h 0xcc682b33b6b9bd30eac50820dfa1f34c0a0ed894cbd1fab05fd2d88714ad344c
==========================================================
Chain                    arbitrum-main
==========================================================
Tx Base Info
Gas Limit                3111971
Gas Price                302862846
Value                    0.002020191587119 ARBITRUMETH
Status                   0
==========================================================
TransactionId            01e91e80d8c6692bc7a42112e03236b70000000062e1ff8bb9587f4d546be0b1
Receiver                 0x0e9D66A7008ca39AE759569Ad1E911d29547E892
Router                   arbitrum-main(ARBITRUMETH) -> polygon-main(MATIC)
SendTokenAddress         0x0000000000000000000000000000000000000000
ReceiveTokenAddress      0x0000000000000000000000000000000000000000
Amount                   0.002000000000000 ARBITRUMETH
SrcSwap                  UniswapV3  WETH --(0.5%)--> USDT --(3.0%)--> UNI --(3.0%)--> USDC
                         AmountOutMin  0.000000000000000 USDC
                         WETH   0x82aF49447D8a07e3bd95BD0d56f35241523fBab1
                         USDT   0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9
                         UNI    0xFa7F8980b0f1E64A2062791cc3b0871572f1F7f0
                         USDC   0xFF970A61A04b1cA14834A43f5dE4533eBDDB5CC8
Stargate                 USDC(1) -> USDC(1)
                         MinAmount 3.349296000000000 USDC
                         DstGas    343678
DstSwap                  UniswapV3  USDC --(0.5%)--> WMATIC
                         AmountOutMin  3.724766091594995 WMATIC
                         USDC   0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174
                         WMATIC 0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270
```
