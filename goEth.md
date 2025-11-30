### ERC20 transfer
```
	// 6. 构造 transfer(address,uint256) 的 calldata（最稳方式）
	transferFnSignature := []byte("transfer(address,uint256)")
	methodID := crypto.Keccak256Hash(transferFnSignature).Bytes()[:4]

	paddedTo := common.LeftPadBytes(toAddress.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedTo...)
	data = append(data, paddedAmount...)
```
1. transfer ERC20 token use transfer method, if using call to transfer, it has reentrancy risk. trasfer function has
a 2300 gas limit, which will block reentrancy attack, cause the most easist fallback need greater than 5000 gas.
2. `paddedTo := common.LeftPadBytes(toAddress.Bytes(), 32)` change address to 32 bytes, becuase evm read.
3. `methodID := crypto.Keccak256Hash(transferFnSignature).Bytes()[:4]` get its function selector.
memory per 32 bytes(as one slot)

```
tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(11155111), // Sepolia
		Nonce:     nonce,
		To:        &linkTokenAddress,
		Value:     big.NewInt(0),
		Gas:       gasLimit + 10000, // 加点 buffer
		GasFeeCap: feeCap,
		GasTipCap: tip,
		Data:      data,
	})

	// 10. 签名 + 发送
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(big.NewInt(11155111)), privateKey)
	if err != nil {
		log.Fatal("签名失败:", err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
```
1. go code doesn't transfer token, it just send message to token contract, and contract doing transfer.
2. ETH was stored in wallet, but LINK or other ERC20 token was stored in contract mapping.
ETH use account mode, but ERC20 token use balance mode。

```
世界状态（World State）:
┌─────────────────────────────────────┐
│ 地址: 0x123...                        │
│ balance: 5.0 ETH ← 协议直接记录       │
│ ...                                 │
└─────────────────────────────────────┘

LINK 合约存储：
┌─────────────────────────────────────┐
│ balances[0x123...] = 100 LINK         │ ← 合约自己维护的 mapping
│ balances[0xabc...] = 50 LINK          │
└─────────────────
```
3. ERC20 token will verify signature to check if this transaction was sent by its public address.


### execute contract function using eth client
```
callOpt := &bind.CallOpts{Context: context.Background()}
valueInContract, err := storeContract.Items(callOpt, key)
```
1. create a read-only settings object that tells the Eth node, it just need to read data.
2. reads data directly from the blockchain. this is off-chain tansaction, no gas, no waiting from mining.
original contract:  
```
contract Store {
  event ItemSet(bytes32 key, bytes32 value);

  string public version;
  mapping (bytes32 => bytes32) public items;

  constructor(string memory _version) {
    version = _version;
  }

  function setItem(bytes32 key, bytes32 value) external {
    items[key] = value;
    emit ItemSet(key, value);
  }
}
```
**CallOpts vs TransactOpts** — The Ultimate Comparison (2025)

| Feature                  | `CallOpts` (Read)                          | `TransactOpts` (Write)                          |
|--------------------------|---------------------------------------------|--------------------------------------------------|
| **Purpose**              | Read data from contract (`view`/`pure`)     | Change contract state (write/modify)            |
| **Needs private key?**   | No                                          | Yes                                             |
| **Costs gas?**           | No (free & instant)                         | Yes                                             |
| **Needs signature?**     | No                                          | Yes (ECDSA signature created automatically)     |
| **Wait for mining?**     | No (returns immediately)                    | Yes (must wait for block inclusion)             |
| **Speed**                | < 200ms                                     | 10–30 seconds (depends on network)              |
| **Typical functions**    | `balanceOf()`, `name()`, `symbol()`, `Items()`, `totalSupply()` | `transfer()`, `approve()`, `setItem()`, `mint()`, `upgradeTo()` |
| **Created with**         | `&bind.CallOpts{}`<br>`&bind.CallOpts{From: addr}` | `bind.NewKeyedTransactorWithChainID(pk, chainID)` |
| **Context support**      | Yes (`Context: ctx`)                        | Yes                                             |
| **Gas settings**         | Not applicable                              | `GasLimit`, `GasPrice`, `GasTipCap`, `GasFeeCap` |
| **Value (send ETH)**     | Not applicable                              | `auth.Value = big.NewInt(1e18)`                 |

```go
callOpt := &bind.CallOpts{Context: context.Background()}
value, err := instance.Items(callOpt, key)
balance, err := token.BalanceOf(callOpt, address)
name, err := token.Name(callOpt)
```


### contract execution
## Two Ways to Call Contracts in Go

| Feature                          | Method A — Easy Way (`abigen` + `bind`)                                 | Method B — Hardcore Way (Raw ABI + Manual Tx)                              |
|----------------------------------|--------------------------------------------------------------------------|-----------------------------------------------------------------------------|
| On-chain result                  | Identical — contract runs the function                                  | Identical — contract runs the function                                     |
| Difficulty                       | ★☆☆☆☆ (Beginner)                                                         | ★★★★★ (Expert)                                                             |
| Lines of code                    | 5–10 lines                                                               | 50+ lines                                                                  |
| Need Solidity + abigen?          | Yes                                                                      | No — only need ABI JSON                                                    |
| Can call any contract instantly? | No (must generate binding first)                                         | Yes — just paste ABI                                                       |
| Used by                          | Tutorials, dApps, personal projects                                      | Chainlink, MetaMask, wallets, bridges, DeFi bots                           |
| Real-world analogy               | Driving a Tesla (automatic)                                              | Building a Formula 1 car from scratch                                      |
| Best for                         | Learning, small projects, 1–10 contracts                                 | Production backends, supporting 1000+ tokens, maximum flexibility          |

### two way of signature
```
// Example 1: using bind to call contract methods
auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
auth.Value = big.NewInt(0)
auth.GasLimit = 300000
tx, err := myContract.MyMethod(auth, arg1, arg2) // bind builds & signs tx internally

// Example 2: build + sign raw tx manually
tx := types.NewTransaction(nonce, toAddr, value, gasLimit, gasPrice, data)
signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
if err != nil { ... }
err = client.SendTransaction(ctx, signedTx)
```
1. Use bind.NewKeyedTransactorWithChainID for contract interactions via generated bindings(simpler, 
handles ABI encoding, nonce/gas helpers).  
2. Use types.SignTx when you need low level control or are sending raw transactions(not using generated bindings)













