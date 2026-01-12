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


### token governance system
key state variables:
```

// mapping[0xAlic]= 0xBob, alice delegated her votes to bob.
mapping(address => address) public delegates;

// record numbers of user checkpoints history snapshot
mapping(address => uint32) public numCheckpoints;

// checkpoints[someAddress][0], [1], [2] record user specific time and its checkpoint snapshot
mapping(address => mapping(uint32 => Checkpoint)) public checkpoints;

struct Checkpoint {
    uint32 blockNumber;   // At which block did this happen?
    uint224 votes;        // How many votes did he have from this block onward?
}
```
key functions:  
```
function _delegate(address delegator, address delegatee) private {
    address currentDelegate = delegates[delegator];
    uint256 delegatorBalance = _balances[delegator];
    delegates[delegator] = delegatee;

    emit DelegateChanged(delegator, currentDelegate, delegatee);

    _moveDelegates(currentDelegate, delegatee, uint224(delegatorBalance));
}

function _moveDelegates(
    address from,    // old delegate (losing votes)
    address to,      // new delegate (gaining votes)
    uint224 amount   // how many votes are moving
) private {
    if (from == to) return;          // nothing to do
    if (amount == 0) return;         // nothing to do

    // ── Decrease votes from the OLD delegate ─────────────────────
    if (from != address(0)) {
        uint32 fromRepNum = numCheckpoints[from]; // number of checkpoints already written
        uint224 fromRepOld = fromRepNum > 0 
            ? checkpoints[from][fromRepNum - 1].votes 
            : 0;  // get the last snapshot checkpoint votes
        uint224 fromRepNew = fromRepOld - amount; // produce the new votes according to last one.

        _writeCheckpoint(from, fromRepNum, fromRepOld, fromRepNew);
    }

    // ── Increase votes for the NEW delegate ──────────────────────
    if (to != address(0)) {
        uint32 toRepNum = numCheckpoints[to];
        uint224 toRepOld = toRepNum > 0 
            ? checkpoints[to][toRepNum - 1].votes 
            : 0;
        uint224 toRepNew = toRepOld + amount;

        _writeCheckpoint(to, toRepNum, toRepOld, toRepNew);
    }
}
```
**note: when token move(transfer,tax,burn, etc). this function moves the voting power from the old delegate to new delegate**
transfer flow:  
```
Alice (delegated to Vitalik) 
    → transfers 1M FLOKI to → 
Bob   (delegated to CZ)

→ _transfer() calls:
  _moveDelegates(
      from = Vitalik's address,
      to   = CZ's address,
      amount = 1_000_000 * 1e9   // with decimals
  )

→ Vitalik’s latest checkpoint: votes -= 1M
→ CZ’s latest checkpoint:       votes += 1M
→ Events emitted: DelegateVotesChanged for both
```
write a new checkpoint:  
```
 function _writeCheckpoint(
        address delegatee,
        uint32 nCheckpoints,
        uint224 oldVotes,
        uint224 newVotes
    ) private {
        uint32 blockNumber = uint32(block.number); // get current blocknum. 

        if (nCheckpoints > 0 && checkpoints[delegatee][nCheckpoints - 1].blockNumber == blockNumber) {
        	// if same with last block, just update votes.
            checkpoints[delegatee][nCheckpoints - 1].votes = newVotes;
        } else {
        	// produce a new checkpoint.
            checkpoints[delegatee][nCheckpoints] = Checkpoint(blockNumber, newVotes);
            // incre checkpoint num.
            numCheckpoints[delegatee] = nCheckpoints + 1;
        }

        emit DelegateVotesChanged(delegatee, oldVotes, newVotes);
    }
```
**why need take a snapshot in the past**
when the community propose a proposal, if getting the latest votes, attacker can buy most of its token and get related votes, then 
he can votes "NO" with 60%, -- proposal fails.  
the is called last minute attack or flash loan governance attack, so using the latest votes would be cheating and completely unfair.

**Floki tax system:**  
How it works: Every time someone buys or sells FLOKI on a DEX (like Uniswap or PancakeSwap), a tiny 0.3% tax is automatically deducted.
Example: You buy $100 worth of FLOKI → you pay $100.30 ($0.30 tax goes to treasury).  
What gets taxed? Only buys and sells (not wallet-to-wallet transfers).  
Who pays? The buyer/seller — it's baked into the token code.  
Why tax? Funds the "Floki ecosystem" (games, DeFi, NFTs). Goal: Make enough revenue from products (like Valhalla game) to eventually remove the tax entirely.  


**What is the Treasury System? (The "Savings Jar")**

How it works: The tax money (0.3% of every trade) goes straight to the Floki treasury — a multi-signature wallet (needs 3+ team members to approve spends).
What's in it? Mostly FLOKI tokens, ETH/BNB, and stablecoins (USDC). As of late 2025, it's valued at ~$10–20M (fluctuates with market).
What do they spend it on?
Marketing: Ads, listings on exchanges (e.g., $10M deal with DWF Labs in 2024).
Development: Building Valhalla (NFT game), FlokiFi Locker (DeFi tool), trading bots.
Buybacks/Burns: 25% of FlokiFi fees auto-buy and burn FLOKI (reduces supply).
Charity/Community: Donations, airdrops (e.g., massive burn proposal in Feb 2024).

Transparency: Treasury addresses are public (e.g., on Etherscan), and multisig requires 3 signatures for withdrawaw.









