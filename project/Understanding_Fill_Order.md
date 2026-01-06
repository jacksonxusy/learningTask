# Understanding "Fill Order" in NFT Marketplaces

## What is "Fill Order"?

**"Fill Order" = "Complete the Trade" = "Buy the NFT"**

Think of it like a restaurant order:
1. **Create Order:** "I want a burger" (written down but not cooked yet)
2. **Fill Order:** Chef cooks the burger and gives it to you (order completed)

---

## Concrete Example: Alice Sells to Bob

### Part 1: Create Order (Listing)

**Alice's Actions:**
```
1. Alice owns CoolCat #42
2. Alice creates a "sell order" (listing) at 0.5 ETH
3. NFT is locked in Vault (escrow)
4. Order is stored on-chain

Status: Order exists but NOT filled yet
```

**On-chain state after listing:**
```solidity
// Order exists
orders[0x9f8e...] = Order {
    maker: Alice,
    price: 0.5 ETH,
    nft: CoolCat #42
}

// NFT in escrow
Vault.NFTBalance[0x9f8e...] = 42  // Vault holds the NFT
CoolCat.ownerOf(42) = Vault       // Vault owns it temporarily
```

**Database state:**
```sql
ob_order_localhost:
  order_id: 0x9f8e...
  maker: Alice
  price: 0.5 ETH
  status: ACTIVE (not filled yet)  ← Order exists but not completed
```

---

### Part 2: Fill Order (Buying)

**Bob's Actions:**
```
1. Bob sees Alice's listing
2. Bob sends 0.5 ETH to buy the NFT
3. Smart contract "fills" Alice's order:
   - Takes 0.5 ETH from Bob
   - Gives 0.4875 ETH to Alice (minus 2.5% fee)
   - Gives NFT from Vault to Bob

Status: Order is now FILLED (completed)
```

**On-chain state after filling:**
```solidity
// Order marked as filled
filledAmount[0x9f8e...] = 1  // This order was filled

// NFT transferred out of Vault
Vault.NFTBalance[0x9f8e...] = 0     // Vault no longer holds it
CoolCat.ownerOf(42) = Bob           // Bob owns it now

// Alice and Bob balances updated
Alice.balance += 0.4875 ETH
Bob.NFTBalance[CoolCat #42] = true
```

**Database state:**
```sql
ob_order_localhost:
  order_id: 0x9f8e...
  maker: Alice
  taker: Bob                    ← Bob filled the order
  price: 0.5 ETH
  status: FILLED                ← Order is complete!
  quantity_remaining: 0         ← Nothing left to buy

ob_activity_localhost:
  activity_type: Sale           ← New record showing the sale
  maker: Bob
  taker: Alice
  price: 0.5 ETH
```

---

## Visual Timeline

```
Time 1: Create Order (Alice Lists)
┌─────────────────────────────────────────────┐
│ Alice:  Owns CoolCat #42                   │
│         Creates sell order at 0.5 ETH       │
│                                             │
│ Vault:  Holds CoolCat #42 (escrow)         │
│                                             │
│ Order:  ACTIVE (not filled)                │
│         Waiting for buyer...                │
└─────────────────────────────────────────────┘

⏰ Time passes...

Time 2: Fill Order (Bob Buys)
┌─────────────────────────────────────────────┐
│ Bob:    Sends 0.5 ETH                      │
│         Receives CoolCat #42                │
│                                             │
│ Alice:  Receives 0.4875 ETH                │
│         No longer owns NFT                  │
│                                             │
│ Vault:  Empty (NFT transferred out)        │
│                                             │
│ Order:  FILLED (completed)                 │
│         Trade finished!                     │
└─────────────────────────────────────────────┘
```

---

## Two Types of "Fill"

### Type 1: Direct Fill (Off-Chain Orderbook)

**Example: OpenSea**

```javascript
// Alice creates order (off-chain - just signature)
const aliceOrder = {
    maker: Alice,
    nft: CoolCat #42,
    price: 0.5 ETH,
    signature: "0xabcdef..."  // Alice's signature
}

// Stored in OpenSea database, NO blockchain transaction yet

// Bob "fills" the order (on-chain transaction)
await seaport.fulfillOrder(
    aliceOrder,      // Alice's signed order
    { value: 0.5 ETH }
)
// NOW the blockchain transaction happens
// NFT transferred Alice → Bob
// Order is "filled" (completed)
```

**Flow:**
```
Create: Off-chain (free, instant)
Fill:   On-chain (costs gas, completes trade)
```

---

### Type 2: Match Orders (On-Chain Orderbook)

**Example: This Project**

```javascript
// Alice creates order (on-chain transaction)
await orderBook.makeOrders([{
    side: List,
    maker: Alice,
    nft: CoolCat #42,
    price: 0.5 ETH
}])
// Costs gas, NFT escrowed in Vault
// Order stored on-chain

// Bob "fills" by matching the order (on-chain transaction)
await orderBook.matchOrders([{
    sellOrder: aliceOrder,   // Alice's on-chain order
    buyOrder: bobBuyOrder    // Bob's temporary buy order
}], { value: 0.5 ETH })
// NFT transferred Vault → Bob
// Order is "filled" (completed)
```

**Flow:**
```
Create: On-chain (costs gas, NFT escrowed)
Fill:   On-chain (costs gas, completes trade via matching)
```

---

## Key Terminology

| Term | Meaning | Status |
|------|---------|--------|
| **Create Order** | List NFT for sale | Order exists, waiting |
| **Fill Order** | Complete the trade | Order finished |
| **Match Order** | Same as "fill" | Order finished |
| **Active Order** | Not filled yet | Can still be bought |
| **Filled Order** | Trade completed | Cannot be bought anymore |
| **Cancelled Order** | Seller removed it | Cannot be bought anymore |

---

## Real-World Analogy

Think of it like **eBay**:

1. **Create Order = List item**
   - You post "iPhone for $500"
   - iPhone is now "for sale" (but still in your possession)
   - Status: **Active listing**

2. **Fill Order = Someone buys it**
   - Buyer pays $500
   - You ship iPhone to buyer
   - Status: **Sold** (listing completed)

**In this project:**
- **Create Order** = `makeOrders()` - Alice lists NFT
- **Fill Order** = `matchOrders()` - Bob buys NFT
- **Escrow** = Vault - Holds NFT while listed (unlike eBay where you keep it)

---

## Complete Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    FULL ORDER LIFECYCLE                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. CREATE ORDER (makeOrders)                               │
│  ┌─────────┐                                                │
│  │  Alice  │  "I want to sell CoolCat #42 for 0.5 ETH"     │
│  └────┬────┘                                                │
│       │                                                      │
│       │ Transaction: makeOrders()                           │
│       │                                                      │
│       ▼                                                      │
│  ┌──────────────┐                                           │
│  │   Contract   │  Order Status: ACTIVE                     │
│  │              │  NFT Location: Vault (escrowed)           │
│  │              │  Waiting for buyer...                     │
│  └──────────────┘                                           │
│                                                              │
│  ═══════════════════════════════════════════════════════    │
│                                                              │
│  2. FILL ORDER (matchOrders)                                │
│  ┌─────────┐                                                │
│  │   Bob   │  "I want to buy CoolCat #42 for 0.5 ETH"      │
│  └────┬────┘                                                │
│       │                                                      │
│       │ Transaction: matchOrders() + 0.5 ETH               │
│       │                                                      │
│       ▼                                                      │
│  ┌──────────────┐                                           │
│  │   Contract   │  1. Validates match                       │
│  │              │  2. ETH: Bob → Alice (0.4875)            │
│  │              │  3. NFT: Vault → Bob                      │
│  │              │  4. Protocol fee: 0.0125 ETH             │
│  └──────────────┘                                           │
│       │                                                      │
│       ▼                                                      │
│  ┌──────────────┐                                           │
│  │   Result     │  Order Status: FILLED                     │
│  │              │  NFT Owner: Bob                           │
│  │              │  Alice Balance: +0.4875 ETH               │
│  └──────────────┘                                           │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Checking Order Status

```solidity
// Check if order is filled
uint256 filled = filledAmount[orderKey];

if (filled == 0) {
    // Order is ACTIVE (not filled)
    // Can still buy it!
} else {
    // Order is FILLED
    // Already sold, cannot buy
}
```

**In database:**
```sql
-- Active orders
SELECT * FROM ob_order_localhost WHERE order_status = 0;

-- Filled orders
SELECT * FROM ob_order_localhost WHERE order_status = 1;
```

---

## Code References

### Create Order
**File:** `EasySwapOrderBook.sol`
```solidity
function makeOrders(LibOrder.Order[] calldata newOrders)
    external payable
    returns (OrderKey[] memory newOrderKeys)
{
    // Creates order
    // Escrows NFT in Vault
    // Emits LogMake event
}
```

### Fill Order
**File:** `EasySwapOrderBook.sol`
```solidity
function matchOrders(LibOrder.MatchDetail[] calldata matchDetails)
    external payable
    returns (bool[] memory successes)
{
    // Matches buy and sell orders
    // Transfers NFT and ETH
    // Updates filledAmount
    // Emits LogMatch event
}
```

### Check Filled Amount
**File:** `EasySwapOrderBook.sol`
```solidity
mapping(OrderKey => uint256) public filledAmount;

// If filledAmount[orderKey] > 0, order is filled
```

---

## Summary

**"Fill Order"** simply means **"complete the purchase"** or **"execute the trade"**!

- **Create Order:** Seller lists NFT (order waiting)
- **Fill Order:** Buyer purchases NFT (order completed)
- **Active:** Order not filled yet
- **Filled:** Order completed, trade finished

Think of it like a vending machine:
1. **Create Order:** You select "Coke, $2" (order created)
2. **Fill Order:** You insert $2, machine gives you Coke (order filled)

---

*Study Notes - NFT Marketplace Order Lifecycle*
