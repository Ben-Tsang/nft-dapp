// networks.js (补充后完整代码)
export const SUPPORTED_NETWORKS = {
  // Sepolia测试链
  "0xaa36a7": {
    chainId: "0xaa36a7",
    chainName: "Sepolia Testnet",
    rpcUrls: ["https://sepolia.infura.io/v3/你的Infura Project ID"], // 替换成你的ID
    blockExplorerUrls: ["https://sepolia.etherscan.io"],
    nativeCurrency: {
      name: "Sepolia Ether",
      symbol: "ETH",
      decimals: 18,
    },
  },
  "0x7A69": {
    chainId: "0x7A69",
    chainName: "Hardhat Network",
    rpcUrls: ["http://127.0.0.1:8545"],
    blockExplorerUrls: [],
    nativeCurrency: {
      name: "Ether",
      symbol: "ETH",
      decimals: 18,
    },
  },
};

// ========== 新增工具函数（核心） ==========
// 1. 默认选中的链ID（可直接改这里）
export const DEFAULT_CHAIN_ID = "0x7A69";

// 2. 将对象转为数组（方便组件遍历渲染按钮）
export const getNetworkList = () => {
  // 把SUPPORTED_NETWORKS的value转成数组，保留所有字段
  return Object.values(SUPPORTED_NETWORKS);
};

// 3. 根据链ID获取单个链的配置
export const getNetworkByChainId = (chainId) => {
  return SUPPORTED_NETWORKS[chainId] || null;
};

// 4. 生成prompt提示文本（用户友好型：1→Hardhat，2→Sepolia）
export const getNetworkPromptText = () => {
  const networkList = getNetworkList();
  // 给每个链分配简易编号（1、2），用户输入编号即可
  return networkList
    .map((network, index) => `${index + 1}: ${network.chainName}`)
    .join("\n");
};

// 5. 根据用户输入的简易编号（1、2）获取对应链配置
export const getNetworkByPromptIndex = (inputIndex) => {
  const networkList = getNetworkList();
  // 输入的是字符串（如"1"），转成数字后减1对应数组索引
  const index = Number(inputIndex) - 1;
  return networkList[index] || null;
};
