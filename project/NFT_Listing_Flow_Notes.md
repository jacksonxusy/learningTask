# NFT Listing Flow - Deep Dive Study Notes

## Overview

This document explains the complete flow of listing an NFT for sale in the EasySwap NFT Marketplace, focusing on **Smart Contract** and **Backend** implementation details.

---

## Architecture Layers

```
User → Frontend → Smart Contract (On-Chain) → Event Emission → Sync Service → Database
```

---

## Example Scenario

**Alice lists CoolCat #42 for 0.5 ETH**

| Parameter | Value |
|-----------|-------|
| Seller | `0xAlice1234567890abcdef1234567890abcdef1234` |
| NFT Collection | `0xCoolCats000000000000000000000000000000CC` |
| Token ID | `42` |
| Price | `0.5 ETH` (500000000000000000 wei) |
| Expiry | `1735084800` (Dec 25, 2024) |
| Salt | `1703419389123` (random nonce) |

---

## Phase 1: Smart Contract Execution

### Step 1: NFT Approval (Prerequisite)

Before listing, Alice must approve the **Vault Contract** to transfer her NFT:

```solidity
// Alice calls on CoolCats NFT contract:
CoolCats.approve(
    0x49D92FC524260F69dfCb4415386dD03BfE211858,  // Vault address
    42  // Token ID
);
```

**Why Vault, not OrderBook?** The Vault is the escrow contract that holds assets during active orders.

---

### Step 2: Call `makeOrders()` Function

**File:** `EasySwapOrderBook.sol` (lines 133-170)

Alice sends transaction with order data:

```solidity
Order({
    side: 0,          // List (selling)
    saleKind: 1,      // FixedPriceForItem
    maker: 0xAlice...,
    nft: Asset({
        tokenId: 42,
        collection: 0xCoolCats...,
        amount: 1
    }),
    price: 500000000000000000,  // 0.5 ETH in wei
    expiry: 1735084800,
    salt: 1703419389123
})
```

**Contract Processing:**

```solidity
function makeOrders(LibOrder.Order[] calldata newOrders) 
    external payable whenNotPaused nonReentrant 
{
    for (uint256 i = 0; i < newOrders.length; ++i) {
        OrderKey newOrderKey = _makeOrderTry(newOrders[i], buyPrice);
        newOrderKeys[i] = newOrderKey;
    }
}
```

---

### Step 3: Internal Processing `_makeOrderTry()`

**File:** `EasySwapOrderBook.sol` (lines 307-357)

#### 3.1 Validation Checks

```solidity
if (
    order.maker == _msgSender() &&              // Only maker can create
    Price.unwrap(order.price) != 0 &&           // Price > 0
    order.salt != 0 &&                          // Salt != 0
    (order.expiry > block.timestamp || order.expiry == 0) &&  // Not expired
    filledAmount[LibOrder.hash(order)] == 0     // Order doesn't exist
) {
    // Process order...
}
```

#### 3.2 Generate OrderKey (Unique ID)

**File:** `LibOrder.sol` (lines 83-99)

```solidity
function hash(Order memory order) internal pure returns (OrderKey) {
    return OrderKey.wrap(
        keccak256(
            abi.encodePacked(
                ORDER_TYPEHASH,
                order.side,
                order.saleKind,
                order.maker,
                hash(order.nft),
                Price.unwrap(order.price),
                order.expiry,
                order.salt  // ← Makes each order unique!
            )
        )
    );
}
```

**Result:** `orderKey = 0x9f8e7d6c5b4a3912abcd1234567890ef...`

**Key Insight:** The `salt` ensures uniqueness even if same NFT is listed multiple times at same price.

---

#### 3.3 Deposit NFT to Vault (Escrow)

```solidity
if (order.side == LibOrder.Side.List) {
    // Transfer NFT from Alice to Vault
    IEasySwapVault(_vault).depositNFT(
        newOrderKey,           // 0x9f8e7d6c...
        order.maker,           // 0xAlice...
        order.nft.collection,  // 0xCoolCats...
        order.nft.tokenId      // 42
    );
}
```

**Vault Contract** (`EasySwapVault.sol` lines 57-66):

```solidity
function depositNFT(
    OrderKey orderKey,
    address from,
    address collection,
    uint256 tokenId
) external onlyEasySwapOrderBook {
    // Transfer NFT: Alice → Vault
    IERC721(collection).safeTransferNFT(from, address(this), tokenId);
    
    // Record deposit
    NFTBalance[orderKey] = tokenId;
}
```

**State Changes:**
- **CoolCats Contract:** `_owners[42]` changes from `0xAlice...` to `0xVault...`
- **Vault Contract:** `NFTBalance[0x9f8e7d6c...]` = `42`

**📌 Critical Point:** Alice no longer owns the NFT! It's escrowed in the Vault.

---

#### 3.4 Add Order to Storage

```solidity
_addOrder(order);
```

This stores the order in a **Red-Black Tree** data structure, indexed by (collection, side, price) for efficient price-based querying.

**State Changes:**
```solidity
orders[0x9f8e7d6c...] = DBOrder({
    order: Order(...),
    next: 0x0000...
})
```

---

#### 3.5 Emit LogMake Event

```solidity
emit LogMake(
    newOrderKey,    // 0x9f8e7d6c...
    order.side,     // 0 (indexed)
    order.saleKind, // 1 (indexed)
    order.maker,    // 0xAlice... (indexed)
    order.nft,      // Asset struct
    order.price,    // 500000000000000000
    order.expiry,   // 1735084800
    order.salt      // 1703419389123
);
```

**Raw Event Log Structure:**

```json
{
  "address": "0xOrderBookContract...",
  "topics": [
    "0xfc37f2ff...",  // Event signature (LogMake)
    "0x00...00",      // side = 0 (List)
    "0x00...01",      // saleKind = 1 (Item)
    "0x...Alice"      // maker address
  ],
  "data": "0x9f8e7d6c..."  // Encoded: orderKey, nft, price, expiry, salt
}
```

---

## Phase 2: Backend Sync Service (Go)

### Step 1: Event Polling Loop

**File:** `EasySwapSync/service/orderbookindexer/service.go` (lines 105-181)

```go
func (s *Service) SyncOrderBookEventLoop() {
    lastSyncBlock := getFromDB()  // e.g., 5123450
    
    for {  // Infinite loop
        // 1. Get current blockchain height
        currentBlockNum, _ := s.chainClient.BlockNumber()
        
        // 2. Calculate batch (10 blocks at a time)
        startBlock := lastSyncBlock
        endBlock := startBlock + 10
        
        // 3. Fetch logs from blockchain
        query := types.FilterQuery{
            FromBlock: big.NewInt(startBlock),
            ToBlock:   big.NewInt(endBlock),
            Addresses: []string{OrderBookAddress},
        }
        logs, _ := s.chainClient.FilterLogs(s.ctx, query)
        
        // 4. Route each log by event signature
        for _, log := range logs {
            switch log.Topics[0].String() {
            case LogMakeTopic:     // "0xfc37f2ff..."
                s.handleMakeEvent(log)  // ← Process listing
            case LogCancelTopic:
                s.handleCancelEvent(log)
            case LogMatchTopic:
                s.handleMatchEvent(log)
            }
        }
        
        // 5. Update progress
        lastSyncBlock = endBlock + 1
        saveTooDB(lastSyncBlock)
    }
}
```

**Key Points:**
- Polls every 10 seconds
- Processes 10 blocks per batch
- Resumes from `last_indexed_block` if service restarts

---

### Step 2: Parse LogMake Event

**File:** `service.go` (lines 183-288)

```go
func (s *Service) handleMakeEvent(log ethereumTypes.Log) {
    // 1. Decode non-indexed data from log.Data
    var event struct {
        OrderKey [32]byte
        Nft      struct {
            TokenId        *big.Int
            CollectionAddr common.Address
            Amount         *big.Int
        }
        Price  *big.Int
        Expiry uint64
        Salt   uint64
    }
    
    s.parsedAbi.UnpackIntoInterface(&event, "LogMake", log.Data)
    
    // 2. Extract indexed fields from topics
    side := uint8(new(big.Int).SetBytes(log.Topics[1].Bytes()).Uint64())
    saleKind := uint8(new(big.Int).SetBytes(log.Topics[2].Bytes()).Uint64())
    maker := common.BytesToAddress(log.Topics[3].Bytes())
    
    // 3. Determine order type
    var orderType int64
    if side == Bid {
        orderType = saleKind == FixForCollection ? CollectionBidOrder : ItemBidOrder
    } else {
        orderType = ListingOrder  // = 1
    }
```

---

### Step 3: Save to Database

```go
    // 4. Create Order record
    newOrder := multi.Order{
        CollectionAddress: event.Nft.CollectionAddr.String(),
        TokenId:           event.Nft.TokenId.String(),
        OrderID:           "0x" + hex.EncodeToString(event.OrderKey[:]),
        OrderStatus:       OrderStatusActive,  // 0
        OrderType:         orderType,          // 1 (Listing)
        Price:             decimal.NewFromBigInt(event.Price, 0),
        Maker:             maker.String(),
        Taker:             ZeroAddress,
        ExpireTime:        int64(event.Expiry),
        QuantityRemaining: 1,
        Size:              1,
        Salt:              int64(event.Salt),
    }
    
    // 5. Insert into ob_order_* table
    s.db.Table("ob_order_sepolia").
        Clauses(clause.OnConflict{DoNothing: true}).
        Create(&newOrder)
    
    // 6. Create Activity record
    newActivity := multi.Activity{
        ActivityType:      Listing,  // 3
        Maker:             maker.String(),
        CollectionAddress: event.Nft.CollectionAddr.String(),
        TokenId:           event.Nft.TokenId.String(),
        Price:             decimal.NewFromBigInt(event.Price, 0),
        TxHash:            log.TxHash.String(),
        BlockNumber:       int64(log.BlockNumber),
        EventTime:         blockTime,
    }
    
    // 7. Insert into ob_activity_* table
    s.db.Table("ob_activity_sepolia").
        Clauses(clause.OnConflict{DoNothing: true}).
        Create(&newActivity)
}
```

---

## Final Database State

### `ob_order_sepolia` Table

| Column | Value |
|--------|-------|
| `order_id` | `0x9f8e7d6c5b4a3912abcd...` |
| `order_status` | `0` (Active) |
| `order_type` | `1` (Listing) |
| `collection_address` | `0xcoolcats...` |
| `token_id` | `42` |
| `price` | `500000000000000000` |
| `maker` | `0xalice...` |
| `taker` | `0x0000...` |
| `expire_time` | `1735084800` |
| `quantity_remaining` | `1` |

### `ob_activity_sepolia` Table

| Column | Value |
|--------|-------|
| `activity_type` | `3` (List) |
| `maker` | `0xalice...` |
| `collection_address` | `0xcoolcats...` |
| `token_id` | `42` |
| `price` | `500000000000000000` |
| `tx_hash` | `0xdef456789...` |
| `block_number` | `5123456` |

---

## Key Architectural Patterns

### 1. Event-Driven Indexing

**Problem:** Querying blockchain directly is slow and expensive.

**Solution:** 
- Smart contract emits events
- Backend service polls and indexes events
- Frontend queries fast SQL database

### 2. Escrow via Vault

**Why separate Vault from OrderBook?**
- **Security:** Isolates asset storage from trading logic
- **Upgradeability:** Can upgrade OrderBook without moving assets
- **Auditing:** Clear separation of concerns

### 3. Salt for Uniqueness

**Why use salt?**
- Allows same user to create multiple identical orders
- Prevents hash collisions
- Enables order editing (cancel old + create new with different salt)

### 4. Red-Black Tree Storage

**Why not simple array?**
- O(log n) insertion/deletion
- O(log n) to find best price
- Efficient for order matching

---

## Summary Flow

```
1. User approves NFT to Vault
2. User calls makeOrders() with order data
3. Contract validates order parameters
4. Contract generates unique orderKey (hash with salt)
5. Contract transfers NFT to Vault (escrow)
6. Contract stores order in Red-Black Tree
7. Contract emits LogMake event
8. Sync service polls blockchain for events
9. Sync service decodes LogMake event
10. Sync service inserts into ob_order_* and ob_activity_* tables
11. Frontend queries backend API (reads from MySQL)
```

---

## What Happens Next?

When a buyer wants to purchase:
1. Buyer calls `matchOrders()` with matching buy order
2. Contract validates both orders
3. ETH flows: Buyer → Seller (minus protocol fee)
4. NFT flows: Vault → Buyer
5. `LogMatch` event emitted
6. Sync service updates order status to `filled`
7. Activity record created with type `Sale`

---

## Code References

| Component | File | Key Functions |
|-----------|------|---------------|
| Order struct | `LibOrder.sol` | `hash()` |
| Main entry | `EasySwapOrderBook.sol` | `makeOrders()`, `_makeOrderTry()` |
| Escrow | `EasySwapVault.sol` | `depositNFT()` |
| Event polling | `service.go` | `SyncOrderBookEventLoop()` |
| Event parsing | `service.go` | `handleMakeEvent()` |

---

*Study Notes - NFT Marketplace Architecture*
*Focus: Smart Contract & Backend Implementation*
