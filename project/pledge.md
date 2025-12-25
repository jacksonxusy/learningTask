  depositLend (存款借贷)

  - Role: Lender (出借方)
  - Action: Deposits lendToken (usually stablecoins like USDT/USDC)
  - Purpose: Provide liquidity to earn fixed interest
  - Token: Receives spToken (share token) representing their lending position
  - Expected Return: Fixed interest rate + principal at maturity

  depositBorrow (存款质押)

  - Role: Borrower (借款方)
  - Action: Deposits borrowToken (collateral like BTC, ETH) as security
  - Purpose: Provide collateral to borrow stablecoins
  - Token: Receives jpToken (debt token) representing their borrowing position
  - Expected Action: Can borrow up to collateral limit, must maintain collateralization ratio

  💡 Key Business Flow

  Pool Creation Phase (MATCH state):
  ┌─────────────────┐    ┌─────────────────┐
  │   Lenders       │    │    Borrowers    │
  │                 │    │                 │
  │ depositLend()   │    │ depositBorrow() │
  │                 │    │                 │
  │ Get spToken     │    │ Get jpToken     │
  │ (share token)   │    │ (debt token)    │
  └─────────────────┘    └─────────────────┘
           ↓                       ↓
  ┌─────────────────────────────────────────┐
  │           PledgePool Contract          │
  │                                         │
  │ LendToken Supply ←→ BorrowToken Supply │
  │      (稳定币供给)      (质押品供给)       │


  eg:


  What Actually Happens During Liquidation

  Before Liquidation:
  ┌─────────────────┐    ┌─────────────────┐
  │   Borrower      │    │     Lender      │
  │                 │    │                 │
  │ Collateral: ETH │◄──►│   Lend: USDC    │
  │ Value: $10,000  │    │   Amount: $6,666│
  │ Borrowed: $6,666│    │ (Overcollateral)│
  │ Ratio: 150%     │    │                 │
  └─────────────────┘    └─────────────────┘

  Price drops: ETH $10,000 → $7,000
  New ratio: $7,000 ÷ $6,666 = 105% (below 150% threshold)
  ↓ TRIGGER LIQUIDATION

  Actual Liquidation Flow

  NOT what you described - here's what really happens:

  sequenceDiagram
      participant PriceOracle
      participant PledgePool
      participant Liquidator
      participant Borrower
      participant Lender

      PriceOracle->>PledgePool: ETH price: $10k → $7k
      PledgePool->>PledgePool: Check: 105% < 150% threshold
      PledgePool->>Liquidator: LIQUIDATION OPPORTUNITY!
      Liquidator->>PledgePool: Repay borrower's debt ($6,666 USDC)
      PledgePool->>Liquidator: Send collateral (1 ETH worth $7,000)
      Liquidator->>Liquidator: Profit: $7,000 - $6,666 = $334
      PledgePool->>Lender: Return your funds + interest
---------------------------------------------------------
  🛡️ Key Components Explained

  1. Multi-Signature Address Storage

  contract multiSignatureClient{
      uint256 private constant multiSignaturePositon = uint256(keccak256("org.multiSignature.storage"));

      constructor(address multiSignature) public {
          require(multiSignature != address(0),"multiSignatureClient : Multiple signature contract address is zero!");
          saveValue(multiSignaturePositon,uint256(multiSignature));  // Store the multi-sig contract address
      }
  }

  Purpose: Stores the address of the main multi-signature contract that governs this client.

  2. The validCall Modifier - Security Gateway

  modifier validCall(){
      checkMultiSignature();  // Verify multi-sig approval
      _;                     // Execute function if approved
  }

  This modifier is used on critical functions in PledgePool like:
  - setFee() - Setting lending/borrowing fees
  - setSwapRouterAddress() - Changing DEX router
  - setFeeAddress() - Changing fee recipient
  - Pool creation and management functions

  3. Multi-Signature Verification Logic

  function checkMultiSignature() internal view {
      bytes32 msgHash = keccak256(abi.encodePacked(msg.sender, address(this)));
      address multiSign = getMultiSignatureAddress();

      // Check if this call has been approved by enough signatures
      uint256 newIndex = IMultiSignature(multiSign).getValidSignature(msgHash,defaultIndex);
      require(newIndex > defaultIndex, "multiSignatureClient : This tx is not aprroved");
  }

  How it works:
  1. Hash Creation: Creates a unique hash from msg.sender (caller) + address(this) (target contract)
  2. Multi-Sig Check: Queries the multi-signature contract for approval status
  3. Approval Verification: Requires sufficient signatures (threshold) before allowing execution



  🔄 Multi-Signature Workflow

  Step 1: Creating a Proposal

  // In multiSignature.sol
  function createApplication(address to) external returns(uint256) {
      bytes32 msghash = getApplicationHash(msg.sender, to);  // Hash of caller + target
      uint256 index = signatureMap[msghash].length;
      signatureMap[msghash].push(signatureInfo(msg.sender,new address[](0)));
      emit CreateApplication(msg.sender,to,msghash);
      return index;
  }

  Step 2: Multiple Owners Sign

  function signApplication(bytes32 msghash) external onlyOwner {
      signatureMap[msghash][defaultIndex].signatures.addWhiteListAddress(msg.sender);
      // Each owner adds their signature to the proposal
  }

  Step 3: Approval Check

  function getValidSignature(bytes32 msghash,uint256 lastIndex) external view returns(uint256){
      signatureInfo[] storage info = signatureMap[msghash];
      for (uint256 i=lastIndex;i<info.length;i++){
          if(info[i].signatures.length >= threshold){  // Check if signatures >= required threshold
              return i+1;
          }
      }
      return 0;  // Not enough signatures
  }
----------------------------------------------- settle
 Example Scenario

  Pool Parameters:
  - borrowSupply: 50 ETH (total collateral deposited)
  - martgageRate: 150000000 (150% collateralization ratio)
  - lendToken: USDC (stablecoin)
  - borrowToken: ETH

  Oracle Prices:
  - prices[0]: 1,000,000 (USDC/USD = $1.00)
  - prices[1]: 3,000,000,000 (ETH/USD = $3,000)

  Constants:
  - calDecimal: 1e18 (1,000,000,000,000,000,000)
  - baseDecimal: 1e8 (100,000,000)

  📊 Line 1: Calculate Total Collateral Value

  uint256 totalValue = pool.borrowSupply.mul(prices[1].mul(calDecimal).div(prices[0])).div(calDecimal);

  Step-by-Step Calculation:

  1. prices[1].mul(calDecimal) - Normalize ETH price
  3,000,000,000 × 1,000,000,000,000,000,000 = 3e27
  2. .div(prices[0]) - Divide by USDC price to get ETH/USDC ratio
  3e27 ÷ 1,000,000 = 3e21
  2. This gives us: 1 ETH = 3,000 USDC
  3. pool.borrowSupply.mul(...) - Multiply by ETH amount
  50 × 3e21 = 150e21
  4. .div(calDecimal) - Remove the decimal scaling
  150e21 ÷ 1e18 = 150,000

  Result: totalValue = 150,000 (Total collateral in USDC terms: 50 ETH × $3,000 = $150,000)

  🛡️ Line 2: Calculate Effective Lending Value

  uint256 actualValue = totalValue.mul(baseDecimal).div(pool.martgageRate);

  Step-by-Step Calculation:

  1. totalValue.mul(baseDecimal) - Scale for precision
  150,000 × 100,000,000 = 15e12
  2. .div(pool.martgageRate) - Apply collateralization ratio
  15e12 ÷ 150,000,000 = 100,000

  Result: actualValue = 100,000 (Amount that can be safely lent out)



  💰 Settlement Logic Explained

  Scenario 1: Borrowers' Collateral < Lenders' Supply

  if (pool.lendSupply > actualValue) {
      // More people want to lend than can be supported by collateral
      data.settleAmountLend = actualValue;     // Limited by collateral value
      data.settleAmountBorrow = pool.borrowSupply; // All collateral used
  }

  Example:
  - Lenders deposited: $100,000 USDC
  - Borrowers deposited: 1 ETH (worth $80,000)
  - Collateralization ratio: 150%
  - Effective lending capacity: $80,000 ÷ 1.5 = $53,333
  - Settlement: Only $53,333 of the $100,000 can be deployed

  Scenario 2: Borrowers' Collateral ≥ Lenders' Supply

  else {
      // Enough collateral to support all lenders
      data.settleAmountLend = pool.lendSupply;
      data.settleAmountBorrow = pool.lendSupply.mul(pool.martgageRate).div(prices[1].mul(baseDecimal).div(prices[0]));
  }

  Example:
  - Lenders deposited: $50,000 USDC
  - Borrowers deposited: 2 ETH (worth $160,000)
  - Collateralization ratio: 150%
  - Required collateral: $50,000 × 1.5 = $75,000 worth of ETH
  - Settlement: All $50,000 can be deployed, using $75,000/$160,000 = 46.875% of collateral


---------------------------------------------
claim Lend

  Pool Scenario

  Pool Data:
  - Total Lend Supply: 100,000 USDC
  - Settle Amount Lend: 100,000 USDC
  - Total spTokens to mint: 100,000 spUSDC

  User Data:
  - Alice deposited: 10,000 USDC
  - Bob deposited: 5,000 USDC
  - Charlie deposited: 85,000 USDC

  Alice's Calculation

  // 1. User Share Calculation
  userShare = 10,000 × 1e18 ÷ 100,000 = 0.1 × 1e18 = 100,000,000,000,000,000

  // 2. spToken Calculation
  spAmount = 100,000 × 100,000,000,000,000,000 ÷ 1e18 = 10,000 spUSDC

  // Result: Alice receives 10,000 spUSDC (10% of total)

  Bob's Calculation

  userShare = 5,000 × 1e18 ÷ 100,000 = 0.05 × 1e18 = 50,000,000,000,000,000
  spAmount = 100,000 × 50,000,000,000,000,000 ÷ 1e18 = 5,000 spUSDC

  // Result: Bob receives 5,000 spUSDC (5% of total)

----------------------------
finish

  Pool Parameters

  Pool Data:
  - settleTime: 1640995200    // Jan 1, 2022
  - endTime: 1643587200        // Jan 31, 2022  
  - interestRate: 5000000    // 5% annual
  - lendFee: 2000000          // 2% lending fee
  - borrowFee: 1000000        // 1% borrow fee
  - settleAmountLend: 100000 USDC
  - settleAmountBorrow: 80 ETH

  Calculations

  1. Time Ratio
  timeRatio = ((1643587200 - 1640995200) × 1e8) ÷ (365 × 86400)
           = (2592000 × 1e8) ÷ 31536000
           = 0.08219 × 1e8 = 8,219,178

  2. Interest Calculation
  interest = 8,219,178 × (5,000,000 × 100,000) ÷ 1e16
           = 8,219,178 × 500,000,000,000 ÷ 1e16
           = 8,219,178 × 0.05 = 410,959

  3. Total Amount Owed
  lendAmount = 100,000 + 410,959 = 100,410 USDC

  4. Sell Amount
  sellAmount = 100,410 × (1e8 + 2,000,000) ÷ 1e8
             = 100,410 × 1.02 = 102,418 USDC worth of ETH

  5. After Swap
  Assume swap returns: 102,500 USDC (better than expected)
  FeeAmount = 102,500 - 100,410 = 2,090 USDC
  LenderAmount = 100,410 USDC
  TreasuryFee = 2,090 USDC

  6. Borrower Collateral Return
  Remaining ETH = 80 - 34.07 = 45.93 ETH (amount not sold)
  BorrowFee = 45.93 × 1% = 0.459 ETH
  BorrowerReturn = 45.93 - 0.459 = 45.471 ETH

------------------
CLIENT                                    SERVER
   │                                         │
   │── Upgrade to WebSocket ────────────────▶│
   │◀── Connection established ─────────────│
   │                                         │
   │          (price changes)                │
   │◀── {"price": "1.234"} ─────────────────│  Server pushes instantly
   │                                         │
   │          (price changes)                │
   │◀── {"price": "1.235"} ─────────────────│  Server pushes instantly
   │                                         │
   │          (no change for 5 seconds)      │
   │          (nothing sent - efficient!)    │
   │                                         │

Aspect  HTTP Polling  WebSocket
Latency Depends on poll interval (e.g., 1s delay) Instant (sub-millisecond)
Efficiency  ❌ Many wasted requests when price unchanged ✅ Only sends when data changes
Server Load ❌ High (handles N requests/sec × clients) ✅ Low (one connection per client)
Complexity  ✅ Simple to implement ❌ More complex (heartbeat, reconnection)
Firewall/Proxy  ✅ Always works  ⚠️ Some proxies block WebSocket
Connection Overhead ❌ TCP handshake + HTTP headers every request  ✅ One-time handshake


---------------------------------
price controller 
┌──────────────────────────────────────────────────────────────────────────────────┐
│                              TIMELINE                                             │
├──────────────────────────────────────────────────────────────────────────────────┤
│                                                                                   │
│   T=0s │ Client connects                                                          │
│        │ ReadAndWrite() starts                                                    │
│        │ Manager.Servers.Store(id, server)                                        │
│        │ 3 goroutines spawn                                                       │
│        │                                                                          │
│   T=1s │ Heartbeat check: LastTime=0s, now=1s → OK                               │
│        │                                                                          │
│   T=2s │ Client sends "ping"                                                      │
│        │ Read goroutine: LastTime = 2s                                            │
│        │ Write goroutine: sends "pong"                                            │
│        │                                                                          │
│   T=3s │ Heartbeat check: LastTime=2s, now=3s → OK                               │
│        │                                                                          │
│   T=5s │ KuCoin price update: "0.0027"                                           │
│        │ StartServer() → s.Send ← "0.0027"                                        │
│        │ Write goroutine: sends price to client                                   │
│        │                                                                          │
│  T=32s │ No ping for 30 seconds (UserPingPongDurTime)                            │
│        │ Heartbeat check: TIMEOUT!                                                │
│        │ Send "heartbeat timeout" to client                                       │
│        │ return → defer → cleanup                                                 │
│        │                                                                          │
└──────────────────────────────────────────────────────────────────────────────────┘


┌──────────────────────────────────────────────────────────────────────────────────┐
│                           PLEDGE BACKEND SYSTEM                                   │
│                                                                                   │
│   ┌─────────────────────┐              ┌─────────────────────┐                   │
│   │   API Server        │              │   Task Scheduler    │ ◄── task.go       │
│   │   (pledge-backend)  │              │   (pledge_task.go)  │                   │
│   │                     │              │                     │                   │
│   │   - HTTP endpoints  │              │   - Pool sync       │                   │
│   │   - WebSocket       │              │   - Price updates   │                   │
│   │   - User auth       │              │   - Balance monitor │                   │
│   └──────────┬──────────┘              └──────────┬──────────┘                   │
│              │                                    │                              │
│              │              ┌─────────────┐       │                              │
│              └─────────────▶│    Redis    │◀──────┘                              │
│                             │   (Cache)   │                                      │
│                             └──────┬──────┘                                      │
│                                    │                                             │
│                             ┌──────▼──────┐                                      │
│                             │    MySQL    │                                      │
│                             │  (Storage)  │                                      │
│                             └─────────────┘                                      │
│                                                                                   │
│   ┌───────────────────────────────────────────────────────────────────────────┐  │
│   │                        External Services                                   │  │
│   │   - Blockchain RPC (Ethereum/BSC)                                          │  │
│   │   - KuCoin API (prices)                                                    │  │
│   └───────────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────────┘

// update poolService.go
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           Data Flow                                             │
│                                                                                 │
│   ┌──────────────────┐         ┌──────────────────┐         ┌───────────────┐  │
│   │  Blockchain      │         │  poolService     │         │  Database     │  │
│   │  (Smart Contract)│────────▶│  UpdatePoolInfo  │────────▶│  MySQL/Redis  │  │
│   │  PledgePoolToken │         │                  │         │               │  │
│   └──────────────────┘         └──────────────────┘         └───────────────┘  │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
store to db logic.
┌─────────────────────────────────────────────────────────────────────────┐
│                    MD5 Change Detection                                  │
│                                                                          │
│   New Data ───▶ MD5("poolData") ───▶ "abc123"                           │
│                                           │                              │
│   Redis Key: "base_info:pool_97_1"        ▼                              │
│   Redis Value: "abc123"     ─────▶ Compare ─────▶ Same? Skip DB write   │
│                                           │                              │
│   Redis Value: "xyz789"     ─────▶ Compare ─────▶ Different? Write DB   │
└─────────────────────────────────────────────────────────────────────────┘




FE get

Approach  Pros  Cons
HTTP for pool data  Simple, cacheable, RESTful  Slightly delayed
WebSocket for prices  Real-time, efficient  More complex, stateful
Pool data changes every 2 minutes → HTTP is sufficient. Price changes multiple times per second → WebSocket is necessary.




--------------
kucoin.go
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                          │
│   KuCoin Exchange                                                                        │
│        │                                                                                 │
│        │ WebSocket (real-time price updates)                                             │
│        ▼                                                                                 │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐
│   │                           kucoin.go                                                 │
│   │                                                                                      │
│   │   case msg := <-mc:                                                                  │
│   │       PlgrPriceChan <- t.Price  ─────────────┬─────────────▶ ws.go (StartServer)    │
│   │       PlgrPrice = t.Price       ─────────────┼─────────────▶ Global variable         │
│   │       db.RedisSetString(...)    ─────────────┼─────────────▶ Redis (backup only)     │
│   │                                              │                                       │
│   └──────────────────────────────────────────────┼───────────────────────────────────────┘
│                                                  │                                       │
│                                    Go Channel    │                                       │
│                                    (in memory)   │                                       │
│                                                  ▼                                       │
│   ┌─────────────────────────────────────────────────────────────────────────────────────┐
│   │                           ws.go - StartServer()                                     │
│   │                                                                                      │
│   │   for {                                                                              │
│   │       select {                                                                       │
│   │       case price := <-kucoin.PlgrPriceChan:  ◀──── Gets from CHANNEL, not Redis!   │
│   │           Manager.Servers.Range(...)                                                 │
│   │               → SendToClient(price)                                                  │
│   │       }                                                                              │
│   │   }                                                                                  │
│   │                                                                                      │
│   └─────────────────────────────────────────────────────────────────────────────────────┘
│                                                  │                                       │
│                                                  ▼                                       │
│                                         Frontend Clients                                 │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘

```
func GetExchangePrice() {
    // On startup, load cached price from Redis
    price, err := db.RedisGetString("plgr_price")
    if err == nil {
        PlgrPrice = price  // Use as initial value
    }
    
    // Then connect to KuCoin for live updates...
}
```
using this just because if the server restarts, it can immediately serve the last known price while waiting kucoin to reconnect.

full context:
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                          │
│   ┌─────────────────┐        ┌─────────────────────────────┐        ┌───────────────┐   │
│   │   Blockchain    │        │   /schedule/services/       │        │    MySQL      │   │
│   │   Smart Contract│───────▶│   poolService.go            │───────▶│    Database   │   │
│   │                 │  READ  │   (UpdatePoolInfo)          │  WRITE │               │   │
│   └─────────────────┘        └─────────────────────────────┘        └───────┬───────┘   │
│                                                                             │           │
│                                                                             │ READ      │
│                                                                             ▼           │
│   ┌─────────────────┐        ┌─────────────────────────────┐        ┌───────────────┐   │
│   │    Frontend     │◀───────│   /api/services/            │◀───────│    MySQL      │   │
│   │    (Browser)    │  HTTP  │   poolService.go            │  READ  │    Database   │   │
│   │                 │        │   (PoolBaseInfo)            │        │               │   │
│   └─────────────────┘        └─────────────────────────────┘        └───────────────┘   │
│                                                                                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
