# NFT Marketplace Security: How EasySwap Prevents Signature & Order Attacks

## Overview

This document explains how the EasySwap NFT marketplace prevents signature-related attacks and ensures secure order management.

---

## Key Finding: On-Chain Order Book Model

Unlike OpenSea/Blur which use off-chain signatures, **EasySwap uses a fully on-chain order book model**. This fundamentally eliminates many signature-related attack vectors.

### Comparison: Off-Chain vs On-Chain Order Model

| Aspect | OpenSea/Blur (Off-Chain) | EasySwap (On-Chain) |
|--------|--------------------------|---------------------|
| **Order Creation** | Sign message off-chain | Call `makeOrders()` on contract |
| **Signature Storage** | Stored in backend database | No signatures needed! |
| **Order Validity** | Verified at match time | Verified at creation time |
| **Signature Attacks** | Vulnerable | Not applicable |
| **Gas Cost** | Pay only at execution | Pay at creation + execution |

---

## Protection Mechanisms

### 1. Maker Verification

**Location**: `EasySwapOrderBook.sol` Line 312

```solidity
function _makeOrderTry(LibOrder.Order calldata order, ...) {
    if (
        order.maker == _msgSender() &&  // ← Only YOU can create YOUR order!
        ...
    ) {
```

**Attack Prevented**: Attacker cannot create orders on behalf of others.

---

### 2. Order Hash (OrderKey) as Unique ID

**Location**: `LibOrder.sol` Lines 83-98

```solidity
function hash(Order memory order) internal pure returns (OrderKey) {
    return OrderKey.wrap(
        keccak256(
            abi.encodePacked(
                ORDER_TYPEHASH,
                order.side,        // List or Bid
                order.saleKind,    // Collection or Item
                order.maker,       // Who created
                hash(order.nft),   // Which NFT
                Price.unwrap(order.price),  // At what price
                order.expiry,      // Until when
                order.salt         // Random nonce
            )
        )
    );
}
```

**Attack Prevented**: Replay attacks - same order params + different salt = different hash

---

### 3. Salt (Nonce) Requirement

```solidity
order.salt != 0 &&  // salt cannot be zero
```

**Purpose of Salt**:
- Ensures each order has a unique hash
- Prevents accidental duplicate orders
- Acts like a nonce for order uniqueness

---

### 4. Filled Amount Tracking

```solidity
filledAmount[LibOrder.hash(order)] == 0  // order cannot be canceled or filled
```

```solidity
// Cancel sets to MAX value
function _cancelOrder(OrderKey orderKey) internal {
    filledAmount[orderKey] = CANCELLED;  // = type(uint256).max
}
```

**Attack Prevented**: Double-fill attacks - once filled or cancelled, order hash is marked.

---

### 5. Expiry Time Validation

```solidity
(order.expiry > block.timestamp || order.expiry == 0) &&
```

**Attack Prevented**: Old/stale order execution.

---

### 6. Asset Custody at Order Creation

```solidity
// When creating a List order - NFT is deposited immediately
IEasySwapVault(_vault).depositNFT(
    newOrderKey,
    order.maker,
    order.nft.collection,
    order.nft.tokenId
);

// When creating a Bid order - ETH is deposited immediately
IEasySwapVault(_vault).depositETH{value: uint256(ETHAmount)}(
    newOrderKey,
    ETHAmount
);
```

**Attack Prevented**: 
- Seller listing NFT they don't own → Transfer fails
- Buyer placing bid without funds → Deposit fails

---

## Order Lifecycle Security Flow

```
  User creates order
        │
        ▼
┌───────────────────────────────────────────────────────────────────┐
│  CHECK 1: order.maker == msg.sender                               │
│           "You can only create orders for yourself"               │
│           ❌ Attacker cannot fake order.maker field               │
└───────────────────────────────────────────────────────────────────┘
        │
        ▼
┌───────────────────────────────────────────────────────────────────┐
│  CHECK 2: order.salt != 0                                         │
│           "Unique identifier for each order"                      │
│           ❌ Prevents duplicate order hashes                       │
└───────────────────────────────────────────────────────────────────┘
        │
        ▼
┌───────────────────────────────────────────────────────────────────┐
│  CHECK 3: filledAmount[orderHash] == 0                            │
│           "Order not already used"                                │
│           ❌ Prevents replay of cancelled/filled orders            │
└───────────────────────────────────────────────────────────────────┘
        │
        ▼
┌───────────────────────────────────────────────────────────────────┐
│  CHECK 4: expiry > block.timestamp                                │
│           "Order not expired"                                     │
│           ❌ Prevents stale order execution                        │
└───────────────────────────────────────────────────────────────────┘
        │
        ▼
┌───────────────────────────────────────────────────────────────────┐
│  ACTION: Deposit assets to Vault                                  │
│           List → Transfer NFT to Vault                            │
│           Bid  → Transfer ETH to Vault                            │
│           ❌ Prevents orders without backing assets                │
└───────────────────────────────────────────────────────────────────┘
        │
        ▼
    ✅ Order Created Successfully
```

---

## Trade-offs: On-Chain vs Off-Chain

| Off-Chain Signatures (OpenSea) | On-Chain Orders (EasySwap) |
|--------------------------------|---------------------------|
| ❌ Complex signature validation | ✅ Simple msg.sender check |
| ❌ Replay attack risks | ✅ filledAmount tracking |
| ❌ Signature front-running | ✅ No signatures to steal |
| ❌ Complex revocation | ✅ Simple cancelOrder() |
| ✅ Gas-free order creation | ❌ Pay gas for every order |
| ✅ Scalable (off-chain storage) | ❌ More on-chain storage |

---

## Interview Summary

> "EasySwap uses a **fully on-chain order book model**, which eliminates the need for cryptographic signatures. 
>
> When a user creates an order, we verify:
> 1. `order.maker == msg.sender` - only you can create your orders
> 2. Unique salt prevents duplicate order hashes
> 3. `filledAmount` mapping tracks if an order was already used (prevents replay)
> 4. Assets are deposited to the Vault immediately, preventing unfunded orders
>
> Unlike OpenSea's off-chain signature model where you need to verify EIP-712 signatures and handle revocation, our on-chain model uses msg.sender for authentication. The tradeoff is higher gas costs for order creation, but we gain simpler security model and eliminate signature-related attack vectors like replay attacks and signature front-running."

---

## Note: EIP-712 Setup

The code imports `EIP712Upgradeable` but it appears to be set up for **potential future use** or **smart contract wallet validation (EIP-1271)**:

```solidity
bytes4 private constant EIP_1271_MAGIC_VALUE = 0x1626ba7e;
```

EIP-1271 allows smart contract wallets (like Gnosis Safe) to sign messages. But the current order flow doesn't require user signatures.

---

## Front-Running (Pre-Emptive Attack) Prevention

### What is Front-Running?

Front-running occurs when an attacker sees a pending transaction in the mempool and executes their own transaction first (with higher gas) to profit.

```
Normal Flow:
1. Alice sees NFT listed at 10 ETH
2. Alice sends tx to buy it
3. Alice gets the NFT ✅

Front-Running Attack:
1. Alice sees NFT listed at 10 ETH
2. Alice sends tx to buy it (visible in mempool)
3. 🔴 Bot sees Alice's tx, sends higher gas tx to buy first
4. Bot gets the NFT
5. Alice's tx fails ❌
```

### Protection Mechanisms

#### 1. Order Maker Verification (Most Important!)

**Location**: `EasySwapOrderBook.sol` `_matchOrder()` function

```solidity
function _matchOrder(...) {
    if (_msgSender() == sellOrder.maker) {
        // Seller accepting a bid
        ...
    } else if (_msgSender() == buyOrder.maker) {
        // Buyer accepting a listing
        ...
    } else {
        revert("HD: sender invalid");  // ← CRITICAL CHECK!
    }
}
```

**How it Works**: Each order has a `maker` field. When executing a match, `msg.sender` must equal the order's maker. Attackers cannot use someone else's order - they must create their own with their own identity.

#### 2. Asset Pre-Custody in Vault

```solidity
// When order is CREATED (not matched):
IEasySwapVault(_vault).depositNFT(orderKey, maker, collection, tokenId);
```

The NFT is **already in the Vault** before matching. There's no way for a front-runner to:
- Steal the NFT before the match
- Cancel the seller's listing (only maker can cancel)

#### 3. Filled Amount Atomicity

```solidity
require(
    filledAmount[sellOrderKey] < sellOrder.nft.amount &&
    filledAmount[buyOrderKey] < buyOrder.nft.amount,
    "HD: order closed"
);

// After successful match:
_updateFilledAmount(sellOrder.nft.amount, sellOrderKey);  // Mark as filled
```

Once filled, order can't be filled again. Two competing buyers - first one wins, second one's tx reverts with refund.

#### 4. Same Maker Check

```solidity
require(sellOrder.maker != buyOrder.maker, "HD: same maker");
```

Prevents wash trading and self-dealing attacks.

### What Front-Running CAN Still Happen

**"Racing to Buy"** is unavoidable but not harmful:

```
1. Seller lists rare NFT at 10 ETH (underpriced!)
2. Alice sees it, sends buy tx
3. Bob also sees it, sends buy tx with higher gas
4. Bob's tx executes first → Bob gets NFT
5. Alice's tx reverts → Alice gets refund (no loss!)

This is just "competition" - not an exploit.
Alice lost the race but didn't lose money.
```

### Comparison: Front-Running Risk

| Attack Type | Off-Chain Signatures (OpenSea) | On-Chain Orders (EasySwap) |
|-------------|--------------------------------|---------------------------|
| **Steal pending tx** | ⚠️ Possible (copy signature) | ❌ Impossible (maker check) |
| **Race to buy** | ⚠️ Yes | ⚠️ Yes (unavoidable) |
| **Sandwich attack** | ⚠️ Possible | ❌ Not applicable (fixed price) |
| **Signature replay** | ⚠️ Risk if not properly handled | ❌ No signatures used |

### Additional Protections (Industry Solutions)

If further protection is needed:

| Method | Description |
|--------|-------------|
| **Private Mempool** | Submit txs through Flashbots (not visible to bots) |
| **Commit-Reveal** | Two-phase buying: commit hash → reveal buy |
| **Time-locked Orders** | Order only valid after X blocks |
| **Whitelist Only** | Allow only certain addresses to match |

### Interview Summary for Front-Running

> "EasySwap prevents front-running through its on-chain order model. Each order has a `maker` field that must match `msg.sender` when executing trades. This means attackers cannot 'steal' someone else's pending transaction - they must create their own order with their own identity.
>
> Additionally, assets are pre-deposited in a Vault when orders are created, so there's no way to front-run the asset transfer itself.
>
> The one thing we can't prevent is 'racing to buy' - if two people want the same NFT, the faster (higher gas) transaction wins. But this is normal market competition, not a security vulnerability. The losing buyer gets their ETH refunded automatically."

---

## Approval Abuse & NFT Theft Prevention

### The Problem: Approval Abuse in Traditional Marketplaces

Traditional marketplaces like OpenSea use `setApprovalForAll`, which creates security risks:

```
OpenSea Model:
1. User approves Seaport for ALL their NFTs
   setApprovalForAll(Seaport, true)
   └── "Seaport can move ANY of my NFTs anytime"

2. User signs order off-chain

3. If signature is leaked/stolen:
   🔴 Attacker can execute the order
   🔴 Seaport moves NFT (user still has approval set!)
   🔴 User loses NFT

Risk: Approval stays active until explicitly revoked!
```

### EasySwap's Solution: Vault Custody Model

**Key Difference: NFT is Transferred, Not Just Approved**

```
EasySwap Model:
1. User creates listing:
   makeOrders([{side: List, tokenId: 123, price: 10 ETH}])
   └── NFT #123 is TRANSFERRED to Vault contract
   └── User no longer holds the NFT (but can cancel to get back)

2. Order exists on-chain with NFT in Vault custody

3. Result:
   ✅ No standing approval needed
   ✅ Only OrderBook can move NFT from Vault
   ✅ Only maker can cancel and retrieve NFT
```

### Protection Mechanisms

#### 1. Vault Access Control

**Location**: `EasySwapVault.sol`

```solidity
modifier onlyEasySwapOrderBook() {
    require(msg.sender == orderBook, "HV: only EasySwap OrderBook");
    _;
}

function withdrawNFT(...) external onlyEasySwapOrderBook {
    // Only OrderBook can call this!
}
```

Only the OrderBook contract can move NFTs in/out of Vault. Even the Vault owner cannot steal NFTs.

#### 2. NFT is Actually Transferred, Not Approved

```solidity
function depositNFT(
    OrderKey orderKey,
    address from,
    address collection,
    uint256 tokenId
) external onlyEasySwapOrderBook {
    // Actually TRANSFER the NFT to Vault
    IERC721(collection).safeTransferNFT(from, address(this), tokenId);
    
    NFTBalance[orderKey] = tokenId;  // Track which order owns it
}
```

User must approve only the **Vault** (not a general marketplace contract), and the NFT is immediately transferred when listing. No lingering approvals.

#### 3. Order-Locked NFT Tracking

```solidity
mapping(OrderKey => uint256) public NFTBalance;

function withdrawNFT(...) external onlyEasySwapOrderBook {
    require(NFTBalance[orderKey] == tokenId, "HV: not match tokenId");
    delete NFTBalance[orderKey];
    
    IERC721(collection).safeTransferNFT(address(this), to, tokenId);
}
```

Each NFT is linked to a specific orderKey. You can't withdraw an NFT unless you have the matching order.

#### 4. Only Maker Can Cancel

```solidity
function _cancelOrderTry(OrderKey orderKey) internal returns (bool success) {
    LibOrder.Order memory order = orders[orderKey].order;

    if (
        order.maker == _msgSender() &&  // ← Only maker can cancel!
        filledAmount[orderKey] < order.nft.amount
    ) {
        // Withdraw NFT back to maker
        IEasySwapVault(_vault).withdrawNFT(
            orderHash,
            order.maker,  // ← Returns to original owner
            order.nft.collection,
            order.nft.tokenId
        );
    }
}
```

Only the original order maker can cancel and retrieve their NFT.

### Comparison: Approval Models

| Aspect | OpenSea (setApprovalForAll) | EasySwap (Vault Custody) |
|--------|----------------------------|--------------------------|
| **Approval Scope** | All NFTs in collection | Only specific NFT being listed |
| **NFT Location** | User's wallet | Vault contract |
| **Approval Duration** | Until manually revoked | Only during listing |
| **Attack Surface** | Signature leak = NFT theft | No signature = no theft vector |
| **Revocation** | Manual (user must remember) | Automatic (cancel order) |

### NFT Protection Flow

```
USER LISTS NFT:
                                                    
  User Wallet                  Vault                  OrderBook
  ┌─────────┐                ┌─────────┐             ┌─────────┐
  │ NFT #123│ ──transfer──►  │ NFT #123│ ◄──tracks── │Order Key│
  │         │                │ locked  │             │ maker   │
  └─────────┘                └─────────┘             └─────────┘
                              ▲
                              │ Only OrderBook
                              │ can move NFT

ATTACKER TRIES TO STEAL:

  ❌ Direct call to Vault.withdrawNFT()
     → Fails: "only EasySwap OrderBook"
     
  ❌ Call OrderBook.cancelOrders() 
     → Fails: "maker != msg.sender"
     
  ❌ Signature replay (like OpenSea)
     → Not applicable: No signatures used!

ONLY VALID WAYS TO MOVE NFT:

  ✅ Maker calls cancelOrders() → NFT returns to maker
  ✅ Buyer matches order → NFT goes to buyer, ETH to seller
```

### Custody vs Approval Risk Levels

| Model | Risk Level | How It Works |
|-------|------------|--------------|
| **setApprovalForAll** | 🔴 High | Contract CAN move your NFTs anytime |
| **approve(tokenId)** | 🟡 Medium | Contract CAN move specific NFT anytime |
| **Vault Custody (EasySwap)** | 🟢 Low | NFT is moved ONCE when listing, controlled by order logic |

### Interview Summary for Approval Abuse Prevention

> "EasySwap uses a **Vault Custody model** instead of the traditional approval-based model. 
>
> When a user lists an NFT, it's immediately **transferred** to the Vault contract - not just approved. The Vault has strict access controls where only the OrderBook can move NFTs, and each NFT is locked to a specific order key.
>
> This eliminates several attack vectors:
> 1. **No standing approvals** - NFT is transferred, not approved
> 2. **No signature-based attacks** - Uses msg.sender, not signatures
> 3. **Only maker can cancel** - Attacker cannot call cancel to steal
> 4. **Vault is single-purpose** - Only OrderBook can interact with it
>
> The tradeoff is that users must transfer their NFT when listing (gas cost), but they gain stronger security guarantees."

---

## Related Files

- `contracts/EasySwapOrderBook.sol` - Main order book logic
- `contracts/OrderValidator.sol` - Order validation and filled amount tracking
- `contracts/libraries/LibOrder.sol` - Order struct and hash functions
- `contracts/EasySwapVault.sol` - Asset custody
