package misssionone

// 使用 Sepolia 测试网络实现基础的区块链交互，包括查询区块和发送交易。
//  具体任务
// 环境搭建
// 安装必要的开发工具，如 Go 语言环境、 go-ethereum 库。
// 注册 Infura 账户，获取 Sepolia 测试网络的 API Key。
// 查询区块
// 编写 Go 代码，使用 ethclient 连接到 Sepolia 测试网络。
// 实现查询指定区块号的区块信息，包括区块的哈希、时间戳、交易数量等。
// 输出查询结果到控制台。
// 发送交易
// 准备一个 Sepolia 测试网络的以太坊账户，并获取其私钥。
// 编写 Go 代码，使用 ethclient 连接到 Sepolia 测试网络。
// 构造一笔简单的以太币转账交易，指定发送方、接收方和转账金额。
// 对交易进行签名，并将签名后的交易发送到网络。
// 输出交易的哈希值。
import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func readAndWrite() {
	client, err := ethclient.Dial("https://eth-sepolia.g.alchemy.com/v2/AJiMn0vUa6ejkX1uhNv9h")
	if err != nil {
		log.Fatal(err)
	}

	blockNumber := big.NewInt(9716222)

	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(block.Number().Uint64())
	fmt.Println(block.Hash())
	fmt.Println(block.Time())
	fmt.Println(len(block.Transactions()))

	privateKey, err := crypto.HexToECDSA("123")
	if err != nil {
		log.Fatal(err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal(err)
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	value := big.NewInt(10000000000000000)
	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	toAddress := common.HexToAddress("0xcc393ecd46ddd30c6a1333746620011ea12f5b35")
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())

}
