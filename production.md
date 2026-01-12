# ApolloX Perpetual DEX 产品需求文档 (PRD)

## 1. 产品概述

### 产品定位与愿景
ApolloX 是一个基于 BSC 的去中心化永续合约交易所，采用 **LP-based AMM** 模型。我们的定位是"去中心化的币安合约"，提供接近中心化交易所的用户体验，同时保持资产自托管的去中心化优势。

愿景是成为 BSC 生态最大的衍生品 DEX，并逐步扩展到多链，成为 DeFi 衍生品的基础设施层。

### 目标用户画像
- **链上交易员**：日交易量 $10k-$100k，追求低滑点和高杠杆。
- **套利/量化交易者**：需要稳定的 API 接口，对延迟敏感。
- **流动性提供者 (LP)**：寻求 15-30% APY 的被动收益。
- **Degen 用户**：追求高倍杠杆 (100x+)，风险偏好高。

### 核心价值主张
- **Zero-slippage 执行**：基于 Oracle 价格，避免传统 AMM 的大额滑点。
- **高资金效率**：支持高达 100x 杠杆。
- **BSC 低成本**：交易成本比 Ethereum L1 低 95%。
- **CEX-grade API**：完全兼容币安 API 规范，开发者零成本迁移。

### 竞品差异化
- **与 GMX 相比**：我们在 BSC 上，Gas 更低，更适合高频交易。
- **与 dYdX 相比**：我们更去中心化，完全链上结算。
- **与 Perpetual Protocol 相比**：我们的 LP 池更简单，风险更可控。

---

## 2. 功能架构

### 核心功能模块
1. **交易引擎 (Trading Engine)**
   - **市价单**：基于 Mark Price 立即成交。
   - **限价单**：链上 order book，keeper 触发执行。
   - **止盈止损**：条件单自动触发。
2. **仓位管理 (Position Management)**
   - 实时盈亏计算。
   - 自动追加保证金/减仓。
   - 多空对冲。
3. **资金费率 (Funding Rate)**
   - 每 8 小时结算。
   - 平衡多空持仓量。
4. **清算系统 (Liquidation)**
   - 当维持保证金率 < 0.5% 时触发。
   - 分级清算：部分清算 → 全部清算。

### 技术架构
```text
[前端 Web/Mobile]
       ↓
[API Gateway - Go Backend]
       ↓
[Blockchain Client]
       ↓
[BSC Smart Contracts - Diamond Pattern]
  ├─ TradingPortalFacet (开平仓)
  ├─ LimitOrderFacet (限价单)
  └─ TradingReaderFacet (查询)
       ↓
[Event Indexer] → [PostgreSQL] → [WebSocket Push]
```

### 数据流程
用户下单 → API 签名验证 → Relayer 提交到链 → 合约执行 → 发出 Event → Indexer 监听 → 更新数据库 → WebSocket 推送给用户

---

## 3. 用户旅程

### 新用户注册和入金
1. 用户连接 MetaMask 钱包。
2. 一键授权 USDT 给合约。
3. 存入保证金 (Deposit)，获得账户余额。
4. 可选：生成 API Key 用于程序化交易。

### 开仓/平仓流程
**开仓：**
1. 选择交易对 (如 BTCUSDT)。
2. 设置杠杆 (1x-100x)。
3. 输入开仓金额，选择多/空。
4. 点击"开多"/"开空"。
5. MetaMask 弹窗确认 → 交易上链。
6. ~3秒后收到 WebSocket 推送，仓位已建立。

**平仓：**
1. 在"仓位"页面点击"平仓"。
2. 选择平仓比例 (25%/50%/100%)。
3. 确认 → 链上执行。
4. 盈亏结算到账户余额。

### 限价单管理
1. 选择"限价"模式。
2. 设置目标价格 (如 BTC $45,000)。
3. 提交订单 → 存入链上 order book。
4. 当 Mark Price 达到目标价，Keeper 自动执行。
5. 用户可随时取消未成交的限价单。

### 资产管理和提现
1. 在"资产"页面查看可用余额。
2. 点击"提现"，输入金额。
3. 提交链上交易，USDT 返回钱包。

---

## 4. 功能详细说明

### 4.1 交易执行
**功能描述**：支持市价单、限价单、止盈止损三种订单类型。

**用户故事**：
- "作为交易员，我希望在行情波动时能够立即成交，而不是等待对手方"
- "作为套利者，我需要挂限价单吃价差，等待最佳入场点"

**业务规则**：
- **市价单**：按 Mark Price ± 0.5% 的保护价立即成交。
- **限价单**：价格偏离当前价 > 0.1% 才允许挂单。
- **最小订单量**：$10 USDT。
- **最大订单量**：$100,000 USDT (单笔)。

**边界条件**：
- 如果 LP 池深度不足，市价单会部分成交。
- 如果用户保证金不足，订单自动拒绝。
- 网络拥堵时，订单可能延迟 10-30 秒。

**KPI**：
- 订单成功率 > 99.5%。
- API 响应时间 < 100ms (P99)。
- 滑点控制在 0.05% 以内。

### 4.2 Event Indexer (核心创新点)
**功能描述**：实时监听链上事件，替代传统的 Kafka 消息队列。

**为什么重要**：
- 保证数据一致性 (链上是唯一真相源)。
- 降低系统复杂度 (去掉 Kafka 依赖)。
- 提升用户体验 (WebSocket 实时推送)。

**技术细节**：
- 每 3 秒轮询一次 BSC 区块。
- 监听 4 种事件：`OpenMarketTrade`, `CloseTradeSuccessful`, `OpenLimitOrder`, `CancelLimitOrder`。
- 事件解码后立即写入 PostgreSQL。
- 同时触发 WebSocket 广播。

**异常处理**：
- 如果 Indexer 宕机，重启后自动从上次同步的区块继续。
- 如果出现链重组 (reorg)，回滚对应的数据库记录。

---

## 5. API 设计

### REST API 核心端点
- **交易类 (TRADE)**
  - `POST /fapi/v1/order` - 下单
  - `DELETE /fapi/v1/order` - 撤单
  - `DELETE /fapi/v1/allOpenOrders` - 全部撤单
- **查询类 (USER_DATA)**
  - `GET /fapi/v2/balance` - 账户余额
  - `GET /fapi/v2/account` - 完整账户信息
  - `GET /fapi/v2/positionRisk` - 仓位风险
  - `GET /fapi/v1/openOrders` - 当前挂单

### 请求示例
**下市价多单：**
```json
POST /fapi/v1/order
{
  "symbol": "BTCUSDT",
  "side": "BUY",
  "type": "MARKET",
  "quantity": "0.1",
  "timestamp": 1704902400000,
  "signature": "..."
}
```

**响应：**
```json
{
  "orderId": 12345,
  "symbol": "BTCUSDT",
  "status": "FILLED",
  "executedQty": "0.1",
  "executedPrice": "50000.00",
  "txHash": "0xabc123..."
}
```

### WebSocket 实时推送
**订阅个人数据流：**
```javascript
ws.send({
  "method": "SUBSCRIBE",
  "params": ["<listenKey>"],
  "id": 1
})
```

**收到订单更新：**
```json
{
  "e": "executionReport",
  "E": 1704902400000,
  "s": "BTCUSDT",
  "X": "FILLED",
  "i": 12345,
  "p": "50000.00",
  "q": "0.1"
}
```

### 错误码定义
- `-1000`：未知错误
- `-1001`：断开连接
- `-1021`：时间戳超出允许范围
- `-2010`：余额不足
- `-2011`：保证金不足
- `-2013`：订单不存在

---

## 6. 风险控制

### 清算机制
**触发条件**：
- 当 `可用保证金 / 持仓价值 < 维持保证金率 (0.5%)` 时触发。

**执行流程**：
1. 清算 Bot 发现可清算仓位。
2. 调用合约 `liquidate()` 函数。
3. 合约从 LP 池对冲仓位。
4. 清算罚金分配：80% 给 LP 池, 20% 给清算 Bot。

### 资金费率机制
**计算公式**：
```
Funding Rate = Clamp(
  (Mark Price - Index Price) / Index价格,
  -0.05%, 
  +0.05%
)
```
**结算时间**：每天 00:00, 08:00, 16:00 UTC。
**影响**：多头支付空头(或反之)，促使合约价格锚定现货。

### 滑点保护
- 用户可设置最大滑点 (如 0.5%)。
- 如果成交价超出保护范围，交易自动回滚。
- **默认保护**：市价单 ±0.5%, 限价单无滑点。

### 风险参数
- **初始保证金率**：1% (100x杠杆)。
- **维持保证金率**：0.5%。
- **最大持仓**：$500,000 USDT (单用户单交易对)。
- **资金费率上限**：±0.05%。

---

## 7. 运营策略

### 流动性激励方案
**ALP (ApolloX LP) 代币机制**：
- 用户存入 USDT → 获得 ALP 代币。
- ALP 价值 = LP 池总价值 / ALP 总供应量。
- LP 分享 70% 的交易手续费。
- LP 承担交易员的盈利 (作为对手方)。
**预期 APY**：15-30% (取决于交易量和交易员胜率)。

### 手续费结构
- **Taker 费率**：0.06% (市价单)
- **Maker 费率**：0.01% (限价单)
- **清算费**：0.5% (额外惩罚)

**收入分配**：
- 70% → LP 池
- 20% → 协议金库 (用于回购代币)
- 10% → 团队运营

### 用户增长策略
- **交易挖矿**：每 $100 交易量 → 1 $APX 代币。
- **推荐计划**：邀请好友，获得其手续费的 20% 返佣。
- **做市商激励**：大额 LP 提供者额外获得代币奖励。
- **跨链桥补贴**：用户从其他链桥接资产到 BSC，补贴 Gas 费。

---

## 8. 技术实现

### 智能合约关键逻辑
采用 **Diamond Pattern (EIP-2535)**：
- `TradingPortalFacet`：处理开平仓逻辑。
- `LimitOrderFacet`：管理限价单队列。
- `TradingReaderFacet`：只读查询。

**核心函数**：
```solidity
function openMarketTrade(
  address pairBase,
  bool isLong,
  uint256 price,
  uint256 qty,
  address tokenIn,
  uint256 amountIn
) external returns (bytes32 tradeHash)
```

**Gas 优化**：
- 使用 `uint256 packed` 存储多个变量。
- Event 数据最小化 (只存索引字段)。
- 批量操作合并为单笔交易。

### 后端服务架构
Go 语言, 分层设计：
```
API Layer (Gin)
  ↓
Service Layer (业务逻辑)
  ↓
Blockchain Client (ethclient)
  ↓
Smart Contracts
```
**关键优化**：
- Redis 缓存 Mark Price (5s TTL)。
- PostgreSQL 读写分离。
- Relayer 交易队列 (防止 nonce 冲突)。

### Event Indexer 同步机制
1. 每 3 秒查询最新区块。
2. 过滤指定合约地址的 Event。
3. 根据 Event 类型解码数据。
4. 写入数据库 + 触发 WebSocket。

**性能**：
- 单 Indexer 可处理 1000 TPS。
- 延迟 < 5 秒 (从链上到用户收到推送)。

---

## 9. 路线图

### MVP (Q1 2026) - ✅ 已完成
- 基础的市价单/限价单交易
- 仓位管理和清算
- Event Indexer 同步系统
- V1 + V2 REST API
- WebSocket 实时推送

### Phase 1: Degen Mode (Q2 2026)
- 最高 500x 杠杆
- 单项保证金 (isolated margin)
- 闪电清仓 (flash liquidation)
- 社交交易 (跟单功能)
**KPI**：日均交易量 $10M

### Phase 2: 跨链扩展 (Q3 2026)
- 部署至 Arbitrum 和 Polygon。
- 跨链资产桥接。
- 统一账户 (一个钱包管理多链仓位)。
**KPI**：多链总交易量 $50M/day

### Phase 3: 生态集成 (Q4 2026)
- SDK / API 文档完善。
- 策略机器人市场。
- 第三方清算 Bot 激励。
**KPI**：API 调用量 1B+/month

---

## 总结
ApolloX Perpetual DEX 通过从 P2P 模式转型为 **LP-based AMM**，结合 BSC 的低成本优势和完全兼容币安 API 的设计，旨在为去中心化衍生品市场带来 CEX 级别的用户体验。我们的核心竞争力在于技术架构的创新 (Event Indexer 替代 Kafka)、风险控制的稳健性以及对开发者友好的 API 设计。
