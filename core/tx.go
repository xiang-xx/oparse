package core

import (
	"context"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/xiang-xx/oparse/config"
	"github.com/xiang-xx/oparse/xabi"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/fatih/color"
	"github.com/shopspring/decimal"
)

const alignment = 25

var withDetail = false

// SoData 代表单笔 swap 数据，用于链路 swap 追踪
type SoData struct {
	TransactionId      [32]byte // 唯一交易ID， 32字节
	Receiver           common.Address
	SourceChainId      *big.Int
	SendingAssetId     common.Address
	DestinationChainId *big.Int
	ReceivingAssetId   common.Address
	Amount             *big.Int
}

// 1. 通用Uniswap/PancakeSwap数据结构
// 2. 代表用fromAmount数量的sendingAssetId换取receivingAssetId
// 3. 从数据流图来看，用SwapData来表示在source swap上从ETH换USDC;
type SwapData struct {
	CallTo           common.Address
	ApproveTo        common.Address
	SendingAssetId   common.Address // eth 是传 0 地址，不管是 v2 还是 v3
	ReceivingAssetId common.Address // token address, eth 是 0 地址, v3 swap 是 weth，不能是 0 地址
	FromAmount       *big.Int       // swap start token amount
	CallData         []byte         //  The swap callData callData = abi.encodeWithSignature("swapExactETHForTokens", minAmount, [sendingAssetId, receivingAssetId], 以太坊SoDiamond地址, deadline)
}

// StargateData 传给 stargate 的数据
type StargateData struct {
	SrcStargatePoolId  *big.Int       // stargate 源 pool id
	DstStargateChainId uint16         // stargete 目的链 chain id，非 evmchainid，是 stargate 自己定义的 id
	DstStargatePoolId  *big.Int       // stargate 目标链 pool id
	MinAmount          *big.Int       // 目标链最小得到数量
	DstGasForSgReceive *big.Int       // 目的链 sgReceive 消耗的 gas,通过 sgReceiveForGas 预估
	DstSoDiamond       common.Address // 目的链 SoDiamond 地址
}

type GenericInputData struct {
	SoData   SoData
	SwapData []SwapData
}

type SoSwapViaStargateInputData struct {
	SoData       SoData
	SwapDataSrc  []SwapData
	StargateData StargateData
	SwapDataDst  []SwapData
}

type FromTokenSwapInputData struct {
	AmountIn     *big.Int
	AmountOutMin *big.Int
	Path         []common.Address
	To           common.Address
	Deadline     *big.Int
}

type FromBalanceSwapInputData struct {
	AmountOutMin *big.Int
	Path         []common.Address
	To           common.Address
	Deadline     *big.Int
}

type SwapV3InputData struct {
	ExactInputParams ExactInputParams
}

type ExactInputParams struct {
	Path             []byte
	Recipient        common.Address
	Deadline         *big.Int
	AmountIn         *big.Int
	AmountOutMinimum *big.Int
}

type MyReceipt struct {
	// types.Receipt
	ReturnCode string `json:"returnCode"`
	ReturnData string `json:"returnData"`
	Status     string `json:"status`
}

func ParseTxOnChain(chain *config.ChainInfo, txHash string, d bool) {
	withDetail = d
	ctx := context.Background()
	client, err := ethclient.Dial(chain.Rpc)
	if err != nil {
		fmt.Printf("dail rpc %s error: %s\n", chain.Rpc, err)
		return
	}
	// 调用 rpc 获取 交易数据
	hash := common.HexToHash(txHash)
	tx, _, err := client.TransactionByHash(ctx, hash)
	if err != nil {
		if err.Error() == "not found" {
			return
		}
		printError("get tx", err)
		return
	}

	printLine()
	printAlignLine("Chain", chain.ChainName)

	receipt, err := client.TransactionReceipt(ctx, hash)
	if err != nil {
		printError("get receipt", err)
	}

	printLine()
	printTxBaseInfo(chain, tx)
	printReceipt(receipt, chain, hash)
	printLine()

	inputData := tx.Data()
	method, err := xabi.SoDiamond.MethodById(inputData[:4])
	if err != nil {
		printError("MethodById", err)
		return
	}

	if method.RawName == "swapTokensGeneric" {
		err = parseSwapTokenGeneric(client, method, inputData[4:])
		if err != nil {
			printError("parseSwapTokenGeneric", err)
		}
	} else if method.RawName == "soSwapViaStargate" {
		err = parseSoSwapViaStargate(client, method, inputData[4:])
		if err != nil {
			printError("soSwapViaStargate", err)
		}
	}
}

func parseSoSwapViaStargate(client *ethclient.Client, method *abi.Method, methodInput []byte) error {
	values, err := method.Inputs.UnpackValues(methodInput)
	if err != nil {
		return err
	}
	inputStructData := &SoSwapViaStargateInputData{}
	err = method.Inputs.Copy(inputStructData, values)
	if err != nil {
		return err
	}
	printSoData(inputStructData.SoData)
	fromChain := config.GetChainByChainId(int(inputStructData.SoData.SourceChainId.Int64()))
	if nil == fromChain {
		return errors.New("not found from chain")
	}
	toChain := config.GetChainByChainId(int(inputStructData.SoData.DestinationChainId.Int64()))
	if nil == fromChain {
		return errors.New("not found to chain")
	}

	printSwapData("SrcSwap", fromChain, inputStructData.SwapDataSrc)

	printStargateData(fromChain, toChain, inputStructData.StargateData)

	printSwapData("DstSwap", toChain, inputStructData.SwapDataDst)

	return nil
}

func printStargateData(fromChain, toChain *config.ChainInfo, stargateData StargateData) {
	stargatePath := ""
	var fromToken Token
	for _, pool := range fromChain.StargatePool {
		if pool.PoolId == int(stargateData.SrcStargatePoolId.Int64()) {
			stargatePath = stargatePath + fmt.Sprintf("%s(%d)", pool.TokenName, pool.PoolId)
			fromToken = Token{
				Address:  pool.TokenAddress,
				Symbol:   pool.TokenName,
				Decimals: pool.Decimal,
				Name:     pool.TokenName,
			}
		}
	}

	for _, pool := range toChain.StargatePool {
		if pool.PoolId == int(stargateData.DstStargatePoolId.Int64()) {
			stargatePath = stargatePath + fmt.Sprintf(" -> %s(%d)", pool.TokenName, pool.PoolId)
		}
	}
	printAlignLine("Stargate", stargatePath)
	// min amount
	printAlignLine("", alignString("MinAmount", 10)+formatToken(stargateData.MinAmount.String(), fromToken))
	printAlignLine("", alignString("DstGas", 10)+stargateData.DstGasForSgReceive.String())
}

func parseSwapTokenGeneric(client *ethclient.Client, method *abi.Method, methodInput []byte) error {
	values, err := method.Inputs.UnpackValues(methodInput)
	if err != nil {
		return err
	}
	inputStructData := &GenericInputData{}
	err = method.Inputs.Copy(inputStructData, values)
	if err != nil {
		return err
	}
	err = printSoData(inputStructData.SoData)
	if err != nil {
		return err
	}
	fromChain := config.GetChainByChainId(int(inputStructData.SoData.SourceChainId.Int64()))
	if nil == fromChain {
		return errors.New("not found from chain")
	}

	return printSwapData("SrcChain", fromChain, inputStructData.SwapData)
}

func printSwapData(where string, chain *config.ChainInfo, swapData []SwapData) error {
	if len(swapData) == 0 {
		printAlignLine(where, "Not Swapped")
	}
	for _, swapItem := range swapData {
		callTo := swapItem.CallTo.String()
		for _, r := range chain.UniswapRouter {
			if r.RouterAddress == callTo {
				printSwapItem(where, chain, r, swapItem)
			}
		}
	}
	return nil
}

func printSwapItem(where string, chain *config.ChainInfo, router config.UniswapRouter, swapItem SwapData) error {
	if router.Type == "IUniswapV2Router02" || router.Type == "IUniswapV2Router02AVAX" {
		return printSwapV2Item(where, chain, router, swapItem)
	} else if router.Type == "ISwapRouter" {
		return printSwapV3Item(where, chain, router, swapItem)
	}
	return nil
}

func printSwapV2Item(where string, chain *config.ChainInfo, router config.UniswapRouter, swapItem SwapData) error {
	var swapAbi *abi.ABI
	if router.Type == "IUniswapV2Router02" {
		swapAbi = &xabi.IUniswapV2Router02
	} else {
		swapAbi = &xabi.IUniswapV2Router02AVAX
	}

	method, err := swapAbi.MethodById(swapItem.CallData[:4])
	if err != nil {
		return err
	}
	inputValues, err := method.Inputs.Unpack(swapItem.CallData[4:])
	if err != nil {
		return err
	}

	var swapPath []common.Address
	var amoutOutMin *big.Int
	if strings.HasPrefix(method.RawName, "swapExactTokens") {
		res := &FromTokenSwapInputData{}
		err = method.Inputs.Copy(res, inputValues)
		if err != nil {
			return err
		}
		swapPath = res.Path
		amoutOutMin = res.AmountOutMin
	} else {
		res := &FromBalanceSwapInputData{}
		err = method.Inputs.Copy(res, inputValues)
		if err != nil {
			return err
		}
		swapPath = res.Path
		amoutOutMin = res.AmountOutMin
	}
	client, err := ethclient.Dial(chain.Rpc)
	if err != nil {
		return err
	}
	paths := []string{}
	tokens := make([]Token, 0)
	for _, tokenAddress := range swapPath {
		token, err := getTokenInfo(client, chain, tokenAddress)
		if err != nil {
			return err
		}
		paths = append(paths, token.Symbol)
		tokens = append(tokens, token)
	}
	printAlignLine(where, router.Name+"  "+strings.Join(paths, " -> "))
	printAlignLine("", "AmountOutMin  "+formatToken(amoutOutMin.String(), tokens[len(tokens)-1]))
	if withDetail {
		for _, token := range tokens {
			printAlignLine("", alignString(token.Symbol, 7)+token.Address)
		}
	}
	return nil
}

func printSwapV3Item(where string, chain *config.ChainInfo, router config.UniswapRouter, swapItem SwapData) error {
	method, err := xabi.ISwapRouter.MethodById(swapItem.CallData[:4])
	if err != nil {
		return err
	}
	inputValues, err := method.Inputs.Unpack(swapItem.CallData[4:])
	if err != nil {
		return err
	}
	res := &SwapV3InputData{}
	err = method.Inputs.Copy(res, inputValues)
	if err != nil {
		return err
	}

	paths, fees := decodePath(res.ExactInputParams.Path)
	client, err := ethclient.Dial(chain.Rpc)
	if err != nil {
		return err
	}
	pathContent := ""
	tokens := make([]Token, 0)
	for i, tokenAddres := range paths {
		token, err := getTokenInfo(client, chain, tokenAddres)
		if err != nil {
			return err
		}
		if i > 0 {
			pathContent = pathContent + fmt.Sprintf(" --(%.1f%%)--> %s", float32(fees[i-1])/1000, token.Symbol)
		} else {
			pathContent = pathContent + token.Symbol
		}
		tokens = append(tokens, token)
	}

	printAlignLine(where, router.Name+"  "+pathContent)
	printAlignLine("", "AmountOutMin  "+formatToken(res.ExactInputParams.AmountOutMinimum.String(), tokens[len(tokens)-1]))
	if withDetail {
		for _, token := range tokens {
			printAlignLine("", alignString(token.Symbol, 7)+token.Address)
		}
	}

	return nil
}

func printSoData(soData SoData) error {
	fromChain := config.GetChainByChainId(int(soData.SourceChainId.Int64()))
	if nil == fromChain {
		return errors.New("not found from chain")
	}
	toChain := config.GetChainByChainId(int(soData.DestinationChainId.Int64()))
	if nil == fromChain {
		return errors.New("not found to chain")
	}
	fromClient, err := ethclient.Dial(fromChain.Rpc)
	if err != nil {
		return err
	}
	toClient, err := ethclient.Dial(toChain.Rpc)
	if err != nil {
		return err
	}
	fromToken, err := getTokenInfo(fromClient, fromChain, soData.SendingAssetId)
	if err != nil {
		return nil
	}
	toToken, err := getTokenInfo(toClient, toChain, soData.ReceivingAssetId)
	if err != nil {
		return nil
	}

	printAlignLine("TransactionId", hex.EncodeToString(soData.TransactionId[:]))
	printAlignLine("Receiver", soData.Receiver.String())
	printAlignLine("Router", fmt.Sprintf("%s(%s) -> %s(%s)", fromChain.ChainName, fromToken.Symbol, toChain.ChainName, toToken.Symbol))
	printAlignLine("SendTokenAddress", soData.SendingAssetId.Hex())
	printAlignLine("ReceiveTokenAddress", soData.ReceivingAssetId.Hex())
	printAlignLine("Amount", formatToken(soData.Amount.String(), fromToken))
	return nil
}

func formatToken(amount string, token Token) string {
	a, _ := decimal.NewFromString(amount)
	f, _ := a.Div(decimal.NewFromBigInt(big.NewInt(1), int32(token.Decimals))).Float64()
	return fmt.Sprintf("%.15f %s", f, token.Symbol)
}

func printTxBaseInfo(chain *config.ChainInfo, tx *types.Transaction) {
	printAlignLine("Tx Base Info", "")
	printAlignLine("Gas Limit", strconv.Itoa(int(tx.Gas())))
	printAlignLine("Gas Price", tx.GasPrice().String())
	printAlignLine("Value", formatToken(tx.Value().String(), Token{
		Decimals: 18,
		Symbol:   chain.CurrancySymbol,
		Name:     chain.CurrancySymbol,
	}))
}

func printAlignLine(left string, content string) {
	left = alignString(left, alignment)
	fmt.Println(left + color.HiBlueString("%s", content))
}

func alignString(s string, l int) string {
	if len(s) < l {
		s = s + strings.Repeat(" ", l-len(s))
	}
	return s
}

func printReceipt(receipt *types.Receipt, chain *config.ChainInfo, hash common.Hash) {
	if nil == receipt {
		return
	}
	printAlignLine("Status", strconv.Itoa(int(receipt.Status)))

	if receipt.Status == 0 {
		ctx := context.Background()
		var r *MyReceipt
		rpcClient, _ := rpc.DialContext(context.Background(), chain.Rpc)
		err := rpcClient.CallContext(ctx, &r, "eth_getTransactionReceipt", hash)
		if err != nil {
			printError("eth_getTransactionReceipt", err)
		}
		returnData, err := hex.DecodeString(strings.TrimPrefix(r.ReturnData, "0x"))
		if err != nil {
			printError("DecodeString", err)
		}
		// 4 bytes function
		// 32 bytes offset
		// 32 bytes length
		// data
		if len(returnData) < 68 {
			return
		}
		lengthData := big.NewInt(0).SetBytes(common.TrimLeftZeroes(returnData[36:68])).Int64()
		if len(returnData) < int(68+lengthData) {
			return
		}
		printAlignLine("ErrorInfo", string(returnData[68:68+lengthData]))
	}
}

func printLine() {
	fmt.Println("==========================================================")
}

func printError(where string, err error) {
	fmt.Print(color.HiRedString("%s err: %s\n", where, err))
}

const (
	AddrSize = 20
	FeeSize  = 3
	Offset   = AddrSize + FeeSize
	DataSize = Offset + AddrSize
)

// encodePath encode swap v3 path to bytes
// pool fee 默认是 3/1000
func encodePath(path []common.Address, fees []int) (encoded []byte, err error) {
	if len(fees) != len(path)-1 {
		return nil, errors.New("invalid fees")
	}

	encoded = make([]byte, 0, len(fees)*Offset+AddrSize)
	for i := 0; i < len(fees); i++ {
		encoded = append(encoded, path[i].Bytes()...)
		feeBytes := big.NewInt(int64(fees[i])).Bytes()
		feeBytes = common.LeftPadBytes(feeBytes, 3)
		encoded = append(encoded, feeBytes...)
	}
	encoded = append(encoded, path[len(path)-1].Bytes()...)
	return
}

func decodePath(pathByte []byte) ([]common.Address, []int) {
	paths := make([]common.Address, 0)
	fees := make([]int, 0)
	// 20 字节 address
	for i := 0; i < len(pathByte); {
		// 读取 path
		paths = append(paths, common.BytesToAddress(pathByte[i:i+AddrSize]))
		i = i + AddrSize
		if i < len(pathByte) {
			feeBytes := common.TrimLeftZeroes(pathByte[i : i+FeeSize])
			fees = append(fees, int(big.NewInt(0).SetBytes(feeBytes).Int64()))
			i += FeeSize
		}
	}
	return paths, fees
}
