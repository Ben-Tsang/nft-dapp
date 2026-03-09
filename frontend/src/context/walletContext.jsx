// /contexts/WalletContext.jsx
import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  useRef,
} from "react"; // 新增useRef
import { connectWallet, getBalance, getProvider } from "@/js/utils/wallet";
import { ethers } from "ethers";

const WalletContext = createContext();

export function WalletProvider({ children }) {
  const [walletData, setWalletData] = useState(null);
  const [contracts, setContracts] = useState({});
  const [contractsRO, setContractsRO] = useState({});
  const [loading, setLoading] = useState(false);
  const [lastConfig, setLastConfig] = useState(null);

  // 新增：用ref缓存最新状态（绕过React状态更新延迟）
  const walletDataRef = useRef(null);
  const contractsRef = useRef({});
  const contractsRORef = useRef({});

  // 监听状态变化，同步更新ref
  useEffect(() => {
    walletDataRef.current = walletData;
    contractsRef.current = contracts;
    contractsRORef.current = contractsRO;
  }, [walletData, contracts, contractsRO]);

  // 工具：根据配置批量挂载合约
  const mountContracts = async (provider, signer, contractsConfig) => {
    if (!contractsConfig || typeof contractsConfig !== "object")
      return [{}, {}];

    const toContract = (runner, { address, abi }) => {
      if (!address || !abi) return undefined;
      const actualABI = Array.isArray(abi) ? abi : abi.abi || abi;
      return new ethers.Contract(address, actualABI, runner);
    };

    const names = Object.keys(contractsConfig);
    const signerMap = {};
    const roMap = {};

    for (const name of names) {
      const cfg = contractsConfig[name];
      signerMap[name] = toContract(signer, cfg);
      roMap[name] = toContract(provider, cfg);
    }
    return [signerMap, roMap];
  };

  // 连接方法保持不变
  const connect = async (targetChainId, contractsConfig) => {
    setLoading(true);
    try {
      const data = await connectWallet(targetChainId);
      const balance = await getBalance(data.provider, data.signer);

      const [signerMap, roMap] = await mountContracts(
        data.provider,
        data.signer,
        contractsConfig
      );

      setWalletData({
        ...data,
        balance: Number(balance).toFixed(4),
      });
      setContracts(signerMap);
      setContractsRO(roMap);
      setLastConfig(contractsConfig);

      console.log("🎉 已连接钱包并挂载合约：", Object.keys(signerMap));
      // 新增：返回最新的signer和合约（关键）
      return {
        signer: data.signer,
        contracts: signerMap,
        address: data.address,
      };
    } catch (e) {
      console.error("连接失败:", e);
      throw e;
    } finally {
      setLoading(false);
    }
  };

  const disconnect = () => {
    setWalletData(null);
    setContracts({});
    setContractsRO({});
    setLastConfig(null);
    // 同步更新ref
    walletDataRef.current = null;
    contractsRef.current = {};
    contractsRORef.current = {};
  };

  // 新增：获取最新的signer（绕过React状态延迟）
  const getLatestSigner = () => walletDataRef.current?.signer;
  // 新增：获取最新的合约（绕过React状态延迟）
  const getLatestContract = (name) => contractsRef.current?.[name];
  const getLatestContractRO = (name) => contractsRORef.current?.[name];

  // 原有获取合约方法保持不变（兼容旧逻辑）
  const getContract = (name) => contracts?.[name];
  const getContractRO = (name) => contractsRO?.[name];

  // 监听链与账号变化逻辑保持不变
  useEffect(() => {
    if (!window.ethereum) return;

    const handleChainChanged = async () => {
      try {
        const provider = getProvider();
        const signer = await provider.getSigner();
        const address = await signer.getAddress();
        const network = await provider.getNetwork();

        const [signerMap, roMap] = await mountContracts(
          provider,
          signer,
          lastConfig || {}
        );
        const bal = await getBalance(provider, signer);

        setWalletData({
          provider,
          signer,
          address,
          chainId: network.chainId,
          networkName: String(network.chainId),
          balance: Number(bal).toFixed(4),
        });
        setContracts(signerMap);
        setContractsRO(roMap);
      } catch (e) {
        console.error("链切换重挂载失败：", e);
      }
    };

    // /contexts/WalletContext.jsx 中的handleAccountsChanged方法
    const handleAccountsChanged = async (accounts) => {
      if (!accounts || accounts.length === 0) {
        // 钱包断开连接：清空所有状态+token
        localStorage.removeItem("token");
        localStorage.removeItem("connectedAddress");
        disconnect();
        // 只提示一次，且提示语更友好
        alert("钱包已断开连接，请重新连接");
        return;
      }

      // 地址切换：清空旧token+旧状态，保留合约配置
      const newAddress = accounts[0];
      const oldAddress = walletDataRef.current?.address;

      // 只有地址真的变化了，才执行清空操作
      if (newAddress !== oldAddress) {
        localStorage.removeItem("token");
        localStorage.removeItem("connectedAddress");
        console.log(`钱包地址已切换：${oldAddress} → ${newAddress}`);

        // 重新挂载合约
        await handleChainChanged();

        // 只弹这一个提示，去掉其他重复提示
        alert("钱包地址已切换，请重新连接钱包完成登录");

        // 核心：1行代码解决所有问题——强制刷新页面
        window.location.reload();
      }
    };

    window.ethereum.on("chainChanged", handleChainChanged);
    window.ethereum.on("accountsChanged", handleAccountsChanged);

    return () => {
      try {
        window.ethereum.removeListener("chainChanged", handleChainChanged);
        window.ethereum.removeListener(
          "accountsChanged",
          handleAccountsChanged
        );
      } catch (_) {}
    };
  }, [lastConfig]);

  const value = useMemo(
    () => ({
      walletData,
      loading,
      connect,
      disconnect,
      getContract,
      getContractRO,
      getLatestSigner, // 新增：暴露获取最新signer的方法
      getLatestContract, // 新增：暴露获取最新合约的方法
      getLatestContractRO, // 新增：暴露获取最新只读合约的方法
    }),
    [walletData, loading, contracts, contractsRO]
  );

  return (
    <WalletContext.Provider value={value}>{children}</WalletContext.Provider>
  );
}

export const useWallet = () => {
  const context = useContext(WalletContext);
  if (!context) {
    throw new Error("useWallet 必须在 WalletProvider 内部使用");
  }
  return context;
};
