package misssionone

// 使用 abigen 工具自动生成 Go 绑定代码，用于与 Sepolia 测试网络上的智能合约进行交互。
//  具体任务
// 编写智能合约
// 使用 Solidity 编写一个简单的智能合约，例如一个计数器合约。
// 编译智能合约，生成 ABI 和字节码文件。
// 使用 abigen 生成 Go 绑定代码
// 安装 abigen 工具。
// 使用 abigen 工具根据 ABI 和字节码文件生成 Go 绑定代码。
// 使用生成的 Go 绑定代码与合约交互
// 编写 Go 代码，使用生成的 Go 绑定代码连接到 Sepolia 测试网络上的智能合约。
// 调用合约的方法，例如增加计数器的值。
// 输出调用结果。

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"golangEth/misssionOne/counter"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func Deploy() {
	client, err := ethclient.Dial("https://eth-sepolia.g.alchemy.com/v2/AJiMn0vUa6ejkX1uhNv9h")
	if err != nil {
		log.Fatal(err)
	}
	privateKey, err := crypto.HexToECDSA("c223ad0da0b27f9bd377457e3633528604219269c919e990030ead1f3e926508")
	publicKey := privateKey.Public()
	publickKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publickKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		log.Fatal(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(300000)
	auth.GasPrice = gasPrice

	address, tx, _, err := counter.DeployCounter(auth, client)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Contract deployed to address:", address.Hex())
	fmt.Println("Transaction hash:", tx.Hash().Hex())

	// 	Contract deployed to address: 0x69182019d58708d371949121d24fC3fd86574D0A
	// Transaction hash: 0xb499ea697789936efcb7dafe8ddb8b3f50e31782173ef5472eb4530d5058629e

}

func Incre() {
	client, err := ethclient.Dial("https://eth-sepolia.g.alchemy.com/v2/AJiMn0vUa6ejkX1uhNv9h")
	if err != nil {
		log.Fatal(err)
	}
	counterContract, err := counter.NewCounter(common.HexToAddress("0x69182019d58708d371949121d24fC3fd86574D0A"), client)
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := crypto.HexToECDSA("c223ad0da0b27f9bd377457e3633528604219269c919e990030ead1f3e926508")
	if err != nil {
		log.Fatal(err)
	}
	opt, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(11155111))
	if err != nil {
		log.Fatal(err)
	}
	tx, err := counterContract.Increment(opt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("tx hash:", tx.Hash().Hex())

	callOpt := &bind.CallOpts{Context: context.Background()}
	value, err := counterContract.Nums(callOpt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("nums value: %s\n", value.String())

}
