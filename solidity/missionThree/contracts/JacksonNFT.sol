// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/token/ERC721/ERC721Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";



contract JacksonNFT is Initializable, ERC721Upgradeable, OwnableUpgradeable, UUPSUpgradeable {

    uint256 public nextTokenId;

    function initialize() public initializer {
        __ERC721_init("JacksonNFT", "JXN");
        __UUPSUpgradeable_init();
        __Ownable_init();

    }

    function mint(address to) public onlyOwner {
        _safeMint(to, nextTokenId);
        nextTokenId++;
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {
      
    }

}

