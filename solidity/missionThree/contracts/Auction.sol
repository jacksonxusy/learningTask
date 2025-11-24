// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

import "@openzeppelin/contracts-upgradeable/token/ERC721/IERC721Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/token/ERC721/IERC721ReceiverUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/token/ERC20/IERC20Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";
import "@openzeppelin/contracts-upgradeable/security/ReentrancyGuardUpgradeable.sol";

contract Auction is Initializable, IERC721ReceiverUpgradeable, UUPSUpgradeable, OwnableUpgradeable, ReentrancyGuardUpgradeable {
    IERC721Upgradeable public nftContract;
    AggregatorV3Interface public priceFeed;
    IERC20Upgradeable public paymentToken;

    mapping(uint256 => AuctionItem) public auctions;
    mapping(uint256 => mapping(address => uint256)) public pendingReturns;

    event AuctionCreated(uint256 indexed tokenId, address indexed seller, uint256 endTime, bool isETH);
    event HighestBidIncreased(uint256 indexed tokenId, address indexed bidder, uint256 amount);
    event AuctionEnded(uint256 indexed tokenId, address indexed winner, uint256 amount);



    struct AuctionItem {
        uint256 tokenId;
        address seller;
        uint256 highestBid;
        address highestBidder;
        uint256 endTime;
        bool ended;
        bool isETH;
    }

    function initialize(address _nftContract, address _priceFeed, address _paymentToken) public initializer {
        __Ownable_init();
        __UUPSUpgradeable_init();
        __ReentrancyGuard_init();
        nftContract = IERC721Upgradeable(_nftContract);
        priceFeed = AggregatorV3Interface(_priceFeed);
        paymentToken = IERC20Upgradeable(_paymentToken);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {

    }

    function onERC721Received(address operator, address from, uint256 tokenId, bytes calldata data)
        external
        pure
        override
        returns (bytes4)
    {
        return IERC721ReceiverUpgradeable.onERC721Received.selector;
    }

    function createAuction(uint256 tokenId, uint256 durationInMinutes, bool useETH) external {
        require(nftContract.ownerOf(tokenId) == msg.sender, "You are not the owner of this token");
        require(nftContract.getApproved(tokenId) == address(this) || nftContract.isApprovedForAll(msg.sender, address(this)), "Auction contract is not approved");
        uint256 endTime = block.timestamp + (durationInMinutes * 1 minutes);
        nftContract.safeTransferFrom(msg.sender, address(this), tokenId);

        auctions[tokenId] = AuctionItem({
            tokenId: tokenId,
            seller: msg.sender,
            highestBid:0,
            highestBidder: address(0),
            endTime: endTime,
            ended: false,
            isETH: useETH
        });

        emit AuctionCreated(tokenId, msg.sender, endTime, useETH);
    }

    // use payable, user can send ETH along with the call, no need to specify amount separately
    function bidWithETH(uint256 tokenId) external payable {
        AuctionItem storage auction = auctions[tokenId];

        require(auction.endTime > 0, "Auction does not exist");
        require(block.timestamp < auction.endTime, "Auction already ended");
        require(auction.isETH, "this auction accepts ERC20 only");
        require(msg.value > auction.highestBid, "There already is a higher or equal bid");

        if (auction.highestBidder != address(0)) {
            pendingReturns[tokenId][auction.highestBidder] += auction.highestBid;
        }
        auction.highestBid = msg.value;
        auction.highestBidder = msg.sender;
        emit HighestBidIncreased(tokenId, msg.sender, msg.value);
    }

    function withdraw(uint256 tokenId) external {
        uint256 amount = pendingReturns[tokenId][msg.sender];
        require(amount > 0, "nothing to withdraw");
        pendingReturns[tokenId][msg.sender] = 0;
        payable(msg.sender).transfer(amount);
    }

     // chainlink price feed returns price with 8 decimals
     // returns price with 18 decimals
    function getLatestETHPrice() public view returns (uint256) {
        (, int256 price, , , ) = priceFeed.latestRoundData();
        require(price > 0, "Invalid price");
        return uint256(price) * 1e10;
    }

   
    function ethToUSD(uint256 ethAmount) public view returns (uint256) {
        return (ethAmount * getLatestETHPrice()) / 1e18;
    }

    // assume payment token is 6 decimals usdc.
    function tokenToUSD(uint256 tokenAmount) public pure returns (uint256) {
        return tokenAmount * 1e12;
    }

    function bidWithERC20(uint256 tokenId, uint256 tokenAmount) external {
        AuctionItem storage auction = auctions[tokenId];

        require(block.timestamp < auction.endTime, "auction ended");
        require(!auction.isETH, "this auction accepts ETH only");
        require(tokenAmount > 0, "amount = 0");

        uint256 bidInUSD = tokenToUSD(tokenAmount);
        uint256 currentHighestBidInUSD = auction.highestBidder == address(0) ?
            0 : tokenToUSD(auction.highestBid);
        require(bidInUSD > currentHighestBidInUSD, "There already is a higher or equal bid");

        

        if (auction.highestBidder != address(0)) {
            pendingReturns[tokenId][auction.highestBidder] += auction.highestBid;
        }

        auction.highestBid = tokenAmount;
        auction.highestBidder = msg.sender;

        // checks - effects - interactions pattern
        paymentToken.transferFrom(msg.sender, address(this), tokenAmount);

        emit HighestBidIncreased(tokenId, msg.sender, tokenAmount);
    }

    function endAuction(uint256 tokenId) external nonReentrant {
        AuctionItem storage auction = auctions[tokenId];

        require(block.timestamp >= auction.endTime, "auction not finished");
        require(!auction.ended, "auction already ended");
        auction.ended = true;

        if (auction.highestBidder != address(0)) {
            nftContract.safeTransferFrom(address(this), auction.highestBidder, tokenId);

            if (auction.isETH) {
                payable(auction.seller).transfer(auction.highestBid);
            } else {
                paymentToken.transfer(auction.seller, auction.highestBid);
            }
            
        } else {
            nftContract.safeTransferFrom(address(this), auction.seller, tokenId);
            emit AuctionEnded(tokenId, address(0), 0);
        }
    }
 
}
