// contract.js (整合合约地址 + ABI + 工具函数)

// ======================== 导入ABI文件 ========================
import NftAbi from "@/abi/NFT.json";
import MarketAbi from "@/abi/NFTMarket.json";

// ======================== 核心配置：合约地址 + ABI 映射 ========================
// 按合约名分组，每个合约名下包含地址映射和ABI
export const CONTRACT_CONFIGS = {
  // NFT合约配置
  nft: {
    abi: NftAbi, // 关联NFT合约的ABI
    addresses: {
      "0xaa36a7": "0x3c7399e7d783709f0Acc1A6D811b80B37f2Cb031", // Sepolia测试网
      "0x7a69": "0x5FbDB2315678afecb367f032d93F642f64180aa3", // Hardhat Network
    },
  },
  // 市场合约配置
  market: {
    abi: MarketAbi, // 关联Market合约的ABI
    addresses: {
      "0xaa36a7": "0x3c04de0815bd05640502937EBaCB837BD59B810E", // Sepolia测试网
      "0x7a69": "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512", // Hardhat Network
    },
  },
};

// ======================== 工具函数：标准化ChainId ========================
/**
 * 标准化ChainId格式（统一转为小写十六进制字符串）
 * @param {string|number} chainId - 原始链ID
 * @returns {string} 标准化后的链ID
 */
function normalizeChainId(chainId) {
  return typeof chainId === "number"
    ? "0x" + chainId.toString(16)
    : chainId.toLowerCase();
}

// ======================== 工具函数：获取合约地址（兼容原有调用） ========================
/**
 * 根据链ID和合约名获取对应的合约地址
 * @param {string|number} chainId - 链ID（支持数字/十六进制字符串）
 * @param {string} contractName - 合约名称（如 "nft" / "market"）
 * @returns {string|null} 合约地址
 */
export function getContractAddress(chainId, contractName) {
  const normalizedChainId = normalizeChainId(chainId);

  // 检查合约配置是否存在
  const contractConfig = CONTRACT_CONFIGS[contractName];
  if (!contractConfig) {
    console.warn(`[合约配置] 未找到名为 "${contractName}" 的合约配置`);
    return null;
  }

  // 检查该链下的地址是否存在
  const address = contractConfig.addresses[normalizedChainId];
  if (!address) {
    console.warn(
      `[合约配置] 链ID "${normalizedChainId}" 上未配置 "${contractName}" 的合约地址`,
    );
    return null;
  }

  // 验证地址格式
  if (!/^0x[a-fA-F0-9]{40}$/.test(address)) {
    console.warn(
      `[合约配置] "${contractName}" 在链 "${normalizedChainId}" 上的地址格式错误：${address}`,
    );
    return null;
  }

  return address;
}

// ======================== 新增工具函数：获取合约完整配置（地址+ABI） ========================
/**
 * 根据链ID和合约名获取完整的合约配置（地址+ABI）
 * @param {string|number} chainId - 链ID
 * @param {string} contractName - 合约名称
 * @returns {Object|null} { address, abi } 或 null
 */
export function getContractConfig(chainId, contractName) {
  const address = getContractAddress(chainId, contractName);
  if (!address) return null;

  return {
    address,
    abi: CONTRACT_CONFIGS[contractName].abi,
  };
}

// ======================== 新增工具函数：获取某链下所有合约的完整配置 ========================
/**
 * 获取指定链下所有已配置合约的完整信息（适配connect函数入参）
 * @param {string|number} chainId - 链ID
 * @returns {Object} { nft: { address, abi }, market: { address, abi }, ... }
 */
export function getAllContractConfigsForChain(chainId) {
  const normalizedChainId = normalizeChainId(chainId);
  const result = {};

  // 遍历所有合约，组装地址+ABI
  Object.entries(CONTRACT_CONFIGS).forEach(([contractName, config]) => {
    const address = config.addresses[normalizedChainId];
    if (address && /^0x[a-fA-F0-9]{40}$/.test(address)) {
      result[contractName] = {
        address,
        abi: config.abi,
      };
    }
  });

  return result;
}

// ======================== 扩展工具函数（保留原有功能） ========================
/**
 * 获取某条链上所有已配置的合约地址（仅返回地址）
 * @param {string|number} chainId - 链ID
 * @returns {Object} { 合约名: 地址 } 的映射对象
 */
export function getAllContractsForChain(chainId) {
  const normalizedChainId = normalizeChainId(chainId);
  const result = {};

  Object.entries(CONTRACT_CONFIGS).forEach(([contractName, config]) => {
    const address = config.addresses[normalizedChainId];
    if (address && /^0x[a-fA-F0-9]{40}$/.test(address)) {
      result[contractName] = address;
    }
  });

  return result;
}
