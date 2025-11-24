import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("AuctionUpgrade", (m) => {
  // Deploy the new implementation (AuctionV2)
  const auctionV2Impl = m.contract("AuctionV2");

  // Return the implementation - upgrade will be done separately
  return { auctionV2Impl };
});