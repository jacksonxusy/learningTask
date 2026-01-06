# Solidity Basic Learning Outline

## 2. Basic Syntax

### 2.1 Contract Structure

- pragma declaration
- contract keyword
- Basic contract template

### 2.2 Data Types

- Value Types: uint, int, bool, address, bytes
- Reference Types: arrays, structs, mappings
- Special Types: string, bytes, enum
  In solidity, if a variable doesn't have a specified type.its type is inferred from the context, 
  defaulting to the smallest unit type, which is uint8 by default. In [uint(1),2,3], 
  all elements are of uint type because the first element is explicitly specified as uint type, 
  and the type of each element in the array follows the type of the first element.

### 2.3 Variable Declarations

Data location keywords (memory, storage, calldata) are only needed for function parameters and
local variables, NOT for global/state variables.
State Variables:

- Always stored in storage by default
- Storage location is implicit since they're part of the contract's permanent state
- No need to specify what's already determined

⏺ Data Location Summary

1. Storage
- Purpose: Permanent contract state

- Usage: Global variables, structs, arrays

- Lifetime: Exists forever until contract is destroyed
  
  contract Example {
    uint256[] public myArray;      // storage - global variable
  
    function add(uint256 value) external {
  
        myArray.push(value);       // storage - modifying global
  
    }
  
    function getArray() external view returns(uint256[] storage) {
  
        return myArray;            // storage reference
  
    }
  }
2. Memory
- Purpose: Temporary data during function execution

- Usage: Function parameters, local variables, return values

- Lifetime: Only exists during function execution
  
  function process(uint256[] memory data) external {
    // memory parameter
    uint256[] memory temp = new uint256[](10);  // memory local variable
    temp[0] = data[0];
  }
  
  function getString() external pure returns(string memory) {
    return "temporary string";    // memory return value
  }
3. Calldata
- Purpose: Read-only input for external functions

- Usage: External function parameters only

- Lifetime: Only during function execution

- Cannot be modified
  
  function externalFunction(uint256[] calldata input) external pure {
    // calldata parameter - read only
    // input[0] = 123;  // ❌ Cannot modify calldata
    uint256 first = input[0];  // ✅ Can read
  }
  
  Quick Rules:
  
  | Data Location | Use For                  | Can Modify? | Lifetime      |
  | ------------- | ------------------------ | ----------- | ------------- |
  | Storage       | Global variables         | ✅ Yes       | Permanent     |
  | Memory        | Local variables, returns | ✅ Yes       | Function only |
  | Calldata      | External parameters      | ❌ No        | Function only |
  
  Key: Storage = permanent, Memory = temporary, Calldata = read-only temporary
  
  Quick Rules:
  
  | When to Use                 | Memory | Calldata    |
  | --------------------------- | ------ | ----------- |
  | External function parameter | ✅      | ✅ (cheaper) |
  | Internal function parameter | ✅      | ❌           |
  | Local variable              | ✅      | ❌           |
  | Need to modify data         | ✅      | ❌           |
  | Read-only, cheapest         | ❌      | ✅           |
  
  Bottom line: Use calldata for external parameters when you don't need to modify the data. Use memory when you need to modify or for 
  internal functions.

**Summary: calldata for input (cheaper, immutable), memory for return (required for construction).**

   Types Summary:

| Type     | Category  | Needs Location?                 |
| -------- | --------- | ------------------------------- |
| address  | Value     | ❌ No                            |
| uint256  | Value     | ❌ No                            |
| bool     | Value     | ❌ No                            |
| string   | Reference | ✅ Yes (memory/storage/calldata) |
| bytes    | Reference | ✅ Yes                           |
| arrays   | Reference | ✅ Yes                           |
| structs  | Reference | ✅ Yes                           |
| mappings | Reference | ✅ Yes (storage only)            |

  Rule: Only reference types (string, bytes, arrays, structs, mappings) need location specifiers. Value types (address, uint, 
  bool) don't.

**note:** In Solidity, state variables cannot have the external visibility specifier.

Function Context:

- Need to explicitly tell Solidity where to put data for efficiency
- Different locations have different gas costs and use cases
- Compiler needs to know how to handle the data

Rule of thumb:

- Global variables = automatically storage
- Function variables = must specify memory, storage, or calldata

## 3. Functions

### 3.1 Function Definition

- Function declaration syntax
- Parameters and return values
- Function visibility: public, private, internal, external

### 3.2 Function Modifiers

- view functions
- pure functions
- payable functions

### 3.3 Built-in Functions

- keccak256, sha256

- require, assert, revert

- this, msg, block, tx
  
  ```solidity
  require(_owners[tokenId] == address(0), "token already minted");
  ```
  
  If condition is true, then go ahead. If no, then revert with error message.

**key difference between call and delegate call**

wallet address A => call contract B ==> call contract C.

conext: contract B.  
msg.sender = A. msg.value = A.
context: contract C.
msg.sender = B. msg.value = B;

wallet address A => call contract B ==> delegate call contract C.

conext: contract B.
msg.sender = A. msg.value = A.
context: contract C.
msg.sender = A. msg.value = A;

eg: in this code sninppet, `tokenOut.transfer(msg.sender, amountOut);`. 
its underlying code msg.sender === this contract  address, not wallet address, becuase this 
contract called him.

```
function swap(uint amountIn, IERC20 tokenIn, uint amountOutMin) external returns (uint amountOut, IERC20 tokenOut){
    require(amountIn > 0, 'INSUFFICIENT_OUTPUT_AMOUNT');
    require(tokenIn == token0 || tokenIn == token1, 'INVALID_TOKEN');

    uint balance0 = token0.balanceOf(address(this));
    uint balance1 = token1.balanceOf(address(this));

    if(tokenIn == token0){
        // 如果是token0交换token1
        tokenOut = token1;
        // 计算能交换出的token1数量
        amountOut = getAmountOut(amountIn, balance0, balance1);
        require(amountOut > amountOutMin, 'INSUFFICIENT_OUTPUT_AMOUNT');
        // 进行交换
        tokenIn.transferFrom(msg.sender, address(this), amountIn);
        tokenOut.transfer(msg.sender, amountOut);
```

** ERC20 when to use transfer function or transferFrom function **
it depends your current entity wants to do what action.

- if you wants to take token from others, use transferFrom.
- if you wants to transfer your token to others, use transfer.  
  in above code swap function, current contract wants to token from user, and transfer tokenOut to user.

** difference between internal call and external call **

```
  function secretSlector() public   pure returns(bytes4){
      return bytes4(keccak256("putCurEpochConPubKeyBytes(bytes)"));
  }

  function hackSlector() external  returns(bytes4){
      secretSlector(); // internal call, inside secretSlector msg.sender = your wallet address.
      this.secretSlector(); // external call, inside secretSlector msg.sender = current contract address.
      return bytes4(keccak256("f1121318093(bytes,bytes,uint64)"));
  } 
```

## 5. Object-Oriented Programming

### 5.1 Inheritance

- is keyword
- Constructors
- super keyword

### 5.2 Interfaces

- interface definition
- interface implementation

### abi.encode

abi was designed to interact with smart contract, it pad into 32 bytes with each parameters, and concat together, if you 
want to interact with contract, you need use abi.encode.

```
function encode() public view returns(bytes memory result) {
    result = abi.encode(x, addr, name, array);
}
```

the results is:

```
000000000000000000000000000000000000000000000000000000000000000a    // x
0000000000000000000000007a58c0be72be218b41c608b7fe7c5bb630736c71    // addr
00000000000000000000000000000000000000000000000000000000000000a0    // name 参数的偏移量
0000000000000000000000000000000000000000000000000000000000000005    // array[0]
0000000000000000000000000000000000000000000000000000000000000006    // array[1]
0000000000000000000000000000000000000000000000000000000000000004    // name 参数的长度为4字节
3078414100000000000000000000000000000000000000000000000000000000    // name
```

**abi.encodePacked**
it will encode data according to minimum space that it need, similar to abi.encode, but it will ignore many zero, still above
encode() function, the result is:

```
0x000000000000000000000000000000000000000000000000000000000000000a7a58c0be72be218b41c608b7fe7c5bb630736c713078414100000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000006
```

because it will not pad into zero, different input may have the same encode result after concating, leading to collision.
`abi.encodePacked("ab", "c") == abi.encodePacked("a", "bc")`. 

### why we need hash fields inside abi.encode?

EIP-712 encode:

```
bytes32 domainSeparator = keccak256(
    abi.encode(
        DOMAIN_TYPEHASH,
        keccak256(bytes(name())),
        block.chainid,
        address(this)
    )
);
```

1. inner `keccak256(bytes(name()))` converts a dynamic-length string into fixed-length, so it can safely be encoded in the domain struct.
2. outer `keccak256(abi.encode(...))` hash the entire encoded struct into a single bytes32. becuase eth wallet can't sign the 
   structs directly, they can only sign a 32 bytes hash.
- abi.encode(string) works fine
- problem arise only if need fixed-32bytes representation for: EIP-712 / mapping keys
- in that case keccak256(bytes(string)) is required.

| Method                                                                      | Total encoded bytes (approx) | Notes                     |
| --------------------------------------------------------------------------- | ---------------------------- | ------------------------- |
| `abi.encode(string)`                                                        | 32 + 32 + padded length      | Grows with string size    |
| `abi.encode(keccak256(string))`                                             | 32                           | Always fixed, saves space |
| **Hashing saves storage and gas when you need a fixed-size representation** |                              |                           |

## 7. Error Handling

- require statements
- assert statements
- revert statements
- custom errors

**unchecked keyword**
only disable automatic overflow/underflow checks for arithmetic operations.
not for division by zero errors/ out of bonds array access error.

normal case:

```
uint8 x = 255;
x = x + 1;  // ❌ will revert (overflow)
```

but if you add unchecked, no overflow/underflow error more.

```
unchecked {
    x = x + 1;   // ✔ no revert, wraps to 0
}
```

why use it?

1. gas optimization, if you are sure the range is safe or overflow behavior is intentional, you can
   remove it to save gas.

example in ERC20:

```
function _spendAllowance(address owner, address spender, uint256 amount) internal virtual {
    uint256 currentAllowance = allowance(owner, spender);
    if (currentAllowance != type(uint256).max) {
        require(currentAllowance >= amount, "ERC20: insufficient allowance");

        unchecked {
            _approve(owner, spender, currentAllowance - amount);
        }
    }
}
```

**revert Example comparison**

Solidity:

```solidity
function divide(uint256 a, uint256 b) public pure returns (uint256) {
    if (b == 0) {
        revert("Division by zero");
    }
    return a / b;
}
```

Java:

```java
int divide(int a, int b) {
    if (b == 0) {
        throw new IllegalArgumentException("Division by zero");
    }
    return a / b;
}
```

Same logic — the main difference is Solidity’s revert undoes all state changes and costs gas, while Java’s throw just stops execution.

## 8. Common Patterns

### 8.1 Security Patterns

- Checks-Effects-Interactions pattern
- Reentrancy protection
- Integer overflow protection

### 9.2 Ether Transfers

- transfer method
- send method
- call method

### 9.3 Calling Other Contracts

- Contract instantiation
- Low-level calls
  withdraw function.
  Without Data (Simple ETH Transfer):

```solidity
// Send ETH with no additional data
(bool success, ) = msg.sender.call{value: 1 ether}("");
//                                                 ↑
//                                            No function call data
```

  With Data (Function Call):

```solidity
// Send ETH AND call a function
(bool success, ) = targetContract.call{value: 1 ether}(
    abi.encodeWithSignature("deposit()")
);
//                                                    ↑
//                                            Function call data
```

**two ways to call other contract**
1 use call to other contract.
format: targetContractAddress.call(bytecode).
bytecode can be acquired by abi.encodeWithSignature("function signature", concrete parameter spilted by comma);
eg: abi.encodeWithSignature("f(uint256,address)", _x, _addr)。
and we can send eth and gas to function:
'targetContractAddress.call{value:amount, gas: gas amount}(bytecode).'

2 use Interfaces:
eg,ERC20 contract can be called like this: ERC20(token).

```solidity
contract SafeCaller {
    IERC20 public token;
    IVault public vault;

    constructor(address _token, address _vault) {
        token = IERC20(_token);
        vault = IVault(_vault);
    }

    function safeTransfer(address to, uint256 amount) external {
        // Type-safe, compile-time checked calls
        bool success = token.transfer(to, amount);
        require(success, "Transfer failed");
    }

    function safeDeposit(uint256 amount) external {
        vault.deposit(amount);
    }
}
```

**FallBack and Recevie**

1. receive ETH
2. handle function call not existed in contract. (**especially in proxy**)

**using X for Y**
| Statement                     | Meaning                                  | Example                          |
| ----------------------------- | ---------------------------------------- | -------------------------------- |
| `using SafeERC20 for IERC20;` | Attach SafeERC20 methods to ERC20 tokens | `token.safeTransfer(to, amount)` |
| `using Address for address;`  | Attach Address helpers to address type   | `addr.isContract()`              |
| `using Math for uint256;`     | Attach Math helpers to uint256 values    | `value.mulDiv(x,y)`              |

## 10. Gas Optimization

- Gas concept
- Optimization techniques
- Gas limits

**digit signature of ETH**  
contract ulitilize function and its parameter to generate transaction 
hash. wallet produce signature via its private key and transction hash. then contract validate its signature, recovering its address (derived from public key), 
cheking if it equals to public key of expected address, if pass, then execute transation. 

### 14.3 Multi-Signature Wallets

- Multi-sig contract principles
- Permission management

## 15. Practical Projects

### 15.1 Simple Projects

- Storage contract

- Voting contract

- Token contract (ERC20)
  ERC20 approval and transferFrom process. 
  The Approval Process:
  
  Step 1: User First Approves the Airdrop Contract

```solidity
// User calls the ERC20 token contract directly:
tokenContract.approve(address(airdropContract), 1000);
```

```solidity
// This sets:
allowance["user_address"]["airdrop_contract_address"] = 1000;
// Meaning: Airdrop contract can spend 1000 of user's tokens
```

  Step 2: User Calls the Airdrop Function

```solidity
// User calls the airdrop contract:
airdropContract.multiTransferToken(
    tokenContract,           // The ERC20 token
    [addr1, addr2, addr3],   // Recipients
    [100, 200, 300]         // Amounts
);
```

  Step 3: Airdrop Contract Uses the Allowance

  // Inside multiTransferToken():
  for (uint256 i; i < _addresses.length; i++) {
      token.transferFrom(msg.sender, _addresses[i], _amounts[i]);
      //                 ↑ sender    ↑ recipient  ↑ amount
  }

  // This calls the ERC20 token's transferFrom function:
  // allowance[user][airdrop_contract] -= amount;
  // balanceOf[user] -= amount;
  // balanceOf[recipient] += amount;

  The Allowance Mapping Structure:

```solidity
mapping(address => mapping(address => uint256)) public allowance;
//                 ↑         ↑             ↑
//              owner    spender    amount allowed
```

  What This Means:

```solidity
allowance["0x123..."]["0x456..."] = 1000;  // 0x456 can spend 1000 of 0x123's tokens
allowance["0x123..."]["0x789..."] = 500;   // 0x789 can spend 500 of 0x123's tokens
allowance["0xABC..."]["0x456..."] = 200;   // 0x456 can spend 200 of 0xABC's tokens
```

  Visual Flow:

```text
User (0x123)                Airdrop Contract (0x456)            Token Contract
     | approve(1000)  ──────────────────→                         |
     |                                         transferFrom(0x123, 0x789, 100) ──→
     |                                                              │
     |                                                              │ Transfer tokens
     |                                                              │
     |                                                          Token Transferred!
```

  Why This Design:

  Security:

- User never loses control of tokens

- Can revoke approval anytime

- Limit the amount each contract can spend
  
  Flexibility:

- One approval = multiple transfers

- Can approve different amounts for different contracts
  
  Example Usage:
  
  // User approves multiple contracts:
  token.approve(uniswap, 1000);    // Uniswap can spend 1000
  token.approve(airdrop, 500);     // Airdrop can spend 500
  token.approve(lending, 200);     // Lending can spend 200
  
  // Each contract checks its own allowance:
  require(allowance[user][msg.sender] >= amount, "Not approved!");

ERC721
approve and setApproval difference:
  Key Differences:

| Feature     | approve()                   | setApprovalForAll()                                                      |
| ----------- | --------------------------- | ------------------------------------------------------------------------ |
| Scope       | One token                   | All tokens                                                               |
| Storage     | mapping(uint256 => address) | mapping(address => mapping(address => bool)) private _operatorApprovals; |
| Gas         | Lower per approval          | Higher (sets for all)                                                    |
| Flexibility | Granular control            | Convenient bulk control                                                  |

- approve is only for one address, if you call it again, the first approval address will lose approval.
- setApprovalForAll allows an NFT owner to grant permission to another address to manage ALL of their NFTs from a specific collection.
   It's a powerful permission mechanism.

```
// Example Usage Contract
  contract NFTUser {
      IERC721 public nftContract;
      NFTMarketplace public marketplace;

      constructor(address _nftContract, address _marketplace) {
          nftContract = IERC721(_nftContract);
          marketplace = NFTMarketplace(_marketplace);
      }

      // Step 1: Approve marketplace to manage ALL your NFTs
      function approveMarketplace() external {
          nftContract.setApprovalForAll(address(marketplace), true);
          console.log("Marketplace approved to manage all NFTs");
      }

      // Step 2: List multiple NFTs without individual approvals
      function listMultipleNFTs(uint256[] tokenIds, uint256[] prices) external {
          for (uint i = 0; i < tokenIds.length; i++) {
              marketplace.listItem(tokenIds[i], prices[i]);
              console.log("NFT", tokenIds[i], "listed for", prices[i], "wei");
          }
      }

      // Step 3: Revoke marketplace approval when done
      function revokeMarketplaceApproval() external {
          nftContract.setApprovalForAll(address(marketplace), false);
          console.log("Marketplace approval revoked");
      }

      // Check if marketplace is approved
      function checkApprovalStatus() external view returns (bool) {
          return nftContract.isApprovedForAll(address(this), address(marketplace));
      }
  }


# Example marketNFT
  contract NFTMarketplace {
      // NFT contract address
      IERC721 public nftContract;

      struct Listing {
          address seller;
          uint256 tokenId;
          uint256 price;
      }

      mapping(uint256 => Listing) public listings;

      event Listed(uint256 indexed tokenId, uint256 price);
      event Sold(uint256 indexed tokenId, address buyer);

      constructor(address _nftContract) {
          nftContract = IERC721(_nftContract);
      }

      // Seller lists their NFT for sale
      // REQUIRES: seller must have called setApprovalForAll(marketplace, true)
      function listItem(uint256 tokenId, uint256 price) external {
          // Check that marketplace is approved to manage seller's NFTs
          require(
              nftContract.isApprovedForAll(msg.sender, address(this)),
              "Marketplace not approved for all NFTs"
          );

          // Verify caller owns the NFT
          require(
              nftContract.ownerOf(tokenId) == msg.sender,
              "Not token owner"
          );

          listings[tokenId] = Listing({
              seller: msg.sender,
              tokenId: tokenId,
              price: price
          });

          emit Listed(tokenId, price);
      }

      // Buyer purchases listed NFT
      function buyItem(uint256 tokenId) external payable {
          Listing memory listing = listings[tokenId];
          require(listing.price > 0, "Item not for sale");
          require(msg.value >= listing.price, "Insufficient payment");

          // Transfer NFT from seller to buyer
          nftContract.transferFrom(listing.seller, msg.sender, tokenId);

          // Pay seller (minus marketplace fee)
          payable(listing.seller).transfer(listing.price);

          // Refund excess payment
          if (msg.value > listing.price) {
              payable(msg.sender).transfer(msg.value - listing.price);
          }

          delete listings[tokenId];
          emit Sold(tokenId, msg.sender);
      }
  }
```

**ERC721 transfer function**.
in ERC721 function, it has an approve function.

```solidity
function _transfer(address owner, address from, address to, uint256 tokenId) private {
    require(from == owner, "not owner");
    require(to != address(0), "transfer to the zero address");
    _approve(owner, address(0), tokenId);
}
```

_approve(owner, address(0), tokenId); this is to set tokenId's approval to address(0), clearing the approval when token is transferred.
  Why this is necessary:

1. Security: When you transfer a token, any existing approvals should be canceled
2. Prevent double-spending: The old owner shouldn't be able to transfer it again after selling
3. State consistency: The new owner starts with a "clean slate" - no approvals exist

**ERC721 transfer checkReceive function.**

```solidity
function _checkOnERC721Received(address from, address to, uint256 tokenId, bytes memory data) private {
    if (to.code.length > 0) {
        try IERC721Receiver(to).onERC721Received(msg.sender, from, tokenId, data) returns (bytes4 retval) {
            if (retval != IERC721Receiver.onERC721Received.selector) {
                revert ERC721InvalidReceiver(to);
            }
        } catch (bytes memory reason) {
            if (reason.length == 0) {
                revert ERC721InvalidReceiver(to);
            } else {
                /// @solidity memory-safe-assembly
                assembly {
                    revert(add(32, reason), mload(reason))
                }
            }
        }
    }
}
```

breaking down:

1. check if recipient is a contract:
   if (to.code.length > 0) {  // to.code.length > 0 = is contract
      // If code length > 0, it's a contract address
      // If code length == 0, it's an EOA (normal wallet)
   }
2. call the contract's receiver function:
   it ensure the receiver implement IERC721Receiver menthod and returned function Signature magic number.
   otherwise, it will thorw an error.
   it just check if the contract claims to support ERC721, not verify if the contract handles NFT correctly.

### 15.2 Intermediate Projects

- NFT contract (ERC721)
- Decentralized exchange
- Lending protocol

**ERC4626 break down**

```
function deposit(uint256 assets, address receiver) public virtual returns (uint256 shares) {

    shares = previewDeposit(assets);

    _asset.transferFrom(msg.sender, address(this), assets);
    _mint(receiver, shares);


    emit Deposit(msg.sender, receiver, assets, shares);
}
```

1. calculates how many vault shares the user will receive.

2. user's asset transferred to valut contract.

3. valut creates new shares for the receiver.
   Correct Order:
   
   1. Calculate ✅ (Check)
   2. Transfer assets ✅ (Interaction 1)
   3. Mint shares ✅ (Effect)
   4. Emit event ✅ (Logging)
   
   Why This Order Matters:

```
  // VULNERABLE (if order was wrong):
  _mint(receiver, shares);           // First mint
  _asset.transferFrom(...);           // Then transfer

  // Attacker could re-enter during transfer and:
  // - Call deposit again before first deposit completes
  // - Get extra shares for same assets
```

deployed two contract(ERC20 and ERC4626), and execute the following actions.  

1. ERC20 contract, mint 10000 tokens. and approve ERC4626 contract address and 10000 token.
2. ERC4626 contract deposit 1000 token to wallet address.

balanceOf and totoalSupply amount change like below:

Before Deposit:
// ERC20 contract:

```
tokenA.balanceOf(myWallet) = 10000.
totoalSupply = 10000.
```

After 1000 Token A Deposit:  
`_asset.transferFrom(msg.sender, address(this), assets);` 
this transfer 1000 token to contract B. then it changed to this:  

```
// ERC20 contract:
tokenA.balanceOf(myWallet) = 9000.
tokenA.balanceOf(valut contract) = 1000.
totoalSupply = 10000.
```

then called ` _mint(receiver, shares);` it changed to this:

```
// ERC4626 contract:
tokenB.balanceOf(myWallet) = 1000.
totoalSupply = 1000;
```

**note: token flow work between the two contracts**

1. checker who is caller.  
   `_asset.transferFrom(msg.sender, address(this), assets);` this caller is _asset.  
   `_mint(receiver, shares);` this caller is current contract.
2. using current caller's state variable to calculate.  
   `_asset.transferFrom(msg.sender, address(this), assets);` this use _asset balance mapping to calculate.  
   `_mint(receiver, shares);` this use current contract balance mapping to calcalate.

**Dex getAmounOut break down**

```
function getAmountOut(uint amountIn, uint reserveIn, uint reserveOut) public pure returns(uint amountOut) {
    require (amountIn > 0, "insufficient input");
    require (reserveIn > 0 && reserveOut > 0, "insufficient liquidity");
    amountOut = amountIn * reserveOut / (reserveIn + amountIn);
}
```

  Core Formula:

  amountOut = amountIn * reserveOut / (reserveIn + amountIn)

  AMM Pool Mechanics:

```
  // Before swap:
  reserveIn = 1000   // Pool has 1000 input tokens
  reserveOut = 2000  // Pool has 2000 output tokens

  // You want to swap 100 input tokens:
  amountIn = 100

  // New pool state after swap:
  newReserveIn = 1000 + 100 = 1100
  newReserveOut = 2000 - amountOut = ?

  // Constant product formula: k = reserveIn * reserveOut
  // 1000 * 2000 = 1100 * newReserveOut
  // newReserveOut = (1000 * 2000) / 1100 = 1818
  // amountOut = 2000 - 1818 = 182
```

They're actually the same mathematically:

// Full formula:
`amountOut = reserveOut - (reserveIn * reserveOut) / (reserveIn + amountIn)`

// Simplified formula:
`amountOut = amountIn * reserveOut / (reserveIn + amountIn)`

** funciton Purpose difference between mint and deposit **. 
Action    User Chooses    Vault Calculates
deposit()    assets    shares
mint()    shares    assets

⏺ **EVM Atomicity Summary**

  What is EVM Atomicity?

  EVM transaction atomicity ensures that every smart contract transaction executes as an "all-or-nothing" operation. Either all state
  changes commit to the blockchain, or none do.

  Key Principles

1. Atomic Execution: Transactions cannot partially succeed
2. State Consistency: Blockchain state never becomes partially updated
3. Rollback on Failure: Any exception/trigger causes complete rollback
4. Gas Consumption: Gas is consumed regardless of success/failure

**upgrade contract can be implemented via two points**

1. delegatecall -- logic contract can use proxy contract context, then no need to concern data change each time upggrading logic contract.
2. fallback function -- when user calls a function that doesn't exist in proxy contract, it will route to fallback method. In fallback method, we 
   can call logict contract to route to real implementation.

**upgrade contract conflict**

1. Type 1 – Logic contract overwrites Proxy’s own slots
   (The Akropolis-style bug)
   The proxy stores its critical data (implementation address, admin, etc.) in low slots (0, 1, …).
   If the logic contract declares a state variable in the same slot, delegatecall will overwrite the proxy’s data → contract dies instantly.
2. Type 2 – New version layout incompatible with old version layout
   Storage on chain never moves. If a variable changes its slot number between V1 → V2, the new code will read the wrong data (classic example: owner() suddenly returns 500 instead of an address).

example one:

```
// V1 (already live with real user data)
contract MyTokenV1 is Initializable, ERC20Upgradeable {
    uint256 public a;        // slot 0 → 1000
    uint256 public b;        // slot 1 → 500
    bool    public paused;   // slot 2 → true
    address public owner;    // slot 3 → 0x1234…abcd
}

// V2 (you only wanted to “clean up” the code order)
contract MyTokenV2 is Initializable, ERC20Upgradeable {
    uint256 public a;        // slot 0 → still 1000  (correct)
    address public owner;    // slot 1 → reads old b = 500 !!!  ← broken
    bool    public paused;   // slot 2 → reads old paused = true (maybe ok)
    uint256 public b;        // slot 3 → reads old owner address as uint256 !!!  ← broken
}
```

slot was assigned by vairable order, now get owner value, but get previous b value, so it is get conflicted.

example two:

```
Storage Slot,Who “owns” this slot in reality?,What is actually stored there?,Which contract wrote it?
slot 0,Proxy contract,implementation address (or admin),Proxy wrote it
slot 1,Proxy contract,owner / beacon / etc.,Proxy wrote it
slot 2,Proxy contract,(usually empty),—
…,…,…,—
slot 100,Logic contract (V1),totalSupply,Logic wrote via delegatecall
slot 101,Logic contract (V1),balances mapping base,Logic wrote via delegatecall
slot 102,Logic contract (V1),paused (bool),Logic wrote via delegatecall
```

proxy contract and logic contract use one single storage. logic contract variable will write to proxy storage using delegatecall.  
so if you just define variable by order in logic contract, it has great possibility to conflict with variable slot in proxy contract.

**solution**

using OpenZeppelin UUPS / Transparent proxy.
it does't store the variable in normal storage slots 0 ,1, 2, 3. instead, they store it in extremely high, "random-looking" slots
that was calculated with keccak256. this has very little chance to conflict.

**OpenZeppelin initialize function usage**

```
function initialize(address initialOwner) public initializer {
    __Ownable_init(initialOwner);
}
```

Reason we cannot use a normal constructor in the logic contract

- When the logic contract is deployed, its constructor runs immediately.

- After deployment, nobody ever uses the logic contract address directly → all that data is lost forever.

- Users only interact with the proxy address.

- The proxy uses delegatecall → the logic code runs, but storage is read/written in the proxy’s storage.

- Since the constructor never ran on the proxy, the proxy’s storage is empty → owner = address(0), totalSupply = 0 → everything broken from day one.
  if call logic contract constructor function at deployment, it will write related variable to logic address storage.
  when we call logic contract function in proxy contract, it can't find correct varibale in proxy contract address storage.
  but if you call initialize function in proxy contract, that set set related variable in proxy contract address.
1. __AccessControl_init() is empty because AccessControl has no constructor logic to replace.
   It still exists to keep initializer ordering consistent during multiple inheritance.

2 . __UUPSUpgradeable_init() is empty because UUPS has no constructor state that needs initialization.
The real logic is inside the inherited upgrade functions, not in the initializer.

```
contract MyToken is Initializable, ERC20Upgradeable, OwnableUpgradeable {
    // NO constructor!!!

    function initialize(string memory name, uint256 initialSupply) public initializer {
        __ERC20_init(name, "MTK");
        __Ownable_init(msg.sender);
        _mint(msg.sender, initialSupply);
    }
}
```

**Openzepplin UUPS and transparent proxy**

```
你只写了两个合约：
┌─────────────────┐          ┌─────────────────┐
│     Box.sol     │          │    BoxV2.sol    │
│ (实现合约 V1)   │          │ (实现合约 V2)   │
└─────────────────┘          └─────────────────┘

当你运行这行代码：
const box = await upgrades.deployProxy(Box, [42], { kind: "uups" });

OpenZeppelin 自动干了 3 件事（你完全不用管）：

        ┌──────────────────────────────────────┐
        │ 1. 部署 Box 实现合约（有代码）          │
        │ 2. 部署一个极小的代理合约（你看不到）   │←←←←←← 你问的 proxy 就在这！
        │ 3. 把代理合约的 implementation 指向 Box │
        └──────────────────────────────────────┘

最终用户只看到一个地址（比如 0x123...abc），这就是代理地址！

// UUPS 必须这样写 —— 空函数体 + onlyOwner
function _authorizeUpgrade(address newImplementation) 
    internal 
    override 
    onlyOwner 
{
    // 故意什么都不写！升级逻辑由 OpenZeppelin 插件完成
    // 你只负责回答一个问题：“谁允许升级？” → onlyOwner
}
```
