import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("NFT", (m) => {
  const nftContract = m.contract("NFT");


  return { nftContract };
});
