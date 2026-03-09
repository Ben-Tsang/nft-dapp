import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";
// 导入NFT部署模块（前提是NFTModule和MarketModule在同一目录）
import NFTModule from "./NFT";

export default buildModule("NFTMarketModule", (m) => {
  // 先获取已部署的NFT合约实例，无需手动填地址
  const nftContract = m.useModule(NFTModule).nftContract;
  // 传入NFT合约实例（Ignition会自动解析为地址）
  const marketContract = m.contract("NFTMarket", [nftContract]);
  return { marketContract };
});