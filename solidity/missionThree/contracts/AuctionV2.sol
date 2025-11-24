// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./Auction.sol";

contract AuctionV2 is Auction {

    // New function that didn't exist in V1
    function version() public pure returns (string memory) {
        return "2.0 - Now";
    }

}