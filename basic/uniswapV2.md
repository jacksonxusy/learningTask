### ZuniswapV2Pair  swap function

 What swap(amount0Out, amount1Out, to, "") actually does:

- amount0Out: Amount of token0 to send OUT of the pair
- amount1Out: Amount of token1 to send OUT of the pair
- to: Recipient of the output tokens

1 these two variable respresent previous reserves from last operation, it is the pool liqidity.

```soli
uint112 private reserve0;
uint112 private reserve1;
```

2. **Why NOT just reserve0_ - amount0Out** is because:
- user deposits before calling mint/swwap, balance increses but reserves haven't updated.
- external transfers - someone could send token directly to the contract.

```solidity
uint256 balance0 = IERC20(token0).balanceOf(address(this)) - amount0Out;
uint256 balance1 = IERC20(token1).balanceOf(address(this)) - amount1Out;

if (balance0 * balance1 < uint256(reserve0_) * uint256(reserve1_))
    revert InvalidK(); //← Ensures k = x*y stays constant
```

3. **why add nonReentrant**
   Checks Effects Interactions pattern is one way of preventing the attack. However, in the rewritten swap function, we cannot use the pattern because the implementation forces us to make external calls (token transfers) before applying effects (updating reserves).

the lock midifier is specially required for the flash swap mechanism. 

1. give output tokens
2. call attacker/recipient contract(reentrancy point)
3. receive input tokens + fee(repayment)
4. check K invariant.
   Without the lock, an attacker could perform a nested, malicious action during the callback that:

**Attacker Action during Callback**: Calls swap again.

**Result**: The second, nested swap might proceed using an outdated or inconsistent reserve state, leading to a financial exploit before the initial call has confirmed that the first loan (Step 1) has been fully repaid and the invariant check (Step 4) has passed.

```solidity
function swap(
        uint256 amount0Out,
        uint256 amount1Out,
        address to,
        bytes calldata data
    ) public nonReentrant {
        if (amount0Out == 0 && amount1Out == 0)
            revert InsufficientOutputAmount();

        (uint112 reserve0_, uint112 reserve1_, ) = getReserves();

        if (amount0Out > reserve0_ || amount1Out > reserve1_)
            revert InsufficientLiquidity();

        if (amount0Out > 0) _safeTransfer(token0, to, amount0Out);
        if (amount1Out > 0) _safeTransfer(token1, to, amount1Out);
        if (data.length > 0)
            IZuniswapV2Callee(to).zuniswapV2Call(
                msg.sender,
                amount0Out,
                amount1Out,
                data
            );

        uint256 balance0 = IERC20(token0).balanceOf(address(this));
        uint256 balance1 = IERC20(token1).balanceOf(address(this));

        uint256 amount0In = balance0 > reserve0 - amount0Out
            ? balance0 - (reserve0 - amount0Out)
            : 0;
        uint256 amount1In = balance1 > reserve1 - amount1Out
            ? balance1 - (reserve1 - amount1Out)
            : 0;

        if (amount0In == 0 && amount1In == 0) revert InsufficientInputAmount();

        // Adjusted = balance before swap - swap fee; fee stays in the contract
        uint256 balance0Adjusted = (balance0 * 1000) - (amount0In * 3);
        uint256 balance1Adjusted = (balance1 * 1000) - (amount1In * 3);

        if (
            balance0Adjusted * balance1Adjusted <
            uint256(reserve0_) * uint256(reserve1_) * (1000**2)
        ) revert InvalidK();

        _update(balance0, balance1, reserve0_, reserve1_);

        emit Swap(msg.sender, amount0Out, amount1Out, to);
    }
```

### TWAP memchanism. (time weighted average price).

in uniswap v1, it use spot price(the instantaneous reserve ratio reserve0/reserve1), an attaker could use
a flash loan to execute a huge trade within a single block, temporarily shifting the price to an arbitrary value.
**exteral protocols(like lending platform) that naively read this spot price would use the wrong valuation, leading to
potential exploits**.

in uniswap v2.  
New Cumulative Price = Old Cumulative Price + Instantaneous Price * Seconds Elapsed.  
This variable thus represents the sum of all time-weighted historical prices since the contract was deployed.

**Conclusion**
The TWAP's security guarantee is that it protects external systems from the manipulation.

The attacker is still able to execute a trade at the poor spot price of 1,500 DAI/ETH. However, they cannot use that momentary price to steal money from a protected lending vault, which is the entire point of the Flash Loan attack. The attacker must still reverse the trade and repay the Flash Loan, and their profit opportunity is blocked by the external protocol using the stable TWAP.

```solidity
function _update(
        uint256 balance0,
        uint256 balance1,
        uint112 reserve0_,
        uint112 reserve1_
    ) private {
        if (balance0 > type(uint112).max || balance1 > type(uint112).max)
            revert BalanceOverflow();

        unchecked {
            uint32 timeElapsed = uint32(block.timestamp) - blockTimestampLast;

            if (timeElapsed > 0 && reserve0_ > 0 && reserve1_ > 0) {
                price0CumulativeLast +=
                    uint256(UQ112x112.encode(reserve1_).uqdiv(reserve0_)) *
                    timeElapsed;
                price1CumulativeLast +=
                    uint256(UQ112x112.encode(reserve0_).uqdiv(reserve1_)) *
                    timeElapsed;
            }
        }

        reserve0 = uint112(balance0);
        reserve1 = uint112(balance1);
        blockTimestampLast = uint32(block.timestamp);

        emit Sync(reserve0, reserve1);
    }
```

### gas optimization

```solidity
address public token0;
address public token1;

uint112 private reserve0;
uint112 private reserve1;
uint32 private blockTimestampLast;

uint256 public price0CumulativeLast;
uint256 public price1CumulativeLast;
```

- SSTORE(saving value to contract storage) and SLOAD are expensive, consuming a lot of gas.
- evm use 32-byte storage slots to read and write. each SSTORE and SLOAD use 32 bytes at a time.  
- First two are address variables. address takes 20 bytes, and two addresses take 40 bytes, which means they have to take separate storage slots. They cannot be stored in one slot since they simply won’t fit.
- Two uint112 variables and one uint32–this looks interesting: 112+112+32=256! This means they can fit in one storage slot! This is why uint112 was chosen for reserves: the reserves variables are always read together, and it’s better to load them from storage at once, not separately. This saves one SLOAD operation, and since reserves are used very often, this is huge gas saving.
- Two uint256 variables. These cannot be packed because each of them takes a full slot.
  It’s also important that the two uint112 variables go after a variable that takes a full slot–this ensures that the first of them won’t be packed in the previous slot.

Slot,   Content,Gas                     Efficiency
Slot 0, "reserve0, reserve1,  blockTimestampLast",Most Efficient (Perfectly packed 3 variables into 32 bytes).
Slot 1, token0,                "Inefficient  (20 bytes used, 12 bytes wasted)."
Slot 2, token1,             "Inefficient (20 bytes used, 12 bytes wasted)."
Slot 3, price0CumulativeLast,     Neutral (Full 32-byte variable).
Slot 4, price1CumulativeLast,       Neutral (Full 32-byte variable).

### safe transfer

```solidity
function _safeTransfer(
        address token,
        address to,
        uint256 value
    ) private {
        (bool success, bytes memory data) = token.call(
            abi.encodeWithSignature("transfer(address,uint256)", to, value)
        );
        if (!success || (data.length != 0 && !abi.decode(data, (bool))))
            revert TransferFailed();
    }
```

In the pair contract, when doing token transfers, we always want to be sure that they’re successful. According to ERC20, transfer method must return a boolean value: true, when it’s successful; fails, when it’s not. Most of tokens implement this correctly, but some tokens don’t–they simply return nothing. Of course, we cannot check token contract implementation and cannot be sure that token transfer was in fact made, but we at least can check transfer result. And we don’t want to continue if a transfer has failed.

### factory creatingPair

1. using create2 opcode to create pair address.
   
   ```solidity
   // loads the contract creation bytecode for ZuniswapV2Pair. the exact bytes to deploy that contract.
   bytes memory bytecode = type(ZuniswapV2Pair).creationCode; 
   // choose a salt derived from the ordered token pair.
   bytes32 salt = keccak256(abi.encodePacked(token0, token1));
   assembly {
    pair := create2(0, add(bytecode, 32), mload(bytecode), salt)
   }
   // bytecode (a bytes variable) points to the start of that block (the length word). 
   // The actual code bytes start 32 bytes after that.
   // so add 32 bytes to get the right bytes then read it via mload(bytecode)
   ```

```solidity
// Create a new address deterministically using bytecode + salt.
// Deploy a new ZuniswapV2Pair contract.
// Get that pair's address.

// then initialize it with related token pair. **smaller address assign as token0, larger address assign as token1**
```

```solidity
IZuniswapV2Pair(pair).initialize(token0, token1);

pairs[token0][token1] = pair;
pairs[token1][token0] = pair;
allPairs.push(pair);

// ZuniswapV2Pair.sol
function initialize(address token0_, address token1_) public {
  if (token0 != address(0) || token1 != address(0))
    revert AlreadyInitialized();

  token0 = token0_;
  token1 = token1_;
}
```

### router addLiquidity

### router_swap function

```solidity
(uint256 amount0Out, uint256 amount1Out) = input == token0
    ? (uint256(0), amountOut)
    : (amountOut, uint256(0));
```

- determins which token is being swapped out based on whether the input token is token0 or token1.
- paramters in path logic order, eg: [A, B, C] A ->B -> C how user wants to swap.
- pair contract: stores tokens in ascending address order(token0 < token1)
- swap function also use pair contract order amount0 -> token0, amount1 -> token1.

```solidity
function _swap(
        uint256[] memory amounts,
        address[] memory path,
        address to_
    ) internal {
        for (uint256 i; i < path.length - 1; i++) {
            (address input, address output) = (path[i], path[i + 1]);
            (address token0, ) = ZuniswapV2Library.sortTokens(input, output);
            uint256 amountOut = amounts[i + 1];
            (uint256 amount0Out, uint256 amount1Out) = input == token0
                ? (uint256(0), amountOut)
                : (amountOut, uint256(0));
            // calcute next pair address.
            address to = i < path.length - 2
                ? ZuniswapV2Library.pairFor(
                    address(factory),
                    output,
                    path[i + 2]
                )
                : to_;
            IZuniswapV2Pair(
                ZuniswapV2Library.pairFor(address(factory), input, output)
            ).swap(amount0Out, amount1Out, to, "");
        }
    }
```

### library contract

#### pair for function.

the function is used to find pair address by factory and token addresses. the straightforward way of doing that
is by fetching pair address from the factory contract, like:

```solidity
ZuniswapV2Factory(factoryAddress).pairs(address(token0), address(token1))
```

**but this would make an external call, which makes the function a little more expenseive.**
uniswap uses are more advanced approach, and this is where we get a benefit from the deterministic address.
this piece of code generates an address in the same way create2 does.

**this will sort token, so no matter what order in paramter, it will generate the same pair address**

```solidity
function pairFor(
        address factoryAddress,
        address tokenA,
        address tokenB
    ) internal pure returns (address pairAddress) {
        (address token0, address token1) = sortTokens(tokenA, tokenB);
        pairAddress = address(
            uint160(
                uint256(
                    keccak256(
                        abi.encodePacked(
                            hex"ff",
                            factoryAddress,
                            keccak256(abi.encodePacked(token0, token1)),
                            keccak256(type(ZuniswapV2Pair).creationCode)
                        )
                    )
                )
            )
        );
    }
```

#### getAmountIn function

we know the exact amount of output tokens we want to get but we don't know how much input tokens we need to give.
Let's return to the swapping formula:
(x+rΔx)(y−Δy)=xy
afer applying basic algebraic operations we get:
Again, after applying basic algebraic operations we get:
Δx= (y−Δy)r / xΔy

```solidity
function getAmountIn(
        uint256 amountOut,
        uint256 reserveIn,
        uint256 reserveOut
    ) public pure returns (uint256) {
        if (amountOut == 0) revert InsufficientAmount();
        if (reserveIn == 0 || reserveOut == 0) revert InsufficientLiquidity();

        uint256 numerator = reserveIn * amountOut * 1000;
        uint256 denominator = (reserveOut - amountOut) * 997;
        // +1 is because the integer division, in solidity result down, which means the result get truncated.
        return (numerator / denominator) + 1;
    }
```

#### flashloan

First thing you need to know about flash loans implementation is that **they can only be used by smart contracts**. Here's how borrowing and repaying happens with flash loans:

A smart contract borrows a flash loan from another contract.
The lender contract sends tokens to the borrowing contract and calls a special function in this contract.
In the special function, the borrowing contract performs some operations with the loan and then transfers the loan back.
The lender contract ensures that the whole amount was paid back. In case when there are fees, it also ensures that they were paid.
Control flow returns to the borrowing contract.

```solidity
function swap(
        uint256 amount0Out,
        uint256 amount1Out,
        address to,
        bytes calldata data
    ) public nonReentrant {
        if (amount0Out == 0 && amount1Out == 0)
            revert InsufficientOutputAmount();

        (uint112 reserve0_, uint112 reserve1_, ) = getReserves();

        if (amount0Out > reserve0_ || amount1Out > reserve1_)
            revert InsufficientLiquidity();

        if (amount0Out > 0) _safeTransfer(token0, to, amount0Out);
        if (amount1Out > 0) _safeTransfer(token1, to, amount1Out);
        if (data.length > 0)
            IZuniswapV2Callee(to).zuniswapV2Call(
                msg.sender,
                amount0Out,
                amount1Out,
                data
            );
            ...
```

```solidity
contract Flashloaner {
    error InsufficientFlashLoanAmount();

    uint256 expectedLoanAmount;

    function flashloan(
        address pairAddress,
        uint256 amount0Out,
        uint256 amount1Out,
        address tokenAddress
    ) public {
        if (amount0Out > 0) {
            expectedLoanAmount = amount0Out;
        }
        if (amount1Out > 0) {
            expectedLoanAmount = amount1Out;
        }

        ZuniswapV2Pair(pairAddress).swap(
            amount0Out,
            amount1Out,
            address(this),
            abi.encode(tokenAddress)
        );
    }

    function zuniswapV2Call(
        address sender,
        uint256 amount0Out,
        uint256 amount1Out,
        bytes calldata data
    ) public {
        // decode bytes to get tokenAddress.
        address tokenAddress = abi.decode(data, (address));
        uint256 balance = ERC20(tokenAddress).balanceOf(address(this));

        if (balance < expectedLoanAmount) revert InsufficientFlashLoanAmount();

        ERC20(tokenAddress).transfer(msg.sender, balance);
    }
}
```

**relations between ZuniswapV2Pair and ZuniswpaV2Router**
ZuniswapV2Pair = low-level AMM pair contract

ZuniswapV2Router = high-level helper / UX layer

- Computes amounts, enforces slippage bounds, and moves tokens between users and pairs.
- Uses ZuniswapV2Library.pairFor(...) to find the pair address (deterministic via Factory/CREATE2).
- Calls pair functions after preparing token transfers:
  Router orchestrates user flows (calculations + transfers) and delegates state changes to Pair. Pair is the authoritative stateful contract holding tokens and enforcing AMM logic.
