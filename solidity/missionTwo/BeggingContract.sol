// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

contract BeggingContract {
    mapping(address => uint) public begs;
    address public owner;

    event Donate(address from, uint amount);

    constructor() {
        owner = msg.sender;
    }

    function donate() public payable {
        require(msg.value > 0, "You need to send some value");
        begs[msg.sender] += msg.value;
        emit Donate(msg.sender, msg.value);
    }

    modifier onlyOwner() {
        require(msg.sender == owner, "You are not the owner of this contract");
        _;
    }

    function withdraw() external onlyOwner payable {
        (bool success, ) = msg.sender.call{value: address(this).balance}("");
        require(success, "Transfer failed.");
    }

    function getDonation(address donor) public view returns (uint) {
        return begs[donor];
    }


}
