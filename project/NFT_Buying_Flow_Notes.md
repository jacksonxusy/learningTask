# 📘 Deep Dive: NFT Buying Flow Architecture

## 1. Overview

This document explains the complete flow when a buyer purchases an NFT from an existing listing, focusing on **Smart Contract** and **Backend** implementation.

**Key Difference from Listing:** The buyer creates a **temporary order** (not saved on-chain) that matches against the seller's **persistent order** (already on-chain).

---

## 2. Example Scenario

**Bob buys CoolCat #42 from Alice's listing at 0.5 ETH**

| Parameter | Value |
|-----------|-------|
| **Buyer** | `0xBob567890abcdef567890abcdef567890abcdef56` |
| **Seller** | `0xAlice1234567890abcdef1234567890abcdef1234` |
| **NFT Collection** | `0xCoolCats000000000000000000000000000000CC` |
| **Token ID** | `42` |
| **Listing Price** | `0.5 ETH` (500000000000000000 wei) |
| **Alice's Order ID** | `0x9f8e7d6c5b4a3912abcd...` (already on-chain) |
| **Protocol Fee** | `2.5%` (250 basis points) |

---

## 3. Phase 1: Smart Contract Execution

### Step 1: Bob Constructs Matching Buy Order

Bob's frontend constructs a temporary buy order:

```solidity
Order bobBuyOrder = {
    side: 1,              // Bid (buying)
    saleKind: 1,          // FixedPriceForItem
    maker: 0xBob...,
    nft: {
        tokenId: 42,
        collection: 0xCoolCats...,
        amount: 1
    },
    price: 500000000000000000,  // 0.5 ETH (matches Alice's price)
    expiry: block.timestamp + 3600,  // Valid for 1 hour
    salt: 1703420000000   // Random nonce
};
```

**📌 Critical:** This order is NOT saved to the blockchain. It's only used for validation during matching.

---

### Step 2: Call `matchOrders()` with ETH

**File:** `EasySwapOrderBook.sol` (lines 248-291)

```solidity
// Bob sends transaction
EasySwapOrderBook.matchOrders([
    MatchDetail({
        sellOrder: aliceSellOrder,  // Fetched from chain/backend
        buyOrder: bobBuyOrder       // Constructed temporarily
    })
], { value: 0.5 ETH })  // Bob sends ETH with transaction
```

**Contract Processing:**

```solidity
function matchOrders(MatchDetail[] calldata matchDetails)
    external payable whenNotPaused nonReentrant
    returns (bool[] memory successes)
{
    uint128 buyETHAmount;  // Track total ETH spent
    
    for (uint256 i = 0; i < matchDetails.length; ++i) {
        // Use delegatecall to process each match
        (bool success, bytes memory data) = address(this).delegatecall(
            abi.encodeWithSignature(
                "matchOrderWithoutPayback(...)",
                matchDetail.sellOrder,
                matchDetail.buyOrder,
                msg.value - buyETHAmount
            )
        );
        
        if (success) {
            if (matchDetail.buyOrder.maker == _msgSender()) {
                // Bob is buying
                uint128 buyPrice = abi.decode(data, (uint128));
                buyETHAmount += buyPrice;  // Track ETH spent
            }
        }
    }
    
    // Refund excess ETH if Bob overpaid
    if (msg.value > buyETHAmount) {
        _msgSender().safeTransferETH(msg.value - buyETHAmount);
    }
}
```

---

### Step 3: Internal Matching `_matchOrder()`

**File:** `EasySwapOrderBook.sol` (lines 473-578)

This function has two branches depending on who initiates the match.

#### 3.1 Validation Phase

```solidity
function _matchOrder(
    LibOrder.Order calldata sellOrder,  // Alice's order
    LibOrder.Order calldata buyOrder,   // Bob's order
    uint256 msgValue                    // 0.5 ETH
) internal returns (uint128 costValue) {
    
    // Generate order keys (unique identifiers)
    OrderKey sellOrderKey = LibOrder.hash(sellOrder);
    // = 0x9f8e7d6c... (Alice's existing order)
    
    OrderKey buyOrderKey = LibOrder.hash(buyOrder);
    // = 0xabc123... (Bob's temporary order)
    
    // ═══════════════════════════════════════════════════════════════
    // VALIDATION: Can these orders be matched?
    // ═══════════════════════════════════════════════════════════════
    _isMatchAvailable(sellOrder, buyOrder, sellOrderKey, buyOrderKey);
    
    // Checks performed:
    // 1. sellOrder.side == List && buyOrder.side == Bid
    // 2. sellOrder.maker != buyOrder.maker (Alice != Bob)
    // 3. NFT matches (same collection + tokenId)
    // 4. Neither order is already filled
    // 5. sellOrder.saleKind == FixedPriceForItem
```

#### 3.2 Buyer Branch (Bob's Case)

```solidity
    // ═══════════════════════════════════════════════════════════════
    // DETERMINE WHO IS CALLING
    // ═══════════════════════════════════════════════════════════════
    
    if (_msgSender() == buyOrder.maker) {  // ✅ Bob is the buyer
        
        // Check if Bob's order exists on-chain
        bool isBuyExist = orders[buyOrderKey].order.maker != address(0);
        // isBuyExist = false (Bob's order is temporary)
        
        // Validate Alice's sell order EXISTS on-chain
        _validateOrder(orders[sellOrderKey].order, false);
        // This ensures Alice's order is actually stored and valid
        
        // Validate Bob's buy order parameters
        _validateOrder(buyOrder, isBuyExist);
        
        uint128 buyPrice = Price.unwrap(buyOrder.price);    // 0.5 ETH
        uint128 fillPrice = Price.unwrap(sellOrder.price);  // 0.5 ETH
        
        // ═══════════════════════════════════════════════════════════
        // ETH PAYMENT VALIDATION
        // ═══════════════════════════════════════════════════════════
        if (!isBuyExist) {  // ✅ Temporary order
            require(msgValue >= fillPrice, "HD: value < fill price");
            // Bob must send at least 0.5 ETH with transaction
        } else {
            // If Bob had a pre-existing bid order, ETH would be in Vault
            require(buyPrice >= fillPrice, "HD: buy price < fill price");
            IEasySwapVault(_vault).withdrawETH(buyOrderKey, buyPrice, address(this));
            _removeOrder(buyOrder);
            _updateFilledAmount(filledAmount[buyOrderKey] + 1, buyOrderKey);
        }
        
        // ═══════════════════════════════════════════════════════════
        // MARK ALICE'S ORDER AS FILLED
        // ═══════════════════════════════════════════════════════════
        _updateFilledAmount(sellOrder.nft.amount, sellOrderKey);
        // filledAmount[0x9f8e7d6c...] = 1 (fully filled)
        
        // ═══════════════════════════════════════════════════════════
        // EMIT MATCH EVENT
        // ═══════════════════════════════════════════════════════════
        emit LogMatch(
            buyOrderKey,    // Bob's order key (indexed topic[1])
            sellOrderKey,   // Alice's order key (indexed topic[2])
            buyOrder,       // Bob's order data
            sellOrder,      // Alice's order data
            fillPrice       // 0.5 ETH
        );
```

---

### Step 4: Asset Transfers

```solidity
        // ═══════════════════════════════════════════════════════════
        // TRANSFER 1: ETH → Alice (minus protocol fee)
        // ═══════════════════════════════════════════════════════════
        uint128 protocolFee = _shareToAmount(fillPrice, protocolShare);
        // protocolFee = 500000000000000000 * 250 / 10000
        //             = 12500000000000000 (0.0125 ETH)
        
        sellOrder.maker.safeTransferETH(fillPrice - protocolFee);
        // Alice receives: 0.5 - 0.0125 = 0.4875 ETH
        
        // Refund Bob if he sent more than needed
        if (buyPrice > fillPrice) {
            buyOrder.maker.safeTransferETH(buyPrice - fillPrice);
        }
        
        // ═══════════════════════════════════════════════════════════
        // TRANSFER 2: NFT → Bob
        // ═══════════════════════════════════════════════════════════
        IEasySwapVault(_vault).withdrawNFT(
            sellOrderKey,              // Alice's order key
            buyOrder.maker,            // Bob (recipient)
            sellOrder.nft.collection,  // CoolCats
            sellOrder.nft.tokenId      // 42
        );
        
        costValue = isBuyExist ? 0 : buyPrice;  // Return 0.5 ETH
    }
}
```

**Vault Execution** (`EasySwapVault.sol` lines 68-78):

```solidity
function withdrawNFT(
    OrderKey orderKey,   // 0x9f8e7d6c...
    address to,          // Bob
    address collection,  // CoolCats
    uint256 tokenId      // 42
) external onlyEasySwapOrderBook {
    require(NFTBalance[orderKey] == tokenId, "HV: not match tokenId");
    delete NFTBalance[orderKey];  // Clear vault storage
    
    // Transfer NFT: Vault → Bob
    IERC721(collection).safeTransferNFT(address(this), to, tokenId);
}
```

**On-Chain State Changes:**

| Contract | Storage | Before | After |
|----------|---------|--------|-------|
| CoolCats | `_owners[42]` | `0xVault...` | `0xBob...` |
| Vault | `NFTBalance[0x9f8e7d6c...]` | `42` | `0` (deleted) |
| OrderBook | `filledAmount[0x9f8e7d6c...]` | `0` | `1` |
| Alice | ETH balance | `X` | `X + 0.4875` |
| Bob | ETH balance | `Y` | `Y - 0.5` |
| OrderBook | ETH balance | `Z` | `Z + 0.0125` |

---

## 4. Phase 2: Backend Sync Service

### Step 1: Event Polling Detects LogMatch

**File:** `service.go` (lines 105-181)

The sync loop detects the event:

```go
for _, log := range logs {
    switch log.Topics[0].String() {
    case LogMatchTopic:  // "0xf629aeca..."
        s.handleMatchEvent(log)  // ← Process sale
    }
}
```

**Raw Event Log:**

```json
{
  "address": "0xOrderBookContract...",
  "topics": [
    "0xf629aecab94607bc43ce4aebd564bf6e61c7327226a797b002de724b9944b20e",  // LogMatch
    "0x...abc123",  // buyOrderKey (Bob's)
    "0x...9f8e7d"   // sellOrderKey (Alice's)
  ],
  "data": "0x..."  // Encoded: makeOrder, takeOrder, fillPrice
}
```

---

### Step 2: Parse LogMatch Event

**File:** `service.go` (lines 290-458)

```go
func (s *Service) handleMatchEvent(log ethereumTypes.Log) {
    // ═══════════════════════════════════════════════════════════════
    // DECODE EVENT DATA
    // ═══════════════════════════════════════════════════════════════
    var event struct {
        MakeOrder Order      // First order in match
        TakeOrder Order      // Second order in match
        FillPrice *big.Int   // Actual sale price
    }
    
    s.parsedAbi.UnpackIntoInterface(&event, "LogMatch", log.Data)
    
    // Extract order IDs from topics
    makeOrderId := "0x" + hex.EncodeToString(log.Topics[1].Bytes())
    // = Bob's buy order key
    
    takeOrderId := "0x" + hex.EncodeToString(log.Topics[2].Bytes())
    // = Alice's sell order key (0x9f8e7d6c...)
    
    // ═══════════════════════════════════════════════════════════════
    // DETERMINE TRADE DIRECTION
    // ═══════════════════════════════════════════════════════════════
    var owner string      // New NFT owner
    var from string       // Seller
    var to string         // Buyer
    var sellOrderId string
    
    if event.MakeOrder.Side == Bid {
        // ✅ MakeOrder is buy order → Buyer initiated (Bob's case)
        owner = event.MakeOrder.Maker.String()     // Bob (new owner)
        collection = event.TakeOrder.Nft.CollectionAddr.String()
        tokenId = event.TakeOrder.Nft.TokenId.String()
        from = event.TakeOrder.Maker.String()      // Alice
        to = event.MakeOrder.Maker.String()        // Bob
        sellOrderId = takeOrderId                  // Alice's order
```

---

### Step 3: Update Alice's Order Status

```go
        // ═══════════════════════════════════════════════════════════
        // UPDATE SELL ORDER → FILLED
        // ═══════════════════════════════════════════════════════════
        s.db.Table("ob_order_sepolia").
            Where("order_id = ?", takeOrderId).  // Alice's 0x9f8e7d6c...
            Updates(map[string]interface{}{
                "order_status":       OrderStatusFilled,  // 0 → 1
                "quantity_remaining": 0,                   // 1 → 0
                "taker":              to,                  // Bob's address
            })
        
        // ═══════════════════════════════════════════════════════════
        // CHECK BOB'S ORDER (may not exist if temporary)
        // ═══════════════════════════════════════════════════════════
        var buyOrder Order
        err := s.db.Table("ob_order_sepolia").
            Where("order_id = ?", makeOrderId).
            First(&buyOrder).Error
        
        if err != nil {
            // Bob's order doesn't exist → it was temporary
            // This is expected for market buys
            return
        }
        
        // If Bob had a pre-existing bid, update it
        if buyOrder.QuantityRemaining > 1 {
            // Partial fill (for collection bids)
            s.db.Update("quantity_remaining", buyOrder.QuantityRemaining - 1)
        } else {
            // Fully filled
            s.db.Updates(map[string]interface{}{
                "order_status":       OrderStatusFilled,
                "quantity_remaining": 0,
            })
        }
    }
```

---

### Step 4: Create Sale Activity Record

```go
    // ═══════════════════════════════════════════════════════════════
    // CREATE ACTIVITY (SALE EVENT)
    // ═══════════════════════════════════════════════════════════════
    blockTime, _ := s.chainClient.BlockTimeByNumber(
        s.ctx, 
        big.NewInt(int64(log.BlockNumber)),
    )
    
    newActivity := Activity{
        ActivityType:      Sale,  // type = 1 (Buy/Sale)
        Maker:             event.MakeOrder.Maker.String(),  // Bob
        Taker:             event.TakeOrder.Maker.String(),  // Alice
        MarketplaceID:     MarketOrderBook,
        CollectionAddress: collection,
        TokenId:           tokenId,
        CurrencyAddress:   EthAddress,
        Price:             decimal.NewFromBigInt(event.FillPrice, 0),
        BlockNumber:       int64(log.BlockNumber),
        TxHash:            log.TxHash.String(),
        EventTime:         int64(blockTime),
    }
    
    s.db.Table("ob_activity_sepolia").
        Clauses(clause.OnConflict{DoNothing: true}).
        Create(&newActivity)
```

---

### Step 5: Update NFT Owner in Database

```go
    // ═══════════════════════════════════════════════════════════════
    // UPDATE ITEM OWNER: Alice → Bob
    // ═══════════════════════════════════════════════════════════════
    s.db.Table("ob_item_sepolia").
        Where("collection_address = ? AND token_id = ?", 
              strings.ToLower(collection), tokenId).
        Update("owner", owner)  // Bob's address
    
    // ═══════════════════════════════════════════════════════════════
    // TRIGGER PRICE UPDATE QUEUE (for floor price recalculation)
    // ═══════════════════════════════════════════════════════════════
    ordermanager.AddUpdatePriceEvent(s.kv, &TradeEvent{
        OrderId:        sellOrderId,
        CollectionAddr: collection,
        EventType:      Buy,
        TokenID:        tokenId,
        From:           from,  // Alice
        To:             to,    // Bob
    }, s.chain)
}
```

---

## 5. Final Database State

### `ob_order_sepolia` (Alice's Order Updated)

| Column | Before | After |
|--------|--------|-------|
| `order_id` | `0x9f8e7d6c...` | `0x9f8e7d6c...` |
| `order_status` | `0` (Active) | `1` (Filled) |
| `order_type` | `1` (Listing) | `1` (Listing) |
| `maker` | `0xalice...` | `0xalice...` |
| `taker` | `0x0000...` | `0xbob...` |
| `quantity_remaining` | `1` | `0` |

### `ob_activity_sepolia` (New Sale Record)

| Column | Value |
|--------|-------|
| `activity_type` | `1` (Sale) |
| `maker` | `0xbob...` (buyer) |
| `taker` | `0xalice...` (seller) |
| `collection_address` | `0xcoolcats...` |
| `token_id` | `42` |
| `price` | `500000000000000000` |
| `tx_hash` | `0xabc123...` |
| `block_number` | `5123500` |
| `event_time` | `1703420000` |

### `ob_item_sepolia` (Owner Updated)

| Column | Before | After |
|--------|--------|-------|
| `owner` | `0xalice...` | `0xbob...` |
| `sale_price` | `0` or previous | `500000000000000000` |

---

## 6. Key Architectural Insights

### 6.1 Two Types of Buy Orders

| Type | Persistent Bid | Temporary Buy |
|------|----------------|---------------|
| **Saved on-chain?** | ✅ Yes | ❌ No |
| **ETH escrowed?** | ✅ In Vault | ❌ Sent with TX |
| **Use case** | Standing offer for collection | Market buy (instant) |
| **Example** | "I'll pay 1 ETH for any CoolCat" | "I want CoolCat #42 now" |

### 6.2 Protocol Fee Structure

```
Sale Price:     0.5 ETH
Protocol Fee:   0.0125 ETH (2.5%)
Seller Receives: 0.4875 ETH
```

Fee is **deducted from seller**, not added to buyer's cost.

### 6.3 Atomic Transaction

All operations happen in one transaction:
1. Validate orders
2. Transfer ETH
3. Transfer NFT
4. Emit event

**Result:** Either everything succeeds or everything reverts. No partial fills for ERC721.

### 6.4 Event-Driven Updates

```
Blockchain (LogMatch) → Sync Service → Database → Frontend
```

The frontend doesn't query the blockchain directly. It queries the fast MySQL database that mirrors blockchain state.

---

## 7. Comparison: Listing vs Buying

| Aspect | Listing Flow | Buying Flow |
|--------|--------------|-------------|
| **Order saved on-chain?** | ✅ Yes | ❌ No (temporary) |
| **Asset escrowed?** | ✅ NFT in Vault | ❌ ETH sent with TX |
| **Who pays gas?** | Seller | Buyer |
| **Event emitted** | `LogMake` | `LogMatch` |
| **DB operation** | INSERT order | UPDATE order status |
| **Activity type** | `3` (List) | `1` (Sale) |

---

## 8. Complete Flow Summary

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Bob fetches Alice's active listing from backend API          │
│ 2. Bob constructs temporary buy order matching Alice's order    │
│ 3. Bob calls matchOrders() with 0.5 ETH                         │
│ 4. Contract validates: orders match, Alice's order exists       │
│ 5. Contract calculates protocol fee (2.5% = 0.0125 ETH)         │
│ 6. Contract transfers ETH: Bob → Alice (0.4875 ETH)             │
│ 7. Contract transfers ETH: Bob → OrderBook (0.0125 ETH fee)     │
│ 8. Contract transfers NFT: Vault → Bob                          │
│ 9. Contract marks Alice's order as filled                       │
│ 10. Contract emits LogMatch event                               │
│ 11. Sync service detects event in next poll                     │
│ 12. Sync service updates Alice's order status → Filled          │
│ 13. Sync service creates Sale activity record                   │
│ 14. Sync service updates NFT owner → Bob                        │
│ 15. Sync service triggers floor price recalculation             │
│ 16. Frontend queries backend, shows Bob as new owner            │
└─────────────────────────────────────────────────────────────────┘
```

---

## 9. Code References

| Component | File | Key Functions |
|-----------|------|---------------|
| Match entry | `EasySwapOrderBook.sol` | `matchOrders()`, `_matchOrder()` |
| Validation | `EasySwapOrderBook.sol` | `_isMatchAvailable()`, `_validateOrder()` |
| NFT transfer | `EasySwapVault.sol` | `withdrawNFT()` |
| Event parsing | `service.go` | `handleMatchEvent()` |
| DB updates | `service.go` | Order status, activity, owner updates |

---

*Study Notes - NFT Marketplace Architecture*  
*Focus: Smart Contract Matching & Backend Event Processing*
