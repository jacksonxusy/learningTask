NFT Marketplace Architecture Overview
🏗️ High-Level Architecture

┌─────────────────────────────────────────────────────────────────────────────┐
│                              USER (Browser/Wallet)                          │
└─────────────────────────────────────────────────────────────────────────────┘
                                        │
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
        ┌─────────────────┐   ┌─────────────────┐   ┌───────────────┐
        │   Frontend      │   │  Smart Contract │   │   Wallet      │
        │  (Next.js)      │   │  (On-Chain DEX) │   │  (MetaMask)   │
        │  nft-market-fe  │   │ EasySwapContract│   │               │
        └────────┬────────┘   └────────┬────────┘   └───────────────┘
                 │                     │
                 ▼                     │
        ┌─────────────────┐            │
        │  Backend API    │            │
        │ EasySwapBackend │◄───────────┘
        │    (Go/Gin)     │
        └────────┬────────┘
                 │
        ┌────────┴────────┐
        ▼                 ▼
┌─────────────────┐  ┌─────────────────┐
│  EasySwapSync   │  │  EasySwapBase   │
│ (Event Indexer) │  │ (Shared Libs)   │
│    Go service   │  │   Go modules    │
└────────┬────────┘  └─────────────────┘
         │
         ▼
┌─────────────────┐
│    MySQL DB     │
│  (ob_* tables)  │
└─────────────────┘

📦 Project Structure & Responsibilities
1. Smart Contracts (EasySwapContract/)
Language: Solidity | Framework: Hardhat

Module	            File	           Responsibility
EasySwapOrderBook	EasySwapOrderBook.sol. Core DEX logic - order matching, creation, cancellation
EasySwapVault	   EasySwapVault.sol.    Escrow for NFTs and ETH during trades
OrderStorage	OrderStorage.sol.     On-chain order storage data structures
OrderValidator	OrderValidator.sol     EIP-712 signature validation, order validity checks
ProtocolManager	 ProtocolManager.sol    Protocol fee management
LibOrder	libraries/LibOrder.sol	Order struct definitions, hashing utilities

Key Contract Functions:
makeOrders()
 → Create limit sell/buy orders
cancelOrders() → Cancel existing orders
editOrders() → Modify order price (cancel + recreate)
matchOrders() → Execute trades (atomic swaps)

3. Backend API (EasySwapBackend/)
Language: Go | Framework: Gin HTTP server

Directory	Responsibility
src/api/v1/	HTTP endpoint handlers
src/service/v1/	Business logic layer
src/dao/	Data access objects (MySQL queries)
src/types/	Request/Response DTOs
src/common/	Shared utilities
API Endpoints (derived from code):

Collection APIs: List collections, get collection details, items, rankings
Order APIs: Query orders by collection/item
Activity APIs: Get trade history, mint events
Portfolio APIs: User holdings, order history
User APIs: User profile info

4. Event Sync Service (EasySwapSync/)
Language: Go | Purpose: Blockchain event indexer

Directory	Responsibility
service/orderbookindexer/	Index on-chain order events
service/collectionfilter/	Filter and sync collection data
service/comm/	Shared service utilities
Indexed Events:

LogMake → New order created
LogCancel → Order cancelled
LogMatch → Trade executed

5. Shared Base Library (EasySwapBase/)
Language: Go | Purpose: Common utilities across backend services

Module	Responsibility
chain/	Multi-chain client abstraction
evm/	EVM-specific utilities
stores/gdb/	GORM database models and queries
ordermanager/	Floor price calculation, expired order cleanup
kit/	Generic utilities (encoding, pagination)
xhttp/	HTTP client wrappers

🔄 Step-by-Step Flow Diagrams
Flow 1: Listing an NFT for Sale
┌───────────┐   ┌───────────┐   ┌───────────┐   ┌───────────┐   ┌───────────┐
│  User UI  │   │ Frontend  │   │  Wallet   │   │ Contract  │   │  Backend  │
└─────┬─────┘   └─────┬─────┘   └─────┬─────┘   └─────┬─────┘   └─────┬─────┘
      │               │               │               │               │
      │ 1. Click      │               │               │               │
      │    "List NFT" │               │               │               │
      │──────────────►│               │               │               │
      │               │ 2. Check NFT  │               │               │
      │               │    approval   │               │               │
      │               │──────────────►│               │               │
      │               │◄──────────────│               │               │
      │               │               │               │               │
      │               │ 3. Approve    │               │               │
      │               │   EasySwapVault               │               │
      │               │──────────────►│               │               │
      │               │      4. Sign  │               │               │
      │               │         TX    │               │               │
      │               │◄──────────────│───────────────►│               │
      │               │               │               │               │
      │               │ 5. makeOrders │               │               │
      │               │   (List order)│               │               │
      │               │──────────────►│───────────────►│               │
      │               │               │               │ 6. NFT → Vault│
      │               │               │               │   (escrow)    │
      │               │               │               │               │
      │               │               │               │ 7. Emit       │
      │               │               │               │    LogMake    │
      │               │               │               │──────────────►│
      │               │               │               │               │ 8. Index event
      │               │               │               │               │    Update DB
      │               │               │               │               │
Code Path:


1. Frontend: app/collections/[name]/[tokenId]/page.tsx → List button
2. Frontend:   contracts/service/orderBookContract.ts →   approveNFT(),   makeOrders()
3. Contract:   EasySwapOrderBook.sol →   makeOrders() → _makeOrderTry()
4. Contract: NFT transferred to EasySwapVault
5. Sync:   orderbookindexer/service.go →   handleMakeEvent()
6. Backend:   dao/items.go → Update ob_order_* table

Flow 2: Buying an NFT (Market Order)
┌───────────┐   ┌───────────┐   ┌───────────┐   ┌───────────┐   ┌───────────┐
│  User UI  │   │ Frontend  │   │  Wallet   │   │ Contract  │   │  Backend  │
└─────┬─────┘   └─────┬─────┘   └─────┬─────┘   └─────┬─────┘   └─────┬─────┘
      │               │               │               │               │
      │ 1. Click "Buy"│               │               │               │
      │──────────────►│               │               │               │
      │               │ 2. Fetch      │               │               │
      │               │    listing    │               │               │
      │               │    order from │               │               │
      │               │    contract   │               │               │
      │               │──────────────────────────────►│               │
      │               │◄──────────────────────────────│               │
      │               │               │               │               │
      │               │ 3. Construct  │               │               │
      │               │   buy order + │               │               │
      │               │   matchOrders │               │               │
      │               │──────────────►│               │               │
      │               │      4. Sign  │ 5. TX + ETH   │               │
      │               │         TX    │   {value}     │               │
      │               │◄──────────────│───────────────►│               │
      │               │               │               │               │
      │               │               │               │ 6. Validate   │
      │               │               │               │    orders     │
      │               │               │               │ 7. Transfer:  │
      │               │               │               │  - ETH → Seller
      │               │               │               │  - NFT → Buyer│
      │               │               │               │ 8. Emit       │
      │               │               │               │    LogMatch   │
      │               │               │               │──────────────►│
      │               │               │               │               │ 9. Index event
      │               │               │               │               │    Create activity
      │               │               │               │               │    Update owner

Code Path:
1. Frontend: Item detail page → "Buy Now" button
2. Frontend:   orderBookContract.ts →   getOrders() to fetch sell order
3. Contract:   EasySwapOrderBook.sol → matchOrders() or matchOrder()
4. Contract: Internal _matchOrder() → validates and executes atomic swap
5. Sync:   orderbookindexer/service.go →   handleMatchEvent()
6. Backend: Creates ob_activity_* record (type=Buy), updates ob_item_* owner


Flow 3: Making a Bid Offer

User → Frontend (construct bid order with saleKind=FixedPriceForItem or FixedPriceForCollection)
     → Wallet (sign TX, attach ETH value)
     → EasySwapOrderBook.makeOrders() (Side=Bid)
     → ETH transferred to EasySwapVault (escrow)
     → LogMake event emitted
     → EasySwapSync indexes bid order
     → Backend updates ob_order table (order_type=4 for item bid, 3 for collection bid)

Flow 4: Accepting a Bid (Seller Action)
Seller → Frontend (fetch bid order)
       → Wallet (sign TX + approve NFT if needed)
       → EasySwapOrderBook.matchOrders() (sellOrder created, matched with buyOrder)
       → NFT → Buyer, ETH (from vault) → Seller
       → LogMatch event
       → Activity record created (type=Sell)

🔧 Key Implementation Patterns
1. On-Chain Order Book (DEX Model)
Orders are stored on-chain in OrderStorage
Assets are escrowed in EasySwapVault upon order creation
Atomic matching ensures trustless trades
2. Event-Driven Architecture
Smart contract emits events (LogMake, LogCancel, LogMatch)
EasySwapSync listens via RPC, indexes to MySQL
Backend serves indexed data to frontend (faster reads)
3. Separation of Concerns
┌─────────────────────────────────────────────────────────────────┐
│ EasySwapBase: Shared libraries (chain clients, DB models)      │
├─────────────────────────────────────────────────────────────────┤
│ EasySwapSync: Blockchain → Database sync                       │
├─────────────────────────────────────────────────────────────────┤
│ EasySwapBackend: Database → REST API                           │
├─────────────────────────────────────────────────────────────────┤
│ Frontend: API + Contract → UI                                  │
└─────────────────────────────────────────────────────────────────┘
4. Order Matching Logic
Sell Order (List): NFT locked in vault, waiting for buyer
Buy Order (Bid): ETH locked in vault, waiting for seller
Match: Construct counter-order with same params, call matchOrders()


📚 Reading Guide (Suggested Order)
Start with Contracts:
IEasySwapOrderBook.sol
 → Interface overview
LibOrder.sol → Order struct definition
EasySwapOrderBook.sol
 → Core logic
Understand Data Flow:
EasySwapSync/orderbookindexer/service.go → Event handling
EasySwapBase/stores/gdb/ → Database schema
Backend API:
EasySwapBackend/src/api/v1/ → HTTP handlers
EasySwapBackend/src/dao/ → Query layer
Frontend Integration:
nft-market-fe/contracts/service/orderBookContract.ts → Contract calls
nft-market-fe/api/ → Backend API clients
nft-market-fe/app/collections/ → UI pages




