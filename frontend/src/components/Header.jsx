import React, { useState, useCallback } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { formatAddress } from "@/js/utils/format";
import { getAllContractConfigsForChain } from "@/js/utils/contract";
import {
  FaBahtSign,
  FaAward,
  FaWallet,
  FaCheck,
  FaCopy,
  FaArrowRight,
} from "react-icons/fa6";
import authApi from "@/js/api/auth";
import { useWallet } from "@/context/WalletContext";
import { ethers } from "ethers";

// ========== 核心修改：导入你现有的networks.js ==========
import {
  SUPPORTED_NETWORKS,
  DEFAULT_CHAIN_ID,
  getNetworkList,
  getNetworkByChainId,
  getNetworkPromptText,
  getNetworkByPromptIndex,
} from "@/js/utils/network.js"; // 路径要和你的实际文件一致

function Header() {
  // ========== 用networks.js的默认链ID ==========
  const [selectedChainId, setSelectedChainId] = useState(DEFAULT_CHAIN_ID);
  const [connecting, setConnecting] = useState(false);
  const [copied, setCopied] = useState(false);
  const navigate = useNavigate();
  const { pathname } = useLocation();

  const {
    walletData,
    loading,
    connect,
    disconnect,
    getLatestSigner,
    getLatestContract,
  } = useWallet();

  const address = walletData?.address;
  const balance = walletData?.balance;
  const networkName = walletData?.networkName;

  const linkTo = (path) => {
    navigate(path);
  };

  const handleWalletConnect = useCallback(async () => {
    if (connecting || loading) return;
    setConnecting(true);
    try {
      localStorage.removeItem("token");
      console.log("选择的链：", selectedChainId);

      const nonceRes = await authApi.nonce();

      const nonce = nonceRes?.nonce || nonceRes;
      if (!nonce) throw new Error("获取nonce失败");

      const contractConfigs = getAllContractConfigsForChain(selectedChainId);

      // ✅ 格式化打印（带缩进，清晰）
      console.log(`===== 链ID ${selectedChainId} 的合约地址 =====`);
      if (contractConfigs && typeof contractConfigs === "object") {
        Object.entries(contractConfigs).forEach(
          ([contractName, contractInfo]) => {
            const address = contractInfo?.address || "无地址";
            console.log(`${contractName}: ${address}`);
          },
        );
      } else {
        console.log("链配置为空或格式错误");
      }
      console.log("==========================================");

      const connectResult = await connect(selectedChainId, contractConfigs);

      let currentSigner = connectResult?.signer || getLatestSigner();
      let nftContract =
        connectResult?.contracts?.nft || getLatestContract("nft");
      let marketContract =
        connectResult?.contracts?.market || getLatestContract("market");
      let signerAddress =
        connectResult?.address || (await currentSigner.getAddress());

      if (!currentSigner || !nftContract || !marketContract) {
        throw new Error("合约或钱包初始化失败");
      }

      const signature = await currentSigner.signMessage(nonce);
      const loginRes = await authApi.login({
        address: signerAddress,
        signature: signature,
        nonce: nonce,
      });

      localStorage.setItem("token", loginRes.token);
      localStorage.setItem("connectedAddress", signerAddress);
      document.getElementById("connectWalletModal").close();

      // ========== 从networks.js获取链名称 ==========
      const selectedNetwork = getNetworkByChainId(selectedChainId);
      alert("连接成功！当前链：" + (selectedNetwork?.chainName || "未知链"));
    } catch (err) {
      console.error("连接失败", err);
      alert(`连接失败: ${err.message}`);
    } finally {
      setConnecting(false);
    }
  }, [
    connecting,
    loading,
    connect,
    selectedChainId,
    getLatestSigner,
    getLatestContract,
  ]);

  const copyAddress = async () => {
    if (!address) return;
    try {
      await navigator.clipboard.writeText(address);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("复制失败", err);
    }
  };

  const handleDisconnect = () => {
    disconnect();
    localStorage.removeItem("token");
  };

  return (
    <div className="fixed top-0 left-0 h-20 w-full pt-3 bg-[#1d1d3b] shadow-[0_2px_10px_rgba(0,0,0,0.3)] z-50 ">
      <div className="max-w-7xl min-w-7xl flex justify-between items-center mx-auto ">
        <div className="ml-5 text-3xl font-bold">
          <span className="text-[#6c5ce7]">NFT</span>
          <span className="text-white">Market</span>
        </div>

        <div role="tablist" className="tabs text-white font-bold">
          <a
            className={`tab ${pathname === "/" ? "tab-active text-blue-400" : ""}`}
            onClick={() => linkTo("/")}
          >
            首页
          </a>
          <a
            className={`tab ${pathname === "/discovery" ? "tab-active text-blue-400" : ""}`}
            onClick={() => linkTo("/discovery")}
          >
            发现
          </a>
          {/* <a
            className={`tab ${pathname === "/rank" ? "tab-active text-blue-400" : ""}`}
            onClick={() => linkTo("/rank")}
          >
            排行榜
          </a> */}
          <a
            className={`tab ${pathname === "/mint" ? "tab-active text-blue-400" : ""}`}
            onClick={() => linkTo("/mint")}
          >
            创建
          </a>
        </div>

        <div className="flex items-center">
          {address ? (
            <div className="flex items-center gap-4 text-white mr-5">
              {/* ========== 链标签：适配networks.js ========== */}
              <div
                className="badge badge-neutral p-5 cursor-pointer"
                onClick={() => {
                  // 生成友好的提示文本（1: Hardhat Network, 2: Sepolia Testnet）
                  const inputIndex = prompt(
                    `请选择链：\n${getNetworkPromptText()}`,
                    "1", // 默认输入1（对应Hardhat）
                  );
                  // 根据输入的编号获取链配置
                  const targetNetwork = getNetworkByPromptIndex(inputIndex);
                  if (targetNetwork) {
                    setSelectedChainId(targetNetwork.chainId);
                  }
                }}
              >
                {/* 优先显示钱包返回的名称，否则从networks.js获取 */}
                {networkName ||
                  getNetworkByChainId(selectedChainId)?.chainName ||
                  "未知链"}
                <FaArrowRight className="text-green-400 ml-2" />
              </div>

              <div
                className="tooltip tooltip-bottom"
                data-tip="点击复制完整地址"
              >
                <div
                  className="flex items-center cursor-pointer bg-white/10 rounded-lg px-3 py-2 hover:bg-white/20 transition-colors"
                  onClick={copyAddress}
                >
                  <FaWallet className="mr-2 text-xl" />
                  <span className="font-mono">{formatAddress(address)}</span>
                  <div className="ml-2">
                    {copied ? (
                      <FaCheck className="text-green-400" />
                    ) : (
                      <FaCopy className="text-gray-300" />
                    )}
                  </div>
                </div>
              </div>
            </div>
          ) : (
            <div
              className="btn btn-primary mr-6"
              onClick={() =>
                document.getElementById("connectWalletModal").showModal()
              }
            >
              {loading ? "连接中..." : "连接钱包"}
            </div>
          )}

          <div className="dropdown dropdown-start">
            <div className="avatar m-1" tabIndex={0}>
              <div className="ring-primary ring-offset-base-100 w-12 rounded-full ring-2 ring-offset-2 cursor-pointer">
                <img src="https://img.daisyui.com/images/profile/demo/spiderperson@192.webp" />
              </div>
            </div>
            <ul
              tabIndex="-1"
              className="dropdown-content menu bg-base-100 rounded-box z-1 mt-3 w-52 p-2 shadow-sm"
            >
              <li>
                <a>余额: {balance} ETH</a>
              </li>
              <li>
                <a onClick={() => linkTo("/operationRecord")}>操作记录</a>
              </li>
              <li>
                <a onClick={() => linkTo("/myNft")}>我的收藏</a>
              </li>
              <li>
                <a>设置</a>
              </li>
              <li>
                <a onClick={handleDisconnect}>断开连接</a>
              </li>
            </ul>
          </div>
        </div>
      </div>

      {/* 钱包弹窗 */}
      <dialog id="connectWalletModal" className="modal">
        <div className="modal-box bg-[#1e1e2e] text-white">
          <form method="dialog">
            <button className="absolute right-5 top-5 hover:font-bold">
              ✕
            </button>
          </form>
          <h3 className="font-bold text-xl">连接钱包</h3>

          {/* ========== 链选择按钮：遍历networks.js的数组 ========== */}
          <div className="mt-4">
            <p className="mb-2">选择链：</p>
            <div className="grid grid-cols-2 gap-2">
              {getNetworkList().map((network) => (
                <button
                  key={network.chainId} // 用链ID作为唯一key
                  className={`p-3 rounded ${selectedChainId === network.chainId ? "bg-purple-600" : "bg-white/10"}`}
                  onClick={() => setSelectedChainId(network.chainId)}
                >
                  {network.chainName}
                </button>
              ))}
            </div>
          </div>

          <div className="flex flex-col pt-6">
            <div
              onClick={handleWalletConnect}
              className={`w-full h-30 rounded-lg bg-[#ffffff0d] p-5 flex items-center justify-center cursor-pointer ${connecting ? "opacity-50" : ""}`}
            >
              <div className="w-10 h-10 rounded-lg bg-purple-700 flex justify-center items-center mr-5 font-bold">
                M
              </div>
              <div>
                <div className="font-bold text-lg">连接 MetaMask</div>

                {connecting && (
                  <div className="text-sm text-blue-400 mt-1">连接中...</div>
                )}
              </div>
            </div>
          </div>
        </div>
      </dialog>
    </div>
  );
}

export default Header;
