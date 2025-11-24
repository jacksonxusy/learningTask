# NFT 拍卖项目 - Solidity 智能合约

这是一个基于 Solidity 开发的 NFT 拍卖系统项目，使用 Hardhat 框架进行智能合约的开发、测试和部署。

## 项目概述

该项目实现了完整的 NFT 拍卖功能，包含以下核心组件：

- **JacksonNFT**: NFT 合约，支持铸造和管理 NFT
- **Auction**: 拍卖合约，支持 ETH 和 ERC20 代币出价
- **Chainlink 价格预言机**: 实时获取 ETH/USD 汇率
- **可升级合约**: 使用 OpenZeppelin 升级模式

## 技术栈

- **智能合约**: Solidity 0.8.28
- **开发框架**: Hardhat 3.0 (Beta)
- **交互库**: Viem
- **测试**: Node.js 原生测试运行器
- **部署**: Hardhat Ignition
- **网络**: Sepolia 测试网

## 合约功能

### JacksonNFT 合约
- ✅ NFT 铸造功能
- ✅ 代币元数据管理
- ✅ 所有权转移
- ✅ 可升级架构

### Auction 合约
- ✅ 创建拍卖（设置时长和支付方式）
- ✅ ETH 出价功能
- ✅ ERC20 代币出价功能
- ✅ 自动退款（针对被超越的出价）
- ✅ 拍卖结束和 NFT 转移
- ✅ Chainlink 价格集成
- ✅ 重入攻击保护

## 已部署合约地址（Sepolia 测试网）

- **Auction 合约**: `0xeB17fe9B4484939e0CEa1d74da35Cc0c8636E3f8`
- **JacksonNFT 合约**: `0x602b759be5f08cA4FFb5bfae1D60AFa3Ed0bBb18`

## 环境准备

### 1. Node.js 版本要求
- 需要使用 Node.js v18.19+ 或 v20.6+ 版本
- 当前版本 v18.16.1 存在兼容性问题

### 2. 安装依赖
```bash
npm install
```

### 3. 配置私钥
设置 Sepolia 测试网私钥：
```bash
npx hardhat keystore set SEPOLIA_PRIVATE_KEY
```

设置 Etherscan API 密钥（用于合约验证）：
```bash
npx hardhat keystore set ETHERSCAN_API_KEY
```

### 4. 获取 Sepolia 测试币
访问以下水龙头获取免费测试 ETH：
- https://sepoliafaucet.com/
- https://www.infura.io/faucet/sepolia
- https://sepoliafaucet.net/

## 使用方法

### 编译合约
```bash
npx hardhat compile
```

### 运行测试
```bash
npx hardhat test
```

### 部署合约

#### 部署到本地网络
```bash
npx hardhat ignition deploy ./ignition/modules/Auction.ts --network hardhat
```

#### 部署到 Sepolia 测试网
```bash
npx hardhat ignition deploy ./ignition/modules/Auction.ts --network sepolia --verify
```

### 合约验证
验证已部署的合约：
```bash
npx hardhat verify --network sepolia 0xeB17fe9B4484939e0CEa1d74da35Cc0c8636E3f8
npx hardhat verify --network sepolia 0x602b759be5f08cA4FFb5bfae1D60AFa3Ed0bBb18
```

## 交互脚本

### NFT 拍卖演示脚本
运行完整的 NFT 拍卖流程演示：

```bash
npx hardhat run scripts/nft-bidding-demo.ts --network sepolia
```

该脚本将演示：
1. 铸造 NFT
2. 批准拍卖合约
3. 创建拍卖
4. 多个出价者竞价
5. 退款机制
6. 实时价格显示

## 项目结构

```
missionThree/
├── contracts/                 # 智能合约
│   ├── JacksonNFT.sol        # NFT 合约
│   ├── Auction.sol           # 拍卖合约
│   └── AuctionV2.sol         # 升级版本拍卖合约
├── ignition/                  # Ignition 部署模块
│   └── modules/
│       ├── JacksonNFT.ts     # NFT 部署模块
│       └── Auction.ts        # 拍卖系统部署模块
├── scripts/                  # 交互脚本
│   └── nft-bidding-demo.ts   # NFT 拍卖演示脚本
├── test/                     # 测试文件
├── hardhat.config.ts         # Hardhat 配置
├── package.json              # 项目依赖
└── README-zh.md             # 中文文档
```

## 核心功能说明

### 拍卖流程
1. **卖家**铸造 NFT 并批准给拍卖合约
2. **卖家**创建拍卖，设置时长和支付方式（ETH/ERC20）
3. **买家**使用 ETH 或 ERC20 代币参与竞价
4. **系统**自动退还被超越的出价
5. **拍卖结束后**，最高出价者获得 NFT，卖家收到资金

### 价格机制
- 使用 Chainlink 价格预言机实时获取 ETH/USD 汇率
- 支持 ETH 和 ERC20 代币出价的统一比较
- 自动进行价格转换和计算

### 安全特性
- 重入攻击保护
- 权限控制
- 可升级架构
- 安全的支付处理

## 注意事项

⚠️ **重要提醒**：
- 本项目仅在 Sepolia 测试网上运行
- 所有交易使用测试 ETH，不涉及真实资金
- 请确保私钥安全，不要在生产环境中使用测试私钥

## 故障排除

### 常见问题

1. **Node.js 版本不兼容**
   ```
   Error: This version of Node.js (v18.16.1) does not support module.register()
   ```
   **解决**: 升级到 Node.js v18.19+ 或 v20.6+

2. **私钥未设置**
   ```
   Error: No SEPOLIA_PRIVATE_KEY configured
   ```
   **解决**: 运行 `npx hardhat keystore set SEPOLIA_PRIVATE_KEY`

3. **余额不足**
   **解决**: 访问 Sepolia 水龙头获取测试 ETH

## 开发团队

本项目基于 Solidity 学习任务开发，展示了智能合约开发的最佳实践和高级功能。

## 许可证

MIT License