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
  |---------------|--------------------------|-------------|---------------|
  | Storage       | Global variables         | ✅ Yes       | Permanent     |
  | Memory        | Local variables, returns | ✅ Yes       | Function only |
  | Calldata      | External parameters      | ❌ No        | Function only |

  Key: Storage = permanent, Memory = temporary, Calldata = read-only temporary

    Quick Rules:

  | When to Use                 | Memory | Calldata    |
  |-----------------------------|--------|-------------|
  | External function parameter | ✅      | ✅ (cheaper) |
  | Internal function parameter | ✅      | ❌           |
  | Local variable              | ✅      | ❌           |
  | Need to modify data         | ✅      | ❌           |
  | Read-only, cheapest         | ❌      | ✅           |

  Bottom line: Use calldata for external parameters when you don't need to modify the data. Use memory when you need to modify or for 
  internal functions.

   Types Summary:

  | Type     | Category  | Needs Location?                 |
  |----------|-----------|---------------------------------|
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

## 4. Control Structures
- if-else statements
- for loops
- while loops
- break and continue

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



## 5. Object-Oriented Programming
### 5.1 Inheritance
- is keyword
- Constructors
- super keyword

### 5.2 Interfaces
- interface definition
- interface implementation

### 5.3 Abstract Contracts
- abstract keyword
- abstract functions

## 6. Events and Logs
- event definition
- emit statement
- event listening

## 7. Error Handling
- require statements
- assert statements
- revert statements
- custom errors

Example comparison

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

### 8.2 Design Patterns
- Factory pattern
- State machine pattern
- Proxy pattern

## 9. Ethereum Interaction
### 9.1 Accounts and Addresses
- Externally Owned Accounts vs Contract Accounts
- Address type operations

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

 **Best Practice Summary**

  When to Use Interfaces:

  - ✅ Known ABI/contract type
  - ✅ Type safety needed
  - ✅ Compile-time checking
  - ✅ Regular contract interactions

  When to Use Low-Level Calls:

  - ✅ Unknown/ dynamic ABI
  - ✅ Emergency fallbacks
  - ✅ Proxy patterns
  - ✅ Sending ETH ({value: amount})

**FallBack and Recevie**

1. receive ETH
2. handle function call not existed in contract. (**especially in proxy)


## 10. Gas Optimization
- Gas concept
- Optimization techniques
- Gas limits

## 11. Security Best Practices
- Reentrancy attack protection
- Integer overflow checks
- Access control
- Input validation

** digit signature of ETH **  
contract ulitilize function and its parameter to generate transaction 
hash. wallet produce signature via its private key and transction hash. then contract validate its signature, recovering its address (derived from public key), 
cheking if it equals to public key of expected address, if pass, then execute transation. 

## 12. Development Tools
### 12.1 Development Environment
- Remix IDE
- Hardhat
- Truffle

### 12.2 Testing Framework
- Unit testing
- Integration testing
- Testnet deployment

## 13. Deployment and Interaction
### 13.1 Testnet Deployment
- Ropsten, Rinkeby, Goerli, Sepolia
- Deployment scripts

### 13.2 Frontend Interaction
- Web3.js
- Ethers.js
- Event listening

## 14. Advanced Topics
### 14.1 Libraries
- Library definition and usage
- Inline assembly

### 14.2 Proxy Contracts
- Upgradeable contracts
- Proxy patterns

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

  | Feature     | approve()                   | setApprovalForAll()      |
  |-------------|-----------------------------|--------------------------|
  | Scope       | One token                   | All tokens               |
  | Storage     | mapping(uint256 => address) | mapping(address => bool) |
  | Gas         | Lower per approval          | Higher (sets for all)    |
  | Flexibility | Granular control            | Convenient bulk control  |

approve is only for one address, if you call it again, the first approval address will lose approval.
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



### 15.3 Complex Projects
- DAO governance contract
- Marketplace for collectibles
- DeFi protocols


