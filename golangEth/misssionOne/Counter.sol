// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

contract Counter {
    int256 public _nums;

    function increment() public {
        _nums += 1;
    }

    function decrement() public {
        _nums -= 1;
    }
    
}