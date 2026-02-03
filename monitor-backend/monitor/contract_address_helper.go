package monitor

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

// calculateContractAddress 计算合约部署地址
//
// 以太坊合约地址计算公式：
// contract_address = keccak256(rlp([deployer_address, nonce]))[12:]
//
// 参数：
//   - deployer: 部署者地址（十六进制字符串）
//   - nonce: 部署者的交易 nonce
//
// 返回：
//   - 合约地址（十六进制字符串）
//
// 示例：
//
//	addr := calculateContractAddress("0x1234...", 5)
//	// 返回: "0xabcd..."
func calculateContractAddress(deployer string, nonce uint64) string {
	// 解析部署者地址
	deployerAddr := common.HexToAddress(deployer)

	// RLP 编码 [deployer_address, nonce]
	// 注意：这里必须使用 []interface{} 类型
	data, err := rlp.EncodeToBytes([]interface{}{
		deployerAddr.Bytes(),
		nonce,
	})
	if err != nil {
		// RLP 编码失败（理论上不应该发生）
		return ""
	}

	// 计算 Keccak256 哈希
	hash := crypto.Keccak256Hash(data)

	// 取后 20 字节（160 位）作为地址
	// hash 是 32 字节，地址是 20 字节，所以从第 12 字节开始取
	contractAddr := common.BytesToAddress(hash[12:])

	return contractAddr.Hex()
}

// calculateContractAddressCreate2 计算 CREATE2 部署的合约地址
//
// CREATE2 地址计算公式（EIP-1014）：
// contract_address = keccak256(0xff ++ deployer_address ++ salt ++ keccak256(init_code))[12:]
//
// 参数：
//   - deployer: 部署者地址
//   - salt: 32 字节的 salt
//   - initCodeHash: 初始化代码的 keccak256 哈希
//
// 返回：
//   - 合约地址
func calculateContractAddressCreate2(deployer string, salt [32]byte, initCodeHash [32]byte) string {
	deployerAddr := common.HexToAddress(deployer)

	// 构建数据：0xff ++ deployer_address ++ salt ++ init_code_hash
	data := make([]byte, 1+20+32+32)
	data[0] = 0xff
	copy(data[1:21], deployerAddr.Bytes())
	copy(data[21:53], salt[:])
	copy(data[53:85], initCodeHash[:])

	// 计算 Keccak256 哈希
	hash := crypto.Keccak256Hash(data)

	// 取后 20 字节作为地址
	contractAddr := common.BytesToAddress(hash[12:])

	return contractAddr.Hex()
}

// isContractAddress 检查地址是否是合约地址（需要 RPC 调用）
// 注意：这个函数需要访问以太坊节点
//
// 参数：
//   - client: ethclient.Client 实例
//   - address: 要检查的地址
//
// 返回：
//   - true: 是合约地址
//   - false: 不是合约地址（EOA 或不存在）
//
// 示例：
//   isContract := isContractAddress(client, "0x1234...")
