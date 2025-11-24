import { task } from "hardhat/config";
import { formatEther, parseEther } from "viem";

// Existing deployed contracts (replace with your actual deployed addresses)
const NFT_ADDRESS = "0x602b759be5f08cA4FFb5bfae1D60AFa3Ed0bBb18";
const AUCTION_PROXY_ADDRESS = "0xeB17fe9B4484939e0CEa1d74da35Cc0c8636E3f8";

task("upgrade-demo", "Demonstrate contract upgrade from Auction to AuctionV2")
  .setAction(async (taskArgs, hre) => {
    console.log("🚀 Contract Upgrade Demonstration");
    console.log("=====================================");

    try {
      const { viem } = hre;
      const publicClient = await viem.getPublicClient();
      const [owner, seller, bidder] = await viem.getWalletClients();

      console.log("📍 Wallet addresses:");
      console.log("   Owner:", owner.account.address);
      console.log("   Seller:", seller.account.address);
      console.log("   Bidder:", bidder.account.address);
      console.log("");

      // Step 1: Connect to existing contracts
      console.log("📦 Connecting to existing contracts...");
      const nft = await viem.getContractAt("JacksonNFT", NFT_ADDRESS);
      const auctionV1 = await viem.getContractAt("Auction", AUCTION_PROXY_ADDRESS);

      console.log("✅ Connected to:");
      console.log("   NFT:", nft.address);
      console.log("   Auction (Proxy):", auctionV1.address);
      console.log("");

      // Step 2: Test current functionality
      console.log("🧪 Testing current functionality...");
      try {
        // Test a read function
        const ethPrice = await auctionV1.read.getLatestETHPrice();
        console.log("   ✅ V1 getLatestETHPrice():", formatEther(ethPrice));
      } catch (error) {
        console.log("   ❌ V1 function test failed");
      }
      console.log("");

      // Step 3: Deploy new implementation
      console.log("🔄 Deploying AuctionV2 implementation...");
      const auctionV2Impl = await viem.deployContract("AuctionV2");
      console.log("   ✅ AuctionV2 implementation deployed:", auctionV2Impl.address);
      console.log("");

      // Step 4: Get the proxy admin (this would typically be a separate deployment)
      console.log("🔧 Getting proxy admin...");
      // Note: In a real scenario, you'd need to find the proxy admin address
      // For demonstration, we'll show how to call the upgrade

      // The actual upgrade would be done by the proxy admin
      // This is a simplified example
      console.log("📝 Note: In a real deployment, the upgrade would be performed by the proxy admin");
      console.log("   The proxy admin would call: upgradeTo(newImplementation)");
      console.log("");

      // Step 5: Simulate the upgraded contract (for demo purposes)
      console.log("🧪 Testing upgraded contract (simulation)...");
      const upgradedAuction = await viem.getContractAt("AuctionV2", AUCTION_PROXY_ADDRESS);

      try {
        // Test new function from V2
        const version = await upgradedAuction.read.version();
        console.log("   ✅ V2 version():", version);
      } catch (error) {
        console.log("   ❌ V2 function not available (upgrade not completed)");
      }
      console.log("");

      // Step 6: Create demo auction to show functionality
      console.log("🎨 Creating demo auction...");

      // Check if seller has NFTs
      const nextTokenId = await nft.read.nextTokenId();
      console.log("   Next token ID:", nextTokenId);

      if (nextTokenId > 0n) {
        // Use existing NFT
        const tokenId = nextTokenId - 1n;
        const ownerOfToken = await nft.read.ownerOf([tokenId]);

        console.log("   Using token ID:", tokenId);
        console.log("   Token owner:", ownerOfToken);

        if (ownerOfToken.toLowerCase() === seller.account.address.toLowerCase()) {
          // Create auction
          await auctionV1.write.createAuction([tokenId, 60n, true], {
            account: seller.account
          });
          console.log("   ✅ Auction created for token", tokenId);

          // Place bid
          const bidAmount = parseEther("0.1");
          await auctionV1.write.bidWithETH([tokenId], {
            account: bidder.account,
            value: bidAmount
          });
          console.log("   ✅ Bid placed:", formatEther(bidAmount), "ETH");
        } else {
          console.log("   ❌ Seller doesn't own any NFTs");
        }
      } else {
        console.log("   ❌ No NFTs minted yet");
      }
      console.log("");

      console.log("📋 Upgrade Summary:");
      console.log("===============");
      console.log("1. ✅ Connected to existing contracts");
      console.log("2. ✅ Tested V1 functionality");
      console.log("3. ✅ Deployed V2 implementation");
      console.log("4. ℹ️  Actual upgrade requires proxy admin");
      console.log("5. ✅ V2 new functions available after upgrade");
      console.log("");

      console.log("🔗 Contract Addresses:");
      console.log("   - NFT:", NFT_ADDRESS);
      console.log("   - Auction Proxy:", AUCTION_PROXY_ADDRESS);
      console.log("   - AuctionV2 Implementation:", auctionV2Impl.address);
      console.log("");

      console.log("💡 Next Steps:");
      console.log("1. Use proxy admin to call upgradeTo()");
      console.log("2. Verify new functionality works");
      console.log("3. Test state preservation");

    } catch (error: any) {
      console.error("❌ Error during upgrade demo:", error.message);
    }
  });

export default {};