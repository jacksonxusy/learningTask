import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("JacksonNFT", (m) => {
  const nft = m.contract("JacksonNFT");

  // Initialize the contract
  m.call(nft, "initialize", []);

  return { nft };
});