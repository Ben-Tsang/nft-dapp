// /js/wallet.js
import { ethers } from "ethers";
import { SUPPORTED_NETWORKS } from "./network";

// ====================================================================================
// 连接钱包：仅处理 provider/signer/链切换，不创建合约
export async function connectWallet(targetChainId) {
  initProvider();

  // 切换链
  await switchChain(targetChainId);

  // 请求账户连接
  const accounts = await window.ethereum.request({ method: "eth_requestAccounts" });
  if (!accounts || accounts.length === 0) {
    throw new Error("用户拒绝连接或没有账户");
  }

  const provider = getProvider();
  const signer = await provider.getSigner();
  const address = await signer.getAddress();

  // 链信息
  const network = await provider.getNetwork();
  const chainConfig = SUPPORTED_NETWORKS[targetChainId];
  const networkName = chainConfig?.chainName ?? String(network.chainId);

  return {
    provider,
    signer,
    address,
    chainId: network.chainId,
    networkName,
  };
}

// ====================================================================================
// 初始化 provider
export function initProvider() {
  if (!window.ethereum) {
    throw new Error("请先安装钱包");
  }
  return new ethers.BrowserProvider(window.ethereum);
}

export function getProvider() {
  const provider = initProvider();
  if (!provider) throw new Error("请先安装钱包");
  return provider;
}

export async function getBalance(provider, signer) {
  if (!signer) throw new Error("请先连接钱包");
  const addr = await signer.getAddress();
  const bal = await provider.getBalance(addr);
  return ethers.formatEther(bal); // 返回 string（ETH）
}

export async function sendTransaction(signer, tx) {
  if (!signer) throw new Error("请先连接钱包");
  return await signer.sendTransaction(tx);
}

// ====================================================================================
// 切换链
export async function switchChain(targetChainId) {
  const chainConfig = SUPPORTED_NETWORKS[targetChainId];
  if (!chainConfig) {
    throw new Error(`不支持的链 ${targetChainId}`);
  }
  try {
    await window.ethereum.request({
      method: "wallet_switchEthereumChain",
      params: [{ chainId: targetChainId }],
    });
  } catch (error) {
    if (error.code === 4902) {
      await window.ethereum.request({
        method: "wallet_addEthereumChain",
        params: [
          {
            chainId: targetChainId,
            chainName: chainConfig.chainName,
            rpcUrls: [chainConfig.rpcUrls],
            nativeCurrency: chainConfig.nativeCurrency,
            blockExplorerUrls: chainConfig.explorer ? [chainConfig.explorer] : [],
          },
        ],
      });
    } else {
      throw new Error("切换链失败: " + error.message);
    }
  }
}

// ====================================================================================
// 检查当前链
export async function checkChain(targetChainId) {
  const currentChainId = await window.ethereum.request({ method: "eth_chainId" });
  return currentChainId === targetChainId;
}
