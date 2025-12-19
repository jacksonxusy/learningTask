### uniswap new variables.

1. L is liqudity with respect to price, not juest token balances, it is derived from invariant x* y = k
2. L is defeined such that L = square root x* y.
3. Uniswap V3 defines price as: P=y/x, the price of token0 in terms of token1.

Liquidity math uses √P instead of P becuase:
√P reduces rounding errors (square roots computed only once at tick boundaries, not dynamically in swaps).

√P ties directly to liquidity math and the amount of tokens moved when price changes.

**Uniswap V3’s math works because**:

- Liquidity L stays constant within a price range.
- Token quantities at any price can be derived from L and √P.
- Swap math becomes linear in price coordinate √P instead of nonlinear in token balances — this simplifies crossing ranges and computing fees precisely.

## Pool Contract

### mint function

Ticks are **price level markers that manage liquidity across the entire pool**:

- Purpose: Track when liquidity becomes active/inactive as price moves
- Scope: Global to the pool - affect ALL users
- Key Function: Gate liquidity activation based on current price
- Data Structure: mapping(int24 => Tick.Info) - one entry per tick boundary
- Updated When: Any user adds liquidity that crosses this tick boundary

Positions are **user-specific liquidity holdings**:

- Purpose: Track individual user's liquidity and fees earned
- Scope: Per-user - only affects the specific position owner
- Key Function: Record how much liquidity a user owns in a specific range
- Data Structure: mapping(bytes32 => Position.Info) - key = owner + tick range
- Updated When: That specific user modifies their position

Each position is uniquely identified by three keys: owner address, lower tick index, and upper tick index. We hash the three to make storing data cheaper: when hashed, every key will take 32 bytes, instead of 96 bytes when owner, lowerTick, and upperTick are separate keys.

**If we use three keys, we need three mappings. Each key would be stored separately and would take 32 bytes since Solidity stores values in 32-byte slots (when packing is not applied).**

```solidity
function get(
    mapping(bytes32 => Info) storage self,
    address owner,
    int24 lowerTick,
    int24 upperTick
) internal view returns (Position.Info storage position) {
    position = self[
        keccak256(abi.encodePacked(owner, lowerTick, upperTick))
    ];
}
```

call the uniswapV3MintCallback method on the caller–this is the callback. It’s expected that the caller (whoever calls mint) is a contract because non-contract addresses cannot implement functions in Ethereum.
The caller is expected to implement uniswapV3MintCallback and transfer tokens to the Pool contract in this function.

```solidity
 IUniswapV3MintCallback(msg.sender).uniswapV3MintCallback(
        amount0,
        amount1
    );
```

**Why store sqrt(y/x)**:

1. Gas Savings: Eliminates expensive sqrt() calculations during swaps
2. Precision: Fixed-point arithmetic with 96 fractional bits (~29 decimal places)
3. No Floats: Everything stays as integers, avoiding float precision issues
4. Math Convenience: Many AMM formulas naturally use √P

**Precision is preserved because**:

- Uses 160-bit integers (not floating point)
- 96 bits of fractional precision
- Deterministic fixed-point arithmetic
- No rounding during storage/manipulation

Ticks represent discrete price points in a logarithmic scale.

**Tick-to-Price Formula**

// Price = 1.0001^tick
price = 1.0001^tick

// sqrtPriceX96 = sqrt(price) × 2^96
sqrtPriceX96 = sqrt(1.0001^tick) × 2^96
sqrtPriceX96 = 1.0001^(tick/2) × 2^96

Example Calculation

```solidity
Let's say:
- Current ETH price = $2,000 USDC
- tick = 85184 (from your code)

Step 1: Calculate price from tick
price = 1.0001^85184

// Using logarithms:
log10(price) = 85184 × log10(1.0001)
log10(price) = 85184 × 0.0000434
log10(price) = 3.698

price = 10^3.698 ≈ 5,000 USDC per ETH

Step 2: Calculate sqrtPriceX96
sqrtPrice = sqrt(5000) ≈ 70.71
sqrtPriceX96 = 70.71 × 2^96 ≈ 70.71 × 7.92×10^28
sqrtPriceX96 ≈ 5.6×10^30
```

**Why sqrtPriceX96?**

Uniswap V3 uses this format for several mathematical reasons:

1. Avoids floating point math - Solidity doesn't support decimals well
2. Enables precise calculations - 96 bits of fractional precision
3. Simplifies liquidity math - many formulas use √P directly
4. Prevents overflow/underflow - square roots keep numbers in manageable ranges

The Formula

```solidity
// Actual price (token1/token0) = (sqrtPriceX96 / 2^96)²
sqrtPriceX96 = sqrt(price) × 2^96
```

Let's Decode Your Example

```solidity
uint160 sqrtPriceX96 = 5604469350942327889444743441197;

Step 1: Remove the 2^96 scaling factor
sqrtPrice = 5604469350942327889444743441197 / 2^96
sqrtPrice = 5604469350942327889444743441197 / 79,228,162,514,264,337,593,543,950,336
sqrtPrice ≈ 0.00007071

Step 2: Square it to get the actual price
price = sqrtPrice² = (0.00007071)² ≈ 0.000005

Step 3: Interpret the price ratio
// If token0 = ETH, token1 = USDC:
price = USDC/ETH = 0.000005
// This means: 1 ETH = 200,000 USDC (seems unrealistic, likely reverse)

// If token0 = USDC, token1 = ETH:
price = ETH/USDC = 0.000005
// This means: 1 USDC = 0.000005 ETH = 1 ETH = 200,000 USDC (also unrealistic)
```

**ether unit**: 
ether is a unit convention that is commonly used for any ERC20 token that has 18 decimal place. including usdc.
USDC Decimal Places

- USDC actually has 6 decimal places, not 18

- But in this demo code, they're using ether (18 decimals) for simplicity

- Real implementation would use 5042 * 10^6 for USDC
  
  ```solidity
  uint256 usdcBalance = 5042 ether;
  ```
  
  When you mint liquidity in Uniswap V3, you need to provide both tokens. The amount0 calculation determines how much of token0 is required for a specific liquidity amount within a price range.

**TickMath.getSqrtRatioAtTick(tick)** Converts tick numbers to square root prices

```solidity
function getSqrtRatioAtTick(int24 tick) returns (uint160 sqrtPriceX96) {
      // Converts tick index to actual price representation
      return sqrt(1.0001^tick) * 2^96
  }
```

**Math.calcAmount0Delta(sqrtPriceA, sqrtPriceB, liquidity)**
Calculates how much token0 is needed:

```solidity
 function calcAmount0Delta(uint160 sqrtPriceA, uint160 sqrtPriceB, uint128 liquidity) {
      if (sqrtPriceA > sqrtPriceB) {
          (sqrtPriceA, sqrtPriceB) = (sqrtPriceB, sqrtPriceA);  // Ensure A < B
      }

      return (liquidity * 2^96) / sqrtPriceA - (liquidity * 2^96) / sqrtPriceB;
  }
```

3. The Formula Explained

The core calculation is:
**amount0 = liquidity * (1/sqrtPrice_lower - 1/sqrtPrice_upper)**

```solidity
Slot0 memory slot0_ = slot0;

amount0 = Math.calcAmount0Delta(
    TickMath.getSqrtRatioAtTick(slot0_.tick),
    TickMath.getSqrtRatioAtTick(upperTick),
    amount
);
```

### swap function.

```solidity
struct SwapState {
  uint256 amountSpecifiedRemaining;  // How much input left to swap
  uint256 amountCalculated;          // How much output calculated so far
  uint160 sqrtPriceX96;              // Current price during swap
  int24 tick;                        // Current tick during swap
}
struct StepState {
  uint160 sqrtPriceStartX96;  // Price at start of this step
  int24 nextTick;             // Next tick boundary we'll hit
  uint160 sqrtPriceNextX96;   // Price at next tick boundary
  uint256 amountIn;           // Input for this step
  uint256 amountOut;          // Output for this step
}

function swap(
        address recipient, // address receiving output tokens
        bool zeroForOne, // direction flag(true = swapping token0->token1, false=token1->token0)
        uint256 amountSpecified, // amount of input tokens that user wants to swap
        bytes calldata data // callback dtat for token transfers
    ) public returns (int256 amount0, int256 amount1) {
        Slot0 memory slot0_ = slot0;

        SwapState memory state = SwapState({
            amountSpecifiedRemaining: amountSpecified,
            amountCalculated: 0,
            sqrtPriceX96: slot0_.sqrtPriceX96, // start at current price
            tick: slot0_.tick // start at current tick
        });

        // the loop continues until we've processed the entire swap amount.
        while (state.amountSpecifiedRemaining > 0) {
            StepState memory step;
            step.sqrtPriceStartX96 = state.sqrtPriceX96;

            // find next tick boundary
            (step.nextTick, ) = tickBitmap.nextInitializedTickWithinOneWord(
                state.tick,
                1,
                zeroForOne
            );

            step.sqrtPriceNextX96 = TickMath.getSqrtRatioAtTick(step.nextTick);

            // this calculates how much we can swap before hitting the next tick boundary.
            (state.sqrtPriceX96, step.amountIn, step.amountOut) = SwapMath
                .computeSwapStep(
                    step.sqrtPriceStartX96,
                    step.sqrtPriceNextX96,
                    liquidity,
                    state.amountSpecifiedRemaining
                );

            // update state
            state.amountSpecifiedRemaining -= step.amountIn; // reduce remaining input
            state.amountCalculated += step.amountOut; // add to output
            state.tick = TickMath.getTickAtSqrtRatio(state.sqrtPriceX96); // update tick.
        }

        // update the pool's price and tick to relect the new state after swap.
        if (state.tick != slot0_.tick) {
            (slot0.sqrtPriceX96, slot0.tick) = (state.sqrtPriceX96, state.tick);
        }

        // positive: tokens flowing into the pool
        // negative: tokens flowing out of pool
        (amount0, amount1) = zeroForOne
            ? (
                int256(amountSpecified - state.amountSpecifiedRemaining), // input used
                -int256(state.amountCalculated) // output (negative)
            )
            : (
                -int256(state.amountCalculated), // output (negative)
                int256(amountSpecified - state.amountSpecifiedRemaining) // input used
            );

        if (zeroForOne) {
            // token0 -> tokne1
            IERC20(token1).transfer(recipient, uint256(-amount1)); // send output

            uint256 balance0Before = balance0();
            IUniswapV3SwapCallback(msg.sender).uniswapV3SwapCallback( // get input
                amount0,
                amount1,
                data
            );
            if (balance0Before + uint256(amount0) > balance0())
                revert InsufficientInputAmount();
        } else {
            IERC20(token0).transfer(recipient, uint256(-amount0));

            uint256 balance1Before = balance1();
            IUniswapV3SwapCallback(msg.sender).uniswapV3SwapCallback(
                amount0,
                amount1,
                data
            );
            if (balance1Before + uint256(amount1) > balance1())
                revert InsufficientInputAmount();
        }

        emit Swap(
            msg.sender,
            recipient,
            amount0,
            amount1,
            slot0.sqrtPriceX96,
            liquidity,
            slot0.tick
        );
    }
```

Real-World Example

  User swaps 1000 USDC for ETH:

1. Start: Current price $2000/ETH
2. Step 1: Swap 300 USDC → 0.15 ETH (price moves to $2005/ETH)
3. Step 2: Cross tick, new liquidity available, price $2010/ETH
4. Step 3: Swap 400 USDC → 0.198 ETH (price moves to $2020/ETH)
5. Step 4: Cross another tick, continue...
6. Final: Total 0.49 ETH received, average price ~$2040/ETH.

**cross tick swap flow**
Scenario: User swaps 1000 USDC for ETH

- Current price: $2000/ETH (tick 85176)
- Next tick: $2015/ETH (tick 85200)
- Liquidity in current range: 50 ETH
- Liquidity in next range: 30 ETH

Step 1: First Tick (Current Range)

```solidity
// Initial state
state.amountSpecifiedRemaining = 1000 USDC
state.sqrtPriceX96 = sqrt(2000) × Q96
state.tick = 85176
liquidity = 50 ETH

// Find next boundary
nextTick = 85200 (where price = $2015)
sqrtPriceNextX96 = sqrt(2015) × Q96

// Calculate swap step
(sqrtPriceNext, amountIn1, amountOut1) = SwapMath.computeSwapStep(
  sqrt(2000) × Q96,    // Current price
  sqrt(2015) × Q96,    // Target price
  50 ETH,              // Available liquidity
  1000 USDC           // Full amount
)

// Result: We can only swap 400 USDC before hitting $2015
amountIn1 = 400 USDC
amountOut1 = 0.199 ETH
newPrice = $2015/ETH
```

Step 2: Update State & Cross Tick

```solidity
// After first iteration
state.amountSpecifiedRemaining = 1000 - 400 = 600 USDC
state.amountCalculated = 0.199 ETH
state.sqrtPriceX96 = sqrt(2015) × Q96
state.tick = 85200

// Loop continues because we still have 600 USDC left
```

Step 3: Second Tick (New Range)

```solidity
// Find next boundary from current tick
nextTick = 85250 (where price = $2030)
sqrtPriceNextX96 = sqrt(2030) × Q96

// New liquidity in this range
liquidity = 30 ETH  // Different from previous tick!

// Calculate second swap step
(sqrtPriceNext, amountIn2, amountOut2) = SwapMath.computeSwapStep(
  sqrt(2015) × Q96,    // Current price
  sqrt(2030) × Q96,    // Target price
  30 ETH,              // Available liquidity in this range
  600 USDC           // Remaining amount
)

// Result: We can swap all 600 USDC
amountIn2 = 600 USDC
amountOut2 = 0.296 ETH
newPrice = $2025/ETH (didn't reach target)
```

Step 4: Final State

```solidity
// Final totals
totalAmountIn = amountIn1 + amountIn2 = 400 + 600 = 1000 USDC
totalAmountOut = amountOut1 + amountOut2 = 0.199 + 0.296 = 0.495 ETH
finalPrice = $2025/ETH
finalTick = 85225
```

**Gas Optimization Techniques**

1. Tick Bitmap Efficiency

```solidity
// Instead of checking every tick:
for (int24 tick = currentTick; tick <= targetTick; tick++) {
  if (ticks[tick].initialized) { /* process */ }
}

// Use bitmap to jump directly to next active tick:
nextTick = tickBitmap.nextInitializedTick(currentTick);
```

2. Math Optimization
   
   ```solidity
   // Reuse calculations between steps
   uint256 liquidityTimesQ96 = liquidity * Q96;  // Calculate once
   // Used in both amount0 and amount1 calculations
   ```

3. Memory vs Storage
   
   ```solidity
   Slot0 memory slot0_ = slot0;  // Read once, use multiple times
   // Avoids multiple storage reads
   ```

### slippage protection

During sandwiching, attackers wrap your swap transactions in their two transactions: one goes before your transaction and other goes after it. In the first transaction, an attacker modifies the state of a pool so that your swap becomes very unprofitable for you and somewhat profiable for the attacker. This is achieved by adjusting pool liquidity so that your trade happens at a lower price. In the second transaction, the attacker reestablish pool liquidity and the price. As a result, you get much fewer tokens than expected due to manipulated prices, and the attacker gets some profit.



```solidity
function swap(
    address recipient,
    bool zeroForOne,
    uint256 amountSpecified,
    uint160 sqrtPriceLimitX96,
    bytes calldata data
) public returns (int256 amount0, int256 amount1) {
    ...
    if (
        zeroForOne
            ? sqrtPriceLimitX96 > slot0_.sqrtPriceX96 ||
                sqrtPriceLimitX96 < TickMath.MIN_SQRT_RATIO
            : sqrtPriceLimitX96 < slot0_.sqrtPriceX96 &&
                sqrtPriceLimitX96 > TickMath.MAX_SQRT_RATIO
    ) revert InvalidPriceLimit();
    ...
```

When selling token x (`zeroForOne` is true), `sqrtPriceLimitX96` must be between the current price and the minimal P​ since selling token x moves the price down. Likewise, when selling token y, `sqrtPriceLimitX96` must be between the current price and the maximal P​ because the price moves up.

**Binary vs. Decimal Fixed-Point**

The page highlights a crucial distinction between two types of fixed-point systems:

- **Binary Fixed-Point ($Q64.96$):** This is what **Uniswap V3** uses.

- **Structure:** 64 bits are used for the integer part and 96 bits for the fractional part.

- **Base:** It is base-2. To convert a whole number to $Q64.96$, you multiply it by $2^{96}$ (or use the bitwise left shift operator << 96).

- **Reason:** Computers and the EVM handle binary operations (like bit shifting) very efficiently.

- **Decimal Fixed-Point ($UD60.18$):** This is common in many other DeFi protocols (and used by the PRBMath library).

- **Structure:** 60 digits for the integer part and 18 for the fractional part (matching the 18 decimals of ETH/ERC20 tokens).

- **Base:** It is base-10. You convert numbers by multiplying by $10^{18}$.

**3. Key Conversions**

The page provides examples of how the same number looks in both systems:

- **For the number 42:**

- **$Q64.96$:** $42 \times 2^{96} \approx 3.32 \times 10^{30}$

- **$UD60.18$:** $42 \times 10^{18} = 42,000,000,000,000,000,000$

**Why Uniswap V3 Uses Q64.96**

Uniswap V3 stores the **square root of the price** ($\sqrt{P}$) as a $Q64.96$ number. The choice of $96$ bits for the fractional part is specific:

- It provides enough precision to handle very small price movements (ticks).
- It allows for a wide range of prices (from nearly zero to incredibly large values) while fitting within a uint160 or uint256.
- It makes the math for "liquidity" and "amounts" work out more cleanly when multiplying and dividing.

### quoter

How it works:

1. Calls the actual pool's swap function
2. Pool calculates exact swap amounts using real AMM logic
3. Pool calls back to quoter for token transfers
4. Quoter reverts with calculated data instead of transferring tokens
5. Main function catches the revert and returns the data

```solidity
contract UniswapV3Quoter {
    struct QuoteParams {
        address pool;
        uint256 amountIn;
        bool zeroForOne;
    }

    function quote(QuoteParams memory params)
        public
        returns (
            uint256 amountOut,
            uint160 sqrtPriceX96After,
            int24 tickAfter
        )
    {
        try
            IUniswapV3Pool(params.pool).swap(
                address(this),
                params.zeroForOne,
                params.amountIn,
                abi.encode(params.pool)
            )
        {} catch (bytes memory reason) {
            return abi.decode(reason, (uint256, uint160, int24));
        }
    }

    function uniswapV3SwapCallback(
        int256 amount0Delta,
        int256 amount1Delta,
        bytes memory data
    ) external view {
        address pool = abi.decode(data, (address));

        uint256 amountOut = amount0Delta > 0
            ? uint256(-amount1Delta)
            : uint256(-amount0Delta);

        (uint160 sqrtPriceX96After, int24 tickAfter) = IUniswapV3Pool(pool)
            .slot0();

        assembly {
            let ptr := mload(0x40) // get free memory pointer
            mstore(ptr, amountOut) // store amountOut (32bytes)
            mstore(add(ptr, 0x20), sqrtPriceX96After) // store sqrtPriceX96After (next 32 bytes)
            mstore(add(ptr, 0x40), tickAfter) // store tickAfter (next 32 bytes)
            revert(ptr, 96) // revert with 96 bytes of data.
        }
    }
}
```

Real-World Usage Example

Frontend Integration

// User wants to swap 1000 USDC for ETH
UniswapV3Quoter quoter = UniswapV3Quoter(QUOTER_ADDRESS);

QuoteParams memory params = QuoteParams({
  pool: POOL_ADDRESS,
  amountIn: 1000 * 10** 6,  // 1000 USDC (6 decimals)
  zeroForOne: false        // USDC → ETH
});

(uint256 amountOut, uint160 sqrtPriceAfter, int24 tickAfter) = quoter.quote(params);

// Show user:
// "You'll receive: amountOut ETH"
// "Price will move to: sqrtPriceAfter"
// "Current gas cost: estimate"

## lib

### tick bitmap

we can calculate compressed tick via this formula: ` actual tick = ((wordPos * 256) + bitPos) * tick spacing`.
**flip tick function:**

1. calculate  wordPos and bit pos. wordPos determined by the integer division of compressed tick by 256, this is key for the tickBitmap mapping, bitPos determined by the remainder of compressed tick divided by 256. this is index within uint256 word.
2. create the mask, `uint256 mask = 1 << bitPos;` the 1 is shifted left by bitPos positions, this create a uint256 where only the target bit is set to 1, and all other bits are 0.
3. fip the bit. `self[wordPos] ^= mask;`  by xor operation, we only invert the target tick position of the word.(0 to 1 or 1 to 0), other 255 tick position not change, the is most efficient way to save gas in evm.
   XOR features:
- X xor 0 = 0 (consistent with 0)
- X xor 1 = -X(invert with 1)

```solidity
function flipTick(
    mapping(int16 => uint256) storage self,
    int24 tick,
    int24 tickSpacing
) internal {
    require(tick % tickSpacing == 0); // ensure that the tick is spaced
    (int16 wordPos, uint8 bitPos) = position(tick / tickSpacing);
    uint256 mask = 1 << bitPos;
    self[wordPos] ^= mask;
}
```

**find next tick function**

1. lte is the flag that sets the direction. When true, we're selling token x and searching for the next initialized tick to the current one, When false, it is the other way around.
   when selling x, we're:
2. taking the current tick's word and bit positions.
3. making a mask where all bits to the rights of the current bit position, including it. eg 000011111111...
4. appling the mask to the current tick's word. this can find all ticks in the right.
5. calculating the next,  BitMath.mostSignificantBit can find the closest bit next to bitpos.

```solidity
function nextInitializedTickWithinOneWord(
        mapping(int16 => uint256) storage self,
        int24 tick,
        int24 tickSpacing,
        bool lte
    ) internal view returns (int24 next, bool initialized) {
        int24 compressed = tick / tickSpacing;
        if (tick < 0 && tick % tickSpacing != 0) compressed--; // round towards negative infinity

        if (lte) {
            (int16 wordPos, uint8 bitPos) = position(compressed);
            // all the 1s at or to the right of the current bitPos
            uint256 mask = (1 << bitPos) - 1 + (1 << bitPos);
            uint256 masked = self[wordPos] & mask;

            // if there are no initialized ticks to the right of or at the current tick, return rightmost in the word
            initialized = masked != 0;
            // overflow/underflow is possible, but prevented externally by limiting both tickSpacing and tick
            next = initialized
                ? (compressed - int24(uint24(bitPos - BitMath.mostSignificantBit(masked)))) * tickSpacing
                : (compressed - int24(uint24(bitPos))) * tickSpacing;
        } else {
            // start from the word of the next tick, since the current tick state doesn't matter
            (int16 wordPos, uint8 bitPos) = position(compressed + 1);
            // all the 1s at or to the left of the bitPos
            uint256 mask = ~((1 << bitPos) - 1);
            uint256 masked = self[wordPos] & mask;

            // if there are no initialized ticks to the left of the current tick, return leftmost in the word
            initialized = masked != 0;
            // overflow/underflow is possible, but prevented externally by limiting both tickSpacing and tick
            next = initialized
                ? (compressed + 1 + int24(uint24((BitMath.leastSignificantBit(masked) - bitPos)))) * tickSpacing
                : (compressed + 1 + int24(uint24((type(uint8).max - bitPos)))) * tickSpacing;
        }
    }
```

### tick

Part A: (liquidityAfter == 0)
Result: true if liquidity is zero after the operation (meaning the last liquidity was removed).

Result: false if liquidity is non-zero after the operation (meaning there is still liquidity here).

Part B: (liquidityBefore == 0)
Result: true if liquidity was zero before the operation (meaning new liquidity is being added).

Result: false if liquidity was non-zero before the operation (meaning there was already liquidity)

**The flipped variable is set to true only when the $\text{Tick}$ crosses the zero-liquidity threshold.**
Case 1: **Liquidity is added where none existed**. (True activation: 0 -> 1)
A = true (before was zero) vs B = false (after is non-zero) -> true.

case2: **The last remaining liquidity is removed**. (True deactivation: 0 > 1)
A = false (before was non-zero) vs B = true (after is zero) -> true.

```solidity
library Tick {
    struct Info {
        bool initialized;
        uint128 liquidity;
    }

    function update(
        mapping(int24 => Tick.Info) storage self,
        int24 tick,
        uint128 liquidityDelta
    ) internal returns (bool flipped) {
        Tick.Info storage tickInfo = self[tick];
        uint128 liquidityBefore = tickInfo.liquidity;
        uint128 liquidityAfter = liquidityBefore + liquidityDelta;

        flipped = (liquidityAfter == 0) != (liquidityBefore == 0);

        if (liquidityBefore == 0) {
            tickInfo.initialized = true;
        }

        tickInfo.liquidity = liquidityAfter;
    }
}
```
