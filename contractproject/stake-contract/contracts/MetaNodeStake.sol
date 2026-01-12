// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/utils/Address.sol";
import "@openzeppelin/contracts/utils/math/Math.sol";

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";

// user lock ETH or other tokens - they earn metanodeToken rewards over time -> rewards 
// are calculated per block (like mining).
contract MetaNodeStake is Initializable, UUPSUpgradeable, PausableUpgradeable, AccessControlUpgradeable {
    
    using SafeERC20 for IERC20;
    using Address for address;
    using Math for uint256;

    bytes32 public constant ADMIN_ROLE = keccak256("admin_role");
    bytes32 public constant UPGRADE_ROLE = keccak256("upgrade_role");

    uint256 public constant ETH_PID = 0;

    // one staking pool, ETH pool, USDT pool. 
    struct Pool {
        address stTokenAddress;
        // how much reward this pool gets vs others, eth pool = 70%, usdt = 30%. 
        uint256 poolWeight;
        uint256 lastRewardBlock;
        // maigc number for reward calculation. eg: 1 staked token = X metanode.
        uint256 accMetaNodePerST;
        // total staked tokens in the pool
        uint256 stTokenAmount;
        uint256 minDepositAmount;
        // number of blocks the staked token must remain locked after a user initiates an unstake request.
        uint256 unstakeLockedBlocks;
    }

    struct UnstakeRequest {
        uint256 amount;
        // the block number when the requested tokens becoms available for final withdraw. 
        // current block + unstakeLockedBlocks.
        uint256 unlockBlock;
    }

    struct User {
        uint256 stAmount;
        // the cumulative value of stAmount * accMetaNodePerST at a time of 
        // the user's last interaction(deposit/unstake/claim). used to calculate new pending rewards.
        uint256 finishedMetaNode;
        // rewards you have earned so far but have not been realized.
        uint256 pendingMetaNode;
        UnstakeRequest[] requests;
    }

    uint256 public startBlock;
    uint256 public endBlock;
    uint256 public metaNodePerBlock;

    bool public withdrawPaused;
    bool public claimPaused;

    IERC20 public metaNode;
    // the sum of the weight numbers that the admin typed then creating pools.
    // When admin calls addPool() or setPoolWeight(),What happens to totalPoolWeight
    // "addPool(ETH, weight = 70, ...)",totalPoolWeight += 70
    // "addPool(USDT, weight = 20, ...)",totalPoolWeight += 20 → now 90
    uint256 public totalPoolWeight;
    Pool[] public pool;
    
    mapping(uint256 => mapping(address => User)) public user;
    event SetMetaNode(IERC20 indexed MetaNode);

    event PauseWithdraw();

    event UnpauseWithdraw();

    event PauseClaim();

    event UnpauseClaim();

    event SetStartBlock(uint256 indexed startBlock);

    event SetEndBlock(uint256 indexed endBlock);

    event SetMetaNodePerBlock(uint256 indexed MetaNodePerBlock);

    event AddPool(
        address indexed stTokenAddress,
        uint256 indexed poolWeight,
        uint256 indexed lastRewardBlock,
        uint256 minDepositAmount,
        uint256 unstakeLockedBlocks
    );

    event UpdatePoolInfo(
        uint256 indexed poolId,
        uint256 indexed minDepositAmount,
        uint256 indexed unstakeLockedBlocks
    );

    event SetPoolWeight(
        uint256 indexed poolId,
        uint256 indexed poolWeight,
        uint256 totalPoolWeight
    );

    event UpdatePool(
        uint256 indexed poolId,
        uint256 indexed lastRewardBlock,
        uint256 totalMetaNode
    );

    event Deposit(address indexed user, uint256 indexed poolId, uint256 amount);

    event RequestUnstake(
        address indexed user,
        uint256 indexed poolId,
        uint256 amount
    );

    event Withdraw(
        address indexed user,
        uint256 indexed poolId,
        uint256 amount,
        uint256 indexed blockNumber
    );

    event Claim(
        address indexed user,
        uint256 indexed poolId,
        uint256 MetaNodeReward
    );

    modifier checkPid(uint256 _pid) {
        require(_pid < pool.length, "invalid pid");
        _;
    }

    modifier whenNotClaimPaused() {
        require(!claimPaused, "claim is paused");
        _;
    }

    modifier whenNotWithdrawPaused() {
        require(!withdrawPaused, "withdraw is paused");
        _;
    }

    function initialize(IERC20 _metaNode, uint256 _startBlock, uint256 _endBlock, uint256 _metaNodePerBlock) public initializer {
        require(_startBlock <= _endBlock && _metaNodePerBlock >0, "invalid parameters");
        __AccessControl_init();
        __UUPSUpgradeable_init();
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(UPGRADE_ROLE, msg.sender);
        _grantRole(ADMIN_ROLE, msg.sender);

        setMetaNode(_metaNode);
        startBlock = _startBlock;
        endBlock = _endBlock;
        metaNodePerBlock = _metaNodePerBlock;

    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(UPGRADE_ROLE) {}

    function setMetaNode(IERC20 _metaNode) public onlyRole(ADMIN_ROLE) {
        metaNode = _metaNode;
    }

    function pauseWithdraw() public onlyRole(ADMIN_ROLE) whenNotWithdrawPaused() {
        withdrawPaused = true;
        emit PauseWithdraw();
    }

    function unpauseWithdraw() public onlyRole(ADMIN_ROLE) whenNotClaimPaused() {
        withdrawPaused = false;
        emit UnpauseWithdraw();
    }

    function pauseClaim() public onlyRole(ADMIN_ROLE) {
        require(!claimPaused, "claim has been already paused");

        claimPaused = true;

        emit PauseClaim();
    }

    /**
     * @notice Unpause claim. Can only be called by admin.
     */
    function unpauseClaim() public onlyRole(ADMIN_ROLE) {
        require(claimPaused, "claim has been already unpaused");

        claimPaused = false;

        emit UnpauseClaim();
    }

    /**
     * @notice Update staking start block. Can only be called by admin.
     */
    function setStartBlock(uint256 _startBlock) public onlyRole(ADMIN_ROLE) {
        require(
            _startBlock <= endBlock,
            "start block must be smaller than end block"
        );

        startBlock = _startBlock;

        emit SetStartBlock(_startBlock);
    }

    /**
     * @notice Update staking end block. Can only be called by admin.
     */
    function setEndBlock(uint256 _endBlock) public onlyRole(ADMIN_ROLE) {
        require(
            startBlock <= _endBlock,
            "start block must be smaller than end block"
        );

        endBlock = _endBlock;

        emit SetEndBlock(_endBlock);
    }

    /**
     * @notice Update the MetaNode reward amount per block. Can only be called by admin.
     */
    function setMetaNodePerBlock(
        uint256 _MetaNodePerBlock
    ) public onlyRole(ADMIN_ROLE) {
        require(_MetaNodePerBlock > 0, "invalid parameter");

        metaNodePerBlock = _MetaNodePerBlock;

        emit SetMetaNodePerBlock(_MetaNodePerBlock);
    }

    function addPool(address _stTokenAddress, uint256 _poolWeight, 
            uint256 _minDepositAmount, uint256 _unstakeLockedBlocks, bool _withUpdate) public onlyRole(ADMIN_ROLE) {
        if (pool.length > 0) {
            require(_stTokenAddress != address(0x0), "invalid staking token addresss");
        } else {
            require(_stTokenAddress == address(0x0), "invalid staking token address");
        }

        require(_unstakeLockedBlocks > 0, "invalid withdraw locked blocks");
        require(block.number < endBlock, "already ended");

        if (_withUpdate) {
            massUpdatePools();
        }

        uint256 lastRewardBlock = block.number > startBlock ? block.number : startBlock;
        totalPoolWeight += _poolWeight;

        pool.push(Pool(
            {
                stTokenAddress: _stTokenAddress,
                poolWeight: _poolWeight,
                lastRewardBlock: lastRewardBlock,
                accMetaNodePerST: 0,
                stTokenAmount: 0,
                minDepositAmount: _minDepositAmount,
                unstakeLockedBlocks: _unstakeLockedBlocks
            }
        ));
        emit AddPool(
            _stTokenAddress,
            _poolWeight,
            lastRewardBlock,
            _minDepositAmount,
            _unstakeLockedBlocks
        );
    }

    // it changes how much % of total rewards a staking pool gets, like changing the APY of ETH staking vs USDT staking
    // Pool,Current Weight,Current %,Gets per block
    // ETH  pool, 70, 70%, 70 MetaNode
    // USDT pool, 30, 30%, 30 MetaNode
    // Total,100, 100%, 100 MetaNode
    function setPoolWeight(uint256 _pid, uint256 _poolWeight, bool _withUpdate) public onlyRole(ADMIN_ROLE) checkPid(_pid) {
        require(_poolWeight > 0, "invalid pool weight");

        if (_withUpdate) {
            massUpdatePools();
        }

        totalPoolWeight = totalPoolWeight - pool[_pid].poolWeight + _poolWeight;

        pool[_pid].poolWeight = _poolWeight;

        emit SetPoolWeight(_pid, _poolWeight, totalPoolWeight);

    }

    //It forces every single pool to recalculate and update its accMetaNodePerST (the magic accumulator) 
    //up to the current block,
    // even if no one has touched that pool for a long time.
    function massUpdatePools() public {
        uint256 length = pool.length;
        for (uint256 pid = 0; pid < length; pid++) {
            updatePool(pid);
        }
    }

    function updatePool(uint256 _pid, uint256 _minDepositAmount, uint256 _unstakeLockedBlocks) public onlyRole(ADMIN_ROLE) checkPid(_pid) {
        pool[_pid].minDepositAmount = _minDepositAmount;
        pool[_pid].unstakeLockedBlocks = _unstakeLockedBlocks;

        emit UpdatePoolInfo(_pid, _minDepositAmount, _unstakeLockedBlocks);
    }

    // bring this pool's reward up to the current block. 
    // it calculates how many new metaNode rewards appeared since last time -> adds them to acctMetaNodePerSt.
    function updatePool(uint256 _pid) public checkPid(_pid) {
        Pool storage pool_ = pool[_pid];
        if (block.number <= pool_.lastRewardBlock) {
            return;
        }
        // how many metaNodes should this pool get since last update?
        (bool success1, uint256 totalMetaNode) = getMultiplier(pool_.lastRewardBlock, block.number).tryMul(pool_.poolWeight);
        require(success1, "overflow");

        // totalMetaNode * poolweigh/totalPoolWeight = how many metanode does this pool have.
        (success1, totalMetaNode) = totalMetaNode.tryDiv(totalPoolWeight);
        require(success1, "overflow");

        uint256 stSupply = pool_.stTokenAmount;
        // if nobody staked, no need to update accumulator.
        if (stSupply > 0) {

            (bool success2, uint256 totalMetaNode_) = totalMetaNode.tryMul(1 ether);
            require(success2, "overflow");
            // get how many metaNodes of one stake token.
            // Block Range,New Reward per Token,accMetaNodePerST becomes
            // 1–100,    +0.5,   0.5e18
            // 101–200,  +0.7,  1.2e18
            // 201–300,  +0.3,  1.5e18
            (success2, totalMetaNode) = totalMetaNode_.tryDiv(stSupply);
            require(success2, "overflow");

            // 4. Add to the magic accumulator
            (bool success3, uint256 accMetaNodePerST) = pool_.accMetaNodePerST.tryAdd(totalMetaNode_);
            require(success3, "overflow");
            pool_.accMetaNodePerST = accMetaNodePerST;
        }
        
        // 5. Mark: "We're now updated to this block"
        pool_.lastRewardBlock = block.number;

        emit UpdatePool(_pid, pool_.lastRewardBlock, totalMetaNode);

    }

    // how many total metanode tokens should be given out from block X to block Y. 
    // number of blocks * metanodePerBlock.
    // getMultiplier(15,100,000, 15,100,010). multiplier = 10 blocks × 10 MetaNode/block = 100 MetaNode
    function getMultiplier(uint256 _from, uint256 _to) public view returns (uint256 multiplier) {
        require(_from <= _to, "invalid block");
        if (_from < startBlock) {
            _from = startBlock;
        }
        if (_to > endBlock) {
            _to = endBlock;
        }
        require(_from <= _to, "end block must be greater than start block");
        bool success;
        (success, multiplier) = (_to - _from).tryMul(metaNodePerBlock);
        require(success, "multiplier overflow");

    }

    function poolLength() external view returns (uint256) {
        return pool.length;
    }


    function pendingMetaNode(uint256 _pid, address _user) external view checkPid(_pid) returns(uint256) {
        return pendingMetaNodeByBlockNumber(_pid, _user, block.number);
    }

    // if the blockchain were frozeen at block X, how many metanode rewards could this user claim right now.
    function pendingMetaNodeByBlockNumber(uint256 _pid, address _user, uint256 _blockNumber) public view checkPid(_pid) returns (uint256) {
        Pool storage pool_ = pool[_pid];
        User storage user_ = user[_pid][_user];
        uint256 accMetaNodePerST = pool_.accMetaNodePerST;
        uint256 stSupply = pool_.stTokenAmount; 

        if (_blockNumber > pool_.lastRewardBlock && stSupply != 0) {
            uint256 multiplier = getMultiplier(pool_.lastRewardBlock, _blockNumber);
            uint256 metaNodePool = (multiplier * pool_.poolWeight) / totalPoolWeight;
            accMetaNodePerST = accMetaNodePerST + (metaNodePool * (1 ether) / stSupply);
        }
        // current - alreadyPaid + pending.
        return (user_.stAmount * accMetaNodePerST) / (1 ether) - user_.finishedMetaNode + user_.pendingMetaNode;
     }

    function stakingBalance(uint256 _pid, address _user) external view checkPid(_pid) returns (uint256) {
        return user[_pid][_user].stAmount;
     }

    function withdrawAmount(uint256 _pid, address _user) public view checkPid(_pid) returns(uint256 requestAmount, uint256 pendingWithdrawAmount) {
        User storage user_ = user[_pid][_user];
        for (uint256 i = 0; i < user_.requests.length; i++) {
            if (user_.requests[i].unlockBlock <= block.number) {
                pendingWithdrawAmount =
                    pendingWithdrawAmount +
                    user_.requests[i].amount;
            }
            requestAmount = requestAmount + user_.requests[i].amount;
        }
    }

    function depositETH() public payable whenNotPaused {
        Pool storage pool_ = pool[ETH_PID];
        require(
            pool_.stTokenAddress == address(0x0),
            "invalid staking token address"
        );

        uint256 _amount = msg.value;
        require(
            _amount >= pool_.minDepositAmount,
            "deposit amount is too small"
        );

        _deposit(ETH_PID, _amount);
    }

    function deposit(uint256 _pid, uint256 _amount) public whenNotPaused checkPid(_pid) {
        require(_pid != 0, "deposit not support ETH staking");
        Pool storage pool_ = pool[_pid];
        require(
            _amount > pool_.minDepositAmount,
            "deposit amount is too small"
        );
        if (_amount > 0) {
            IERC20(pool_.stTokenAddress).safeTransferFrom(msg.sender, address(this), _amount);
        }
        _deposit(_pid, _amount);

    }

    function _deposit(uint256 _pid, uint256 _amount) internal {
        Pool storage pool_ = pool[_pid];
        User storage user_ = user[_pid][msg.sender];

        updatePool(_pid);

        if (user_.stAmount > 0) {
            (bool success1, uint256 accST) = user_.stAmount.tryMul(pool_.accMetaNodePerST);
            require(success1, "user stamount mul perst overflow");
            (success1, accST) = accST.tryDiv(1 ether);
            require(success1, "accST div1 ether overflow");
             (bool success2, uint256 pendingMetaNode_) = accST.trySub(
                user_.finishedMetaNode
            );
            require(success2, "accST sub finishedMetaNode overflow");
            if (pendingMetaNode_ > 0) {
                //todo: 
                (bool success3, uint256 _pendingMetaNode) = user_.pendingMetaNode.tryAdd(pendingMetaNode_);
                require(success3, "user pendingMetaNode overflow");
                user_.pendingMetaNode = _pendingMetaNode;
            }
        }

        if (_amount > 0) {
            (bool success4, uint256 stAmount) = user_.stAmount.tryAdd(_amount);
            require(success4, "user stAmount overflow");
            user_.stAmount = stAmount;
        }

        (bool success5, uint256 stTokenAmount) = pool_.stTokenAmount.tryAdd(_amount);
        require(success5, "pool stTokenAmount overflow");
        pool_.stTokenAmount = stTokenAmount;

        // user_.finishedMetaNode = user_.stAmount.mulDiv(pool_.accMetaNodePerST, 1 ether);
        (bool success6, uint256 finishedMetaNode) = user_.stAmount.tryMul(
            pool_.accMetaNodePerST
        );
        require(success6, "user stAmount mul accMetaNodePerST overflow");

        (success6, finishedMetaNode) = finishedMetaNode.tryDiv(1 ether);
        require(success6, "finishedMetaNode div 1 ether overflow");

        user_.finishedMetaNode = finishedMetaNode;

        emit Deposit(msg.sender, _pid, _amount);


    }

    function _safeMetaNodeTransfer(address _to, uint256 _amount) internal {
        uint256 metaNodeBal = metaNode.balanceOf(address(this));

        if (_amount > metaNodeBal) {
            metaNode.transfer(_to, metaNodeBal);
        } else {
            metaNode.transfer(_to, _amount);
        }
    }

    function _safeETHTransfer(address _to, uint256 _amount) internal {
        (bool success, bytes memory data) = address(_to).call{value: _amount}("");
        require(success, "ETH transfer call failed");
        if (data.length > 0) {
            require(
                abi.decode(data, (bool)),
                "ETH transfer operation did not succeed"
            );
        }
    }

    function unstake(uint256 _pid, uint256 _amount) public whenNotPaused checkPid(_pid) whenNotClaimPaused {
        Pool storage pool_ = pool[_pid];
        User storage user_ = user[_pid][msg.sender];

        require(user_.stAmount >= _amount, "Not enough staking token balance");

        updatePool(_pid);
        // after the customer deposits their stake, their finshedMetaNode is a snapshot of their total reward entitilement
        // if the customer were to immediately check their pending rewards in the next block without any action
        // newly accrued reward = new entitlement - finishedMetaNode
        // since the finishedMetaNode was just set to old entitlement, the different would equal the reward earned in that
        // signle block.
        uint256 pendingMetaNode_ = (user_.stAmount * pool_.accMetaNodePerST) / (1 ether) - user_.finishedMetaNode;
        if (pendingMetaNode_ > 0) {
            user_.pendingMetaNode += pendingMetaNode_;
        }

        if (_amount > 0) {
            user_.stAmount -= _amount;
            user_.requests.push(
                UnstakeRequest({
                    amount: _amount,
                    unlockBlock: block.number + pool_.unstakeLockedBlocks
                })
            );
        }

        pool_.stTokenAmount = pool_.stTokenAmount - _amount;
        user_.finishedMetaNode =
            (user_.stAmount * pool_.accMetaNodePerST) /
            (1 ether);
        emit RequestUnstake(msg.sender, _pid, _amount);
    }

    function withdraw(uint256 _pid) public whenNotPaused checkPid(_pid) whenNotWithdrawPaused {
        Pool storage pool_ = pool[_pid];
        User storage user_ = user[_pid][msg.sender];

        uint256 pendingWithdraw_;
        uint256 popNum_;

        for(uint256 i = 0; i < user_.requests.length; i ++) {
            if (user_.requests[i].unlockBlock > block.number) {
                break;
            }
            pendingWithdraw_ += user_.requests[i].amount;
            popNum_++;
        }

        for (uint256 i = 0; i < user_.requests.length - popNum_; i++) {
            user_.requests[i] = user_.requests[i + popNum_];
        }

        for (uint256 i = 0; i < popNum_; i++) {
            user_.requests.pop();
        }

        if (pendingWithdraw_ > 0) {
            if (pool_.stTokenAddress == address(0x0)) {
                _safeETHTransfer(msg.sender, pendingWithdraw_);
            } else {
                IERC20(pool_.stTokenAddress).safeTransfer(
                    msg.sender,
                    pendingWithdraw_
                );
            }
        }
        emit Withdraw(msg.sender, _pid, pendingWithdraw_, block.number);
    }

    function claim(uint256 _pid) public whenNotPaused checkPid(_pid) whenNotClaimPaused {
        Pool storage pool_ = pool[_pid];
        User storage user_ = user[_pid][msg.sender];

        uint256 pendingMetaNode_ = (user_.stAmount * pool_.accMetaNodePerST) / (1 ether) -
            user_.finishedMetaNode + user_.pendingMetaNode;
        if (pendingMetaNode_ > 0) {
            user_.pendingMetaNode = 0;
            _safeMetaNodeTransfer(msg.sender, pendingMetaNode_);
        }

        user_.finishedMetaNode = (user_.stAmount * pool_.accMetaNodePerST) / (1 ether);

        emit Claim(msg.sender, _pid, pendingMetaNode_);

    }

}