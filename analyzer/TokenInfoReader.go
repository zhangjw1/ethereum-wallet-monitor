package analyzer

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TokenInfoReader 代币信息读取器
type TokenInfoReader struct {
	client *ethclient.Client
}

// TokenInfo 代币基本信息
type TokenInfo struct {
	Name        string
	Symbol      string
	Decimals    uint8
	TotalSupply *big.Int
	IsValid     bool
}

// NewTokenInfoReader 创建代币信息读取器
func NewTokenInfoReader(rpcURL string) (*TokenInfoReader, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum client: %w", err)
	}
	return &TokenInfoReader{client: client}, nil
}

// ReadTokenInfo 读取代币基本信息
func (r *TokenInfoReader) ReadTokenInfo(tokenAddress string) (*TokenInfo, error) {
	addr := common.HexToAddress(tokenAddress)
	info := &TokenInfo{IsValid: false}

	// 检查是否是合约
	code, err := r.client.CodeAt(context.Background(), addr, nil)
	if err != nil || len(code) == 0 {
		return info, fmt.Errorf("not a contract address")
	}

	// 尝试读取 name
	name, err := r.callStringMethod(addr, "name()")
	if err == nil {
		info.Name = name
	}

	// 尝试读取 symbol
	symbol, err := r.callStringMethod(addr, "symbol()")
	if err == nil {
		info.Symbol = symbol
	}

	// 尝试读取 decimals
	decimals, err := r.callUint8Method(addr, "decimals()")
	if err == nil {
		info.Decimals = decimals
	} else {
		info.Decimals = 18 // 默认值
	}

	// 尝试读取 totalSupply
	totalSupply, err := r.callUint256Method(addr, "totalSupply()")
	if err == nil {
		info.TotalSupply = totalSupply
	}

	// 如果至少有 symbol 和 totalSupply，认为是有效的 ERC20
	if info.Symbol != "" && info.TotalSupply != nil {
		info.IsValid = true
	}

	return info, nil
}

// IsERC20Token 检查是否是 ERC20 代币
func (r *TokenInfoReader) IsERC20Token(contractAddress string) bool {
	info, err := r.ReadTokenInfo(contractAddress)
	if err != nil {
		return false
	}
	return info.IsValid
}

// callStringMethod 调用返回 string 的方法
func (r *TokenInfoReader) callStringMethod(contract common.Address, method string) (string, error) {
	// 构造方法签名
	methodID := common.Hex2Bytes(methodSignature(method))

	// 调用合约
	msg := ethereum.CallMsg{
		To:   &contract,
		Data: methodID,
	}

	result, err := r.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return "", err
	}

	// 解析返回值
	if len(result) == 0 {
		return "", fmt.Errorf("empty result")
	}

	// 尝试解析为 string
	stringType, _ := abi.NewType("string", "", nil)
	unpacked, err := abi.Arguments{{Type: stringType}}.Unpack(result)
	if err != nil {
		return "", err
	}

	if len(unpacked) > 0 {
		if str, ok := unpacked[0].(string); ok {
			return str, nil
		}
	}

	return "", fmt.Errorf("failed to parse string")
}

// callUint8Method 调用返回 uint8 的方法
func (r *TokenInfoReader) callUint8Method(contract common.Address, method string) (uint8, error) {
	methodID := common.Hex2Bytes(methodSignature(method))

	msg := ethereum.CallMsg{
		To:   &contract,
		Data: methodID,
	}

	result, err := r.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, fmt.Errorf("empty result")
	}

	// uint8 通常以 uint256 形式返回
	value := new(big.Int).SetBytes(result)
	return uint8(value.Uint64()), nil
}

// callUint256Method 调用返回 uint256 的方法
func (r *TokenInfoReader) callUint256Method(contract common.Address, method string) (*big.Int, error) {
	methodID := common.Hex2Bytes(methodSignature(method))

	msg := ethereum.CallMsg{
		To:   &contract,
		Data: methodID,
	}

	result, err := r.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("empty result")
	}

	return new(big.Int).SetBytes(result), nil
}

// methodSignature 计算方法签名的前4字节
func methodSignature(method string) string {
	// 简化版：直接返回常见方法的签名
	signatures := map[string]string{
		"name()":        "06fdde03",
		"symbol()":      "95d89b41",
		"decimals()":    "313ce567",
		"totalSupply()": "18160ddd",
		"owner()":       "8da5cb5b",
	}

	if sig, ok := signatures[method]; ok {
		return sig
	}

	return strings.Repeat("0", 8)
}

// Close 关闭客户端
func (r *TokenInfoReader) Close() {
	r.client.Close()
}
