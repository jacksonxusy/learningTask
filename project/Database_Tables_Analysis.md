# Database Tables Analysis

## Table Overview

| Table | Purpose | Updated By |
|-------|---------|------------|
| `ob_order_*` | Active orders (listings & bids) | LogMake, LogMatch, LogCancel |
| `ob_activity_*` | Transaction history | All events |
| `ob_item_*` | NFT metadata & ownership | LogMatch, sync jobs |
| `ob_collection_*` | Collection stats | Sync jobs |
| `ob_collection_floor_price_*` | Floor price history | Floor price worker |
| `ob_item_external_*` | NFT images/videos | Media sync |
| `ob_item_trait_*` | NFT attributes | Collection sync |
| `ob_indexed_status` | Sync progress | Event loop |
| `ob_user` | User access control | Admin |

---

## 1️⃣ `ob_order_*` - Orders

| Column | Usage |
|--------|-------|
| `order_id` | Unique order hash from blockchain |
| `order_status` | 0=Active, 1=Filled, 2=Cancelled |
| `order_type` | 1=Listing, 2=CollectionBid, 3=ItemBid |
| `maker` | Who created the order |
| `taker` | Who fulfilled the order |
| `price` | Order price in wei |
| `quantity_remaining` | For collection bids (can buy multiple) |
| `expire_time` | When order expires |

**Updated by:**
- `handleMakeEvent()` → INSERT (new order, status=Active)
- `handleMatchEvent()` → UPDATE (status=Filled, set taker)
- `handleCancelEvent()` → UPDATE (status=Cancelled)

---

## 2️⃣ `ob_activity_*` - Activity History

| Column | Usage |
|--------|-------|
| `activity_type` | 1=Buy, 3=List, 4=CancelListing, 9=CollectionBid, 10=ItemBid |
| `maker` | Initiator (seller for sales) |
| `taker` | Counterparty (buyer for sales) |
| `price` | Transaction price |
| `tx_hash` | Blockchain transaction hash |
| `event_time` | When it happened |

**Updated by:**
- `handleMakeEvent()` → INSERT (List/Bid activity)
- `handleMatchEvent()` → INSERT (Sale activity)
- `handleCancelEvent()` → INSERT (Cancel activity)

---

## 3️⃣ `ob_item_*` - NFT Items

| Column | Usage |
|--------|-------|
| `token_id` | NFT token ID |
| `collection_address` | Which collection |
| `owner` | Current owner address |
| `list_price` | Current listing price (if listed) |
| `sale_price` | Last sale price |

**Updated by:**
- `handleMatchEvent()` → UPDATE (owner = buyer)
- Collection sync jobs → INSERT (new items)

---

## 4️⃣ `ob_collection_*` - Collections

| Column | Usage |
|--------|-------|
| `address` | Collection contract address |
| `name` | Collection name |
| `floor_price` | Lowest listing price |
| `volume_total` | Total trading volume |
| `item_amount` | Total NFTs in collection |
| `owner_amount` | Number of unique holders |

**Updated by:**
- Collection import jobs → INSERT
- Floor price worker → UPDATE (floor_price)

---

## 5️⃣ `ob_collection_floor_price_*` - Floor Price History

| Column | Usage |
|--------|-------|
| `collection_address` | Which collection |
| `price` | Floor price at that time |
| `event_time` | Timestamp |

**Updated by:**
- `UpKeepingCollectionFloorChangeLoop()` → INSERT (price snapshots)

---

## 6️⃣ `ob_item_external_*` - NFT Media

| Column | Usage |
|--------|-------|
| `image_uri` | Original image URL |
| `oss_uri` | CDN/cloud image URL |
| `video_uri` | Video URL (if any) |
| `is_uploaded_oss` | Upload status |

**Updated by:**
- Media sync jobs → INSERT/UPDATE

---

## 7️⃣ `ob_item_trait_*` - NFT Traits

| Column | Usage |
|--------|-------|
| `trait` | Trait name (e.g., "Background") |
| `trait_value` | Value (e.g., "Blue") |

**Updated by:**
- Collection sync jobs → INSERT

---

## 8️⃣ `ob_indexed_status` - Sync Progress

| Column | Usage |
|--------|-------|
| `chain_id` | Which blockchain |
| `last_indexed_block` | Last processed block |
| `index_type` | Type of sync (6=events) |

**Updated by:**
- `SyncOrderBookEventLoop()` → UPDATE (last_indexed_block after each batch)

---

## Event → Table Mapping

| Event | Tables Updated |
|-------|---------------|
| `LogMake` | `ob_order` (INSERT), `ob_activity` (INSERT) |
| `LogMatch` | `ob_order` (UPDATE), `ob_activity` (INSERT), `ob_item` (UPDATE owner) |
| `LogCancel` | `ob_order` (UPDATE), `ob_activity` (INSERT) |

---

## Visual Relationships

```
ob_collection (1) ─────────────────┬───> ob_collection_floor_price (N)
       │                           │
       │ 1:N                       │
       ▼                           │
   ob_item (1) ───────────────────┬┴──> ob_item_external (1)
       │                          │
       │ 1:N                      └───> ob_item_trait (N)
       ▼
   ob_order ──────────────────────────> ob_activity
```

---

*Study Notes - NFT Marketplace Database Schema*
