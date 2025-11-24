import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";
import jacksonNFTModule from "./JacksonNFT.js";

const SEPOLIA_ETH_USD = "0x694AA1769357215DE4FAC081bf1f309aDC325306";

export default buildModule("AuctionModule", (m) => {
  // 先获取已经部署的 NFT
  const { nft } = m.useModule(jacksonNFTModule);

  const auction = m.contract("Auction");

  // 初始化
  m.call(auction, "initialize", [
    nft,
    SEPOLIA_ETH_USD,
    "0x0000000000000000000000000000000000000000",
  ]);

  return { auction, nft };
});