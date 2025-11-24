import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { network } from "hardhat";
import MyNFTModule from "../ignition/modules/JacksonNFT.js";
import AuctionModule from "../ignition/modules/Auction.js";

describe("NFT Auction with Ignition", async function () {
  const { viem, ignition } = await network.connect();
  const testClient = await viem.getTestClient();
  const [owner, seller, bidder] = await viem.getWalletClients();

  it("should deploy and run full auction", async function () {
    // Deploy contracts using AuctionModule which includes JacksonNFTModule
    const { auction, nft } = await ignition.deploy(AuctionModule);

    // Verify the auction is looking at the correct NFT contract
    const auctionNftContract = await auction.read.nftContract();
    console.log("Expected NFT address:", nft.address);
    console.log("Actual NFT address in auction:", auctionNftContract);

    // Mint NFT to seller (this will be tokenId 0)
    await nft.write.mint([seller.account.address]);

    await nft.write.setApprovalForAll([auction.address, true], { account: seller.account });

    // Check what token was minted
    const tokenId = 0n; // First minted token

    // Create auction (5 minutes, accepts ETH)
    await auction.write.createAuction([tokenId, 5n, true], { account: seller.account });

    // Bid with 1 ETH
    await auction.write.bidWithETH([tokenId], {
      account: bidder.account,
      value: 1000000000000000000n // 1 ETH in wei
    });

    // Fast forward time by 301 seconds (5 minutes + 1 second)
    await testClient.increaseTime({ seconds: 301 });
    await testClient.mine({ blocks: 1 });

    // End auction
    await seller.writeContract({
      address: auction.address as `0x${string}`,
      abi: auction.abi,
      functionName: "endAuction" as any,
      args: [tokenId]
    });

    // Check NFT ownership transferred to bidder
    const nftOwner = await nft.read.ownerOf([tokenId]);
    assert.equal(nftOwner.toLowerCase(), bidder.account.address.toLowerCase());
  });
});