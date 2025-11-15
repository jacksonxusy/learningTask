// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

contract Voting {
    struct Candidate {
        uint id;
        string name;
    }
    Candidate[] public candidates;
    mapping(address => uint256) public count;
    address public owner;
    address[] public voters;


    function vote(address _candidateAddress) public {
        if (count[_candidateAddress] == 0) {
            count[_candidateAddress] = 1;
        } else {
            count[_candidateAddress] += 1;
        }
    }

    function getVotes(address _candidateAddress) public view returns (uint256) {
        return count[_candidateAddress];
    }

    function resetVotes() public {
        for (uint i = 0; i < voters.length; i++) {
            count[voters[i]] = 0;
        }
    }
}

contract Algorithm {

    mapping(string => uint256) public map;
    uint64[13] public values = [1000,900,500,400,100,90,50,40,10,9,5,4,1];
    string[13] public symbols = ["M","CM","D","CD","C","XC","L","XL","X","IX","V","IV","I"];

    // reverse string
    function reverseString(string memory _input) public pure returns (string memory) {
        bytes memory _bytes = bytes(_input);
        uint256 length = bytes(_input).length;
        for (uint256 i = 0; i < length / 2; i++) {
            (_bytes[i], _bytes[length - i - 1]) = (_bytes[length - i - 1], _bytes[i]);
        }
        return string(_bytes);
    }

    // roman-to-integer
    function romanToInteger(string memory _input) public pure returns (uint256) {
        uint256 result = 0;
        bytes memory b = bytes(_input);
        for (uint256 i = 0; i < b.length; i++) {
            uint256 value = _charValue(b[i]);
            uint256 nextValue = (i < b.length - 1) ? _charValue(b[i + 1]) : 0;
            if (value < nextValue) {
                result -= value;
            } else {
                result += value;
            }
        }
        return result;
    }

    function _charValue(bytes1 c) internal pure returns (uint256) {
        if (c == "I") return 1;
        if (c == "V") return 5;
        if (c == "X") return 10;
        if (c == "L") return 50;
        if (c == "C") return 100;
        if (c == "D") return 500;
        if (c == "M") return 1000;
        revert("Invalid Roman character");
    }

    function integerToRoman(uint256 _input) public view returns(string memory) {
        string memory result = "";
        for (uint256 i = 0; i < values.length; i++) {
            while (_input >= values[i]) {
                result = string.concat(result, symbols[i]);
                _input -= values[i];
            }
        }
        return result;
    }

    function mergeSortedArray(uint256[] memory nums1, uint m, uint256[] memory nums2, uint n) public pure returns (uint256[] memory) {
        uint256[] memory result = new uint256[](m + n);
        uint j = 0;
        uint k = 0;
        for (uint i = 0; i < m + n; i++) {
            if (j == m) {
                result[i] = nums2[k];
                k++;
                continue;
            }
            if (k == n) {
                result[i] = nums1[j];
                j++;
                continue;
            }

            if (nums1[j] <= nums2[k]) {
                result[i] = nums1[j];
                j++;
            } else if (nums1[j] == nums2[k]) {
                result[i] = nums2[k];
                k++;
                j++;
            }
        }
       return result;
    }

    function binarySearch(int256[] memory nums, int target) public pure returns (int) {
        uint left = 0;
        uint right = nums.length - 1;
        while (left <= right) {
            uint mid = (left + right) / 2;
            if (nums[mid] == target) {
                return int(mid);
            } else if (nums[mid] < target) {
                left = mid + 1;
            } else {
                right = mid - 1;
            }
        }
        return -1;
    }


}
