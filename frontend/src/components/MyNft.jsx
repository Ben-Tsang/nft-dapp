import React, { useState, useEffect, useCallback } from "react";
import { useWallet } from "@/context/WalletContext";
import { ethers } from "ethers";
import api from "@/js/api/api";
import Pagination from "./Pagination";

function MyNft() {
  const [nfts, setNfts] = useState([]); // 存储 NFT 数据
  const [loading, setLoading] = useState(false); // 加载状态
  const { walletData, getContract, getContractRO } = useWallet();
  const [selectedTokenId, setSelectedTokenId] = useState(""); // 记录选中的nft的tokenId
  const [selectedNftName, setSelectedNftName] = useState("");
  const [selectedNftDescription, setSelectedNftDescription] = useState("");
  const [selectedNftPrice, setSelectedNftPrice] = useState(0);
  const [selectedNftImage, setSelectedNftImage] = useState("");
  const [selectedNftListedAt, setSelectedNftListedAt] = useState("");
  const [targetAddress, setTargetAddress] = useState(""); // 转账目标地址
  const [nftPriceWei, setNftPriceWei] = useState(""); // NFT 上架价格
  // 新增：强制刷新标识（解决缓存问题）
  const [refreshKey, setRefreshKey] = useState(0);

  // 分页相关状态
  const [itemsPerPage, setItemsPerPage] = useState(12); // 每页显示数量，默认12个
  const [currentPage, setCurrentPage] = useState(1);
  const [totalItems, setTotalItems] = useState(0); // 实际总项目数

  // 从 walletData 中解构需要的属性
  const address = walletData?.address;
  const nftRO = getContractRO?.("nft"); // 只读 NFT 合约
  const nft = getContract?.("nft"); // 可写 NFT 合约（转账用）
  const market = getContract?.("market"); // Market 合约（上架用）

  const FILEBASE_GATEWAY = import.meta.env?.VITE_FILEBASE_GATEWAY;

  // 工具函数：检查合约是否有特定方法
  const hasFn = useCallback((contract, name, argc) => {
    try {
      const frags =
        contract?.interface?.fragments?.filter((f) => f.type === "function") ||
        [];
      return frags.some(
        (f) => f.name === name && (f.inputs?.length || 0) === argc,
      );
    } catch {
      return false;
    }
  }, []);

  // 从链上获取 NFT 数据（备用方案）
  const queryMyNftFromChain = useCallback(async () => {
    if (!nftRO || !address) {
      console.log("链上查询：缺少合约或地址");
      return { data: [], total: 0 };
    }

    console.log("从链上查询 NFT...");
    try {
      let tokenIds = [];
      let tokenUris = [];

      if (hasFn(nftRO, "tokensOfOwner", 3)) {
        const [ids, uris] = await nftRO.tokensOfOwner(address, 0, 100);
        tokenIds = ids.map((x) => BigInt(x).toString());
        tokenUris = uris;
      } else if (hasFn(nftRO, "tokensOfOwnerWithUri", 3)) {
        const [ids, uris] = await nftRO.tokensOfOwnerWithUri(address, 0, 100);
        tokenIds = ids.map((x) => BigInt(x).toString());
        tokenUris = uris;
      } else {
        const bal = await nftRO.balanceOf(address);
        const balanceNum = Number(bal);
        if (balanceNum === 0) {
          return { data: [], total: 0 };
        }

        let maxCheck = 500;
        if (hasFn(nftRO, "totalSupply", 0)) {
          try {
            const ts = await nftRO.totalSupply();
            maxCheck = Number(ts) + 5;
          } catch {}
        }

        const found = [];
        for (let tokenId = 1; tokenId <= maxCheck; tokenId++) {
          try {
            const owner = await nftRO.ownerOf(tokenId);
            if (owner?.toLowerCase() === address.toLowerCase()) {
              found.push(tokenId);
              if (found.length >= balanceNum) break;
            }
          } catch {
            // 跳过不存在的 tokenId
          }
        }

        tokenIds = found.map(String);
        const uriPromises = found.map(async (tid) => {
          try {
            return await nftRO.tokenURI(tid);
          } catch {
            return "";
          }
        });
        tokenUris = await Promise.all(uriPromises);
      }

      // 获取 metadata
      const nftPromises = tokenIds.map(async (tid, index) => {
        try {
          let tokenUri = tokenUris[index] || "";
          if (typeof tokenUri === "string" && tokenUri.startsWith("ipfs://")) {
            tokenUri = tokenUri.replace("ipfs://", FILEBASE_GATEWAY);
          }

          let meta = {};
          if (tokenUri) {
            const res = await fetch(tokenUri);
            meta = await res.json();
          }

          let img = meta.image || "";
          if (typeof img === "string" && img.startsWith("ipfs://")) {
            img = img.replace("ipfs://", FILEBASE_GATEWAY);
          }

          return {
            tokenId: String(tid),
            name: meta.name || `NFT #${tid}`,
            description: meta.description || "",
            image:
              img ||
              "https://images.unsplash.com/photo-1620641788421-7a1c342ea42e?ixlib=rb-1.2.1&auto=format&fit=crop&w=500&q=80",
            attributes: meta.attributes || [],
            creator: meta.creator || "Unknown",
            owner: address,
          };
        } catch (error) {
          console.error(`Error fetching metadata for token ${tid}:`, error);
          return {
            tokenId: String(tid),
            name: `NFT #${tid}`,
            description: "Metadata not available",
            image:
              "https://images.unsplash.com/photo-1620641788421-7a1c342ea42e?ixlib=rb-1.2.1&auto=format&fit=crop&w=500&q=80",
            attributes: [],
            creator: "Unknown",
            owner: address,
          };
        }
      });

      const nftData = await Promise.all(nftPromises);
      console.log("链上数据获取成功:", nftData.length, "个 NFT");
      return { data: nftData, total: nftData.length };
    } catch (error) {
      console.error("Error fetching NFTs from chain:", error);
      return { data: [], total: 0 };
    }
  }, [nftRO, address, hasFn]);

  // 主查询函数 - 获取我的 NFT（新增refreshKey依赖，强制刷新）
  const queryMyNft = useCallback(
    async (page = 1) => {
      if (!address) {
        console.log("未连接钱包，跳过查询");
        return;
      }

      setLoading(true);
      console.log(
        `开始查询 NFT，地址: ${address}, 页码: ${page}, 每页: ${itemsPerPage}, 刷新标识: ${refreshKey}`,
      );

      try {
        // 优先从 API 获取数据（添加随机参数，禁用缓存）
        const data = await api.myNft(page, itemsPerPage, {
          params: { _t: Date.now() }, // 新增：禁用浏览器/后端缓存
        });

        let formattedNfts = [];
        let totalFromApi = 0;

        // 解析 API 返回的数据
        if (data && data.records && Array.isArray(data.records)) {
          // 处理 NFT 数据
          formattedNfts = await Promise.all(
            data.records.map(async (item) => {
              let meta = {};
              let tokenUri = item.nft_uri || "";

              // 处理 IPFS 链接
              if (
                typeof tokenUri === "string" &&
                tokenUri.startsWith("ipfs://")
              ) {
                tokenUri = tokenUri.replace("ipfs://", FILEBASE_GATEWAY);
              }

              // 获取 metadata
              if (tokenUri) {
                try {
                  const res = await fetch(tokenUri, {
                    cache: "no-cache", // 新增：禁用fetch缓存
                  });
                  meta = await res.json();
                } catch (error) {
                  console.error("Error fetching metadata for", tokenUri, error);
                }
              }

              // 处理图片链接
              let img = meta.image || "";
              if (typeof img === "string" && img.startsWith("ipfs://")) {
                img = img.replace("ipfs://", FILEBASE_GATEWAY);
              }

              // 处理 nft_uri 的图片链接
              let image = item.nft_uri || "";
              if (image && image.startsWith("ipfs://")) {
                image = image.replace("ipfs://", FILEBASE_GATEWAY);
              }

              return {
                tokenId: item.token_id || item.tokenId || "",
                name:
                  meta.name || item.nft_name || `NFT #${item.token_id || ""}`,
                image:
                  img ||
                  image ||
                  "https://images.unsplash.com/photo-1620641788421-7a1c342ea42e?ixlib=rb-1.2.1&auto=format&fit=crop&w=500&q=80",
                description: meta.description || "",
                creator: meta.creator || item.owner_id || address,
                owner: address,
                listed: item.is_listed || false,
                price: item.price || "0",
                listedAt: item.listed_at,
                unlistedAt: item.unlisted_at, // 新增：补充下架时间
              };
            }),
          );

          // 获取总数
          totalFromApi = data.total || data.count || data.records.length;
          console.log(
            "API 返回总数:",
            totalFromApi,
            "格式化 NFTs:",
            formattedNfts.length,
          );
        }

        // 如果 API 返回空数据或失败，尝试从链上获取
        if (formattedNfts.length === 0) {
          console.log("API返回空数据，尝试从链上获取...");
          const chainResult = await queryMyNftFromChain();
          setNfts(chainResult.data);
          setTotalItems(chainResult.total);
        } else {
          setNfts(formattedNfts);
          setTotalItems(totalFromApi);
        }
      } catch (apiError) {
        console.error("API查询失败，尝试从链上获取:", apiError);
        try {
          const chainResult = await queryMyNftFromChain();
          setNfts(chainResult.data);
          setTotalItems(chainResult.total);
        } catch (chainError) {
          console.error("链上查询也失败:", chainError);
          setNfts([]);
          setTotalItems(0);
        }
      } finally {
        setLoading(false);
      }
    },
    [address, itemsPerPage, queryMyNftFromChain, refreshKey], // 新增：依赖refreshKey
  );

  // 处理每页显示数量变化
  const handlePageSizeChange = useCallback((newSize) => {
    console.log("每页显示数量变化:", newSize);
    setItemsPerPage(newSize);
    // 重置到第一页，因为每页数量变化了
    setCurrentPage(1);
  }, []);

  // 处理分页变化
  const handlePageChange = useCallback((page) => {
    console.log("分页变化，跳转到第", page, "页");
    setCurrentPage(page);
  }, []);

  // 手动刷新按钮（新增：更新refreshKey强制刷新）
  const handleRefresh = useCallback(() => {
    console.log("手动刷新 NFT 数据");
    setRefreshKey((prev) => prev + 1); // 触发强制刷新
    queryMyNft(currentPage);
  }, [queryMyNft, currentPage]);

  // 关键：新增延迟刷新函数（解决上链延迟问题）
  const delayedRefresh = useCallback(
    (delay = 3000) => {
      setTimeout(() => {
        setRefreshKey((prev) => prev + 1);
        queryMyNft(currentPage);
      }, delay);
    },
    [queryMyNft, currentPage],
  );

  // 关键：当地址、当前页、刷新标识变化时触发查询
  useEffect(() => {
    console.log(
      "useEffect触发，地址:",
      address,
      "当前页:",
      currentPage,
      "每页:",
      itemsPerPage,
      "刷新标识:",
      refreshKey,
    );
    if (address) {
      queryMyNft(currentPage);
    } else {
      // 未连接钱包时清空数据
      setNfts([]);
      setTotalItems(0);
    }
  }, [address, currentPage, itemsPerPage, queryMyNft, refreshKey]); // 新增：依赖refreshKey

  // 打开转账模态框
  const openTransferModal = useCallback((tokenId) => {
    setSelectedTokenId(tokenId);
    console.log("打开转账模态框，Token ID:", tokenId);
    // 显示转账模态框
    document.getElementById("transferModal").showModal();
  }, []);

  // 打开上架模态框
  const openListModal = useCallback((tokenId) => {
    setSelectedTokenId(tokenId);
    setNftPriceWei(""); // 新增：清空价格输入
    console.log("打开上架模态框，Token ID:", tokenId);
    // 显示上架模态框
    document.getElementById("listModal").showModal();
  }, []);

  // 打开修改价格模态框
  const openSetPriceModal = useCallback((tokenId, price) => {
    setSelectedTokenId(tokenId);
    // 新增：格式化价格显示
    setSelectedNftPrice(price || "0");
    setNftPriceWei(
      price ? parseFloat(ethers.formatEther(price)).toFixed(3) : "",
    );
    console.log("打开修改价格模态框，Token ID:", tokenId, "当前价格:", price);
    // 显示修改价格模态框
    document.getElementById("setPriceModal").showModal();
  }, []);

  // 打开下架模态框
  const openUnlistModal = useCallback(
    (tokenId, price, name, description, image, listedAt) => {
      setSelectedTokenId(tokenId);
      setSelectedNftName(name);
      setSelectedNftDescription(description);
      setSelectedNftPrice(price || "0");
      setSelectedNftImage(image);
      setSelectedNftListedAt(listedAt);
      console.log("打开下架模态框，Token ID:", tokenId);
      console.log("价格:" + selectedNftPrice);
      // 显示下架模态框
      document.getElementById("unlistModal").showModal();
    },
    [],
  );

  // 转账 NFT
  const transferNFT = useCallback(async () => {
    if (!nft) {
      alert("NFT 合约未就绪，请先连接钱包");
      return;
    }
    if (!ethers.isAddress(targetAddress)) {
      alert("接收地址无效");
      return;
    }
    if (!selectedTokenId) {
      alert("未选择 Token");
      return;
    }

    try {
      const confirmed = window.confirm(
        `确认将 NFT #${selectedTokenId} 转账到 ${targetAddress}？`,
      );
      if (!confirmed) return;

      console.log(`开始转账 NFT #${selectedTokenId} 到 ${targetAddress}`);
      const tx = await nft.transferFrom(
        address,
        targetAddress,
        selectedTokenId,
      );
      await tx.wait();
      alert("转账成功！");

      // 转账成功后延迟刷新（解决上链延迟）
      delayedRefresh();
      // 清空输入
      setTargetAddress("");
      // 关闭模态框
      document.getElementById("transferModal").close();
    } catch (error) {
      console.error("转账失败:", error);
      alert("转账失败: " + error.message);
    }
  }, [nft, address, targetAddress, selectedTokenId, delayedRefresh]);

  // 修改价格（修复模态框关闭错误 + 延迟刷新）
  const setPriceNFT = useCallback(async () => {
    if (!market || !nft) {
      alert("合约未就绪，请先连接钱包");
      return;
    }
    if (!nftPriceWei || nftPriceWei === "0") {
      alert("请输入有效的上架价格");
      return;
    }
    if (!selectedTokenId) {
      alert("未选择 Token");
      return;
    }

    try {
      const priceInWei = ethers.parseEther(nftPriceWei);
      const confirmed = window.confirm(
        `确认修改价格 NFT #${selectedTokenId}，价格: ${nftPriceWei} ETH？`,
      );
      if (!confirmed) return;

      console.log("市场合约: ", market);
      console.log(
        `开始修改价格 NFT #${selectedTokenId}，价格: ${priceInWei} wei`,
      );

      // 在 Market 上修改价格
      const listTx = await market.setPrice(selectedTokenId, priceInWei);
      await listTx.wait();
      alert("修改价格成功！");

      // 修改价格成功后延迟刷新
      delayedRefresh();
      // 清空输入
      setNftPriceWei("");
      // 修复：关闭正确的模态框（原为listModal，错误）
      document.getElementById("setPriceModal").close();
    } catch (error) {
      console.error("修改价格失败:", error);
      alert("修改价格失败: " + error.message);
    }
  }, [market, nft, selectedTokenId, nftPriceWei, delayedRefresh]);

  // 上架 NFT（新增延迟刷新）
  const listNFT = useCallback(async () => {
    if (!market || !nft) {
      alert("合约未就绪，请先连接钱包");
      return;
    }
    if (!nftPriceWei || nftPriceWei === "0") {
      alert("请输入有效的上架价格");
      return;
    }
    if (!selectedTokenId) {
      alert("未选择 Token");
      return;
    }

    try {
      const priceInWei = ethers.parseEther(nftPriceWei);
      const confirmed = window.confirm(
        `确认上架 NFT #${selectedTokenId}，价格: ${nftPriceWei} ETH？`,
      );
      if (!confirmed) return;

      console.log("市场合约: ", market);
      console.log(`开始上架 NFT #${selectedTokenId}，价格: ${priceInWei} wei`);

      // 先授权 Market 合约操作 NFT
      console.log("先授权: ", market.target, selectedTokenId);
      if (!market.target) {
        console.log("合约地址为空, 授权失败");
        return;
      }
      const approveTx = await nft.approve(market.target, selectedTokenId);
      await approveTx.wait();
      console.log("授权成功");

      // 在 Market 上上架 NFT
      const listTx = await market.list(selectedTokenId, priceInWei);
      await listTx.wait();
      alert("上架成功！");

      // 上架成功后延迟刷新（关键：解决上链延迟）
      delayedRefresh();
      // 清空输入
      setNftPriceWei("");
      // 关闭模态框
      document.getElementById("listModal").close();
    } catch (error) {
      console.error("上架失败:", error);
      alert("上架失败: " + error.message);
    }
  }, [market, nft, selectedTokenId, nftPriceWei, delayedRefresh]);

  // 下架 NFT（新增延迟刷新）
  const unlistNFT = useCallback(async () => {
    if (!market || !nft) {
      alert("合约未就绪，请先连接钱包");
      return;
    }
    if (!selectedTokenId) {
      alert("未选择 Token");
      return;
    }

    try {
      const confirmed = window.confirm(`确认下架 NFT #${selectedTokenId}？`);
      if (!confirmed) return;

      console.log("市场合约: ", market);
      console.log(`开始下架 NFT #${selectedTokenId}`);
      // 在 Market 上下架 NFT
      const listTx = await market.unlist(selectedTokenId);
      await listTx.wait();
      alert("下架成功！");

      // 下架成功后延迟刷新
      delayedRefresh();
      // 清空输入
      setNftPriceWei("");
      // 关闭模态框
      document.getElementById("unlistModal").close();
    } catch (error) {
      console.error("下架失败:", error);
      alert("下架失败: " + error.message);
    }
  }, [market, nft, selectedTokenId, nftPriceWei, delayedRefresh]);

  // 格式化地址显示
  const truncateAddress = (address) => {
    if (!address) return "";
    if (address.length > 10) {
      return `${address.slice(0, 6)}...${address.slice(-4)}`;
    }
    return address;
  };

  // 计算当前显示的数据范围
  const startItem =
    totalItems > 0
      ? Math.min((currentPage - 1) * itemsPerPage + 1, totalItems)
      : 0;
  const endItem =
    totalItems > 0 ? Math.min(currentPage * itemsPerPage, totalItems) : 0;

  // 加载状态
  if (loading) {
    return (
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-6">
        <div className="w-full text-white mb-8 flex justify-between items-center">
          <div className="font-bold text-2xl">我的 NFT</div>
          <div className="animate-pulse h-4 w-32 bg-gray-700 rounded"></div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {[...Array(8)].map((_, index) => (
            <div key={index} className="animate-pulse">
              <div className="bg-[#1e1e2e] rounded-2xl overflow-hidden border border-gray-800">
                <div className="h-64 w-full bg-gray-800"></div>
                <div className="p-4">
                  <div className="h-6 bg-gray-700 rounded mb-2"></div>
                  <div className="h-4 bg-gray-700 rounded mb-4"></div>
                  <div className="flex justify-between items-center">
                    <div className="h-6 bg-gray-700 rounded w-24"></div>
                    <div className="flex gap-2">
                      <div className="h-8 bg-gray-700 rounded w-16"></div>
                      <div className="h-8 bg-gray-700 rounded w-16"></div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  // 没有连接钱包的状态
  if (!address) {
    return (
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-6">
        <div className="w-full text-white mb-8">
          <div className="font-bold text-2xl">我的 NFT</div>
        </div>

        <div className="bg-gradient-to-br from-[#1e1e2e] to-[#2d2d44] rounded-3xl p-12 text-center shadow-2xl">
          <div className="inline-block p-4 bg-gray-800/50 rounded-full mb-6">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-16 w-16 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1.5}
                d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
              />
            </svg>
          </div>
          <div className="text-2xl font-bold mb-4">请先连接钱包</div>
          <div className="text-gray-400 text-lg max-w-md mx-auto mb-8">
            连接您的钱包以查看和管理您的 NFT 收藏
          </div>
          <div className="text-sm text-gray-500">
            支持 MetaMask, WalletConnect 等钱包
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-6">
      {/* 顶部标题和信息栏 */}
      <div className="w-full text-white mb-8 ">
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
          <div>
            <div className="font-bold text-2xl">我的 NFT 收藏</div>
            <div className="text-gray-400 text-sm mt-1">
              钱包地址: {truncateAddress(address)}
            </div>
          </div>

          <div className="flex items-center gap-4 flex-wrap">
            <div className="text-sm text-gray-300 bg-gray-900/50 px-4 py-2 rounded-lg">
              {totalItems > 0 ? (
                <>
                  <span className="text-purple-400 font-semibold">
                    {totalItems}
                  </span>{" "}
                  个 NFT
                </>
              ) : (
                "暂无 NFT"
              )}
            </div>

            <div className="flex gap-2">
              <button
                onClick={handleRefresh}
                className="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-white rounded-lg transition-all duration-200 flex items-center gap-2"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
                </svg>
                刷新
              </button>
            </div>
          </div>
        </div>
      </div>
      {/* 分页组件 - 顶部 */}
      {/*
      {totalItems > 0 && (
        <div className="mb-6">
          <div className="bg-gradient-to-r from-[#1a1a2e] to-[#2d2d44] rounded-xl p-4 shadow-lg">
            <Pagination
              totalItems={totalItems}
              itemsPerPage={itemsPerPage}
              currentPage={currentPage}
              setCurrentPage={handlePageChange}
              onPageSizeChange={handlePageSizeChange}
            />
          </div>
        </div>
      )}
      */}
      {/* NFT 卡片网格 */}
      {nfts.length > 0 ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6 o">
          {nfts.map((nftItem, index) => (
            <div
              key={`${nftItem.tokenId}-${index}-${refreshKey}`} // 新增：refreshKey确保重新渲染
              className="group bg-gradient-to-br from-[#1e1e2e] to-[#2a2a3e] rounded-2xl flex flex-col overflow-hidden hover:-translate-y-2  transition-all duration-300 ease-in-out shadow-xl hover:shadow-2xl border border-gray-800 hover:border-purple-500/30"
            >
              {/* NFT 图片 */}
              <div className="relative h-64 overflow-hidden">
                <img
                  className="w-full h-full object-cover  transition-transform duration-300"
                  src={nftItem.image}
                  alt={nftItem.name}
                  onError={(e) => {
                    e.target.src =
                      "https://images.unsplash.com/photo-1620641788421-7a1c342ea42e?ixlib=rb-1.2.1&auto=format&fit=crop&w=500&q=80";
                  }}
                />
                <div className="absolute top-3 right-3 bg-black/80 text-white text-xs px-3 py-1 rounded-full">
                  ID: {nftItem.tokenId}
                </div>
                <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/80 to-transparent p-4">
                  <div className="text-white font-bold text-lg truncate">
                    {nftItem.name}
                  </div>
                </div>
              </div>

              {/* NFT 信息 */}
              <div className="p-4 flex-1 flex flex-col">
                {/*<div className="flex-1">
                  <div className="text-gray-400 text-sm  line-clamp-2">
                    {nftItem.description || "暂无描述"}
                  </div>
                </div>*/}

                <div className="mb-4">
                  <div className="flex items-center mb-3">
                    <div className="text-gray-400 text-sm mr-3 ">当前价格:</div>
                    <div className="text-purple-400 font-bold text-sm flex text-nowrap items-center">
                      {nftItem.price !== "0" ? (
                        <>
                          <span>
                            {
                              // 先判断是否为有效值（非null/undefined/空）
                              nftItem.price && nftItem.price !== ""
                                ? `${parseFloat(
                                    ethers.formatEther(nftItem.price),
                                  ).toFixed(2)} ETH`
                                : `未上架`
                            }
                          </span>
                          <svg
                            xmlns="http://www.w3.org/2000/svg"
                            viewBox="0 0 24 24"
                            fill="rgba(211, 211, 211, 0.7)"
                            className="size-5 ml-3 cursor-pointer"
                            onClick={() =>
                              openSetPriceModal(nftItem.tokenId, nftItem.price)
                            }
                          >
                            <path d="M21.731 2.269a2.625 2.625 0 0 0-3.712 0l-1.157 1.157 3.712 3.712 1.157-1.157a2.625 2.625 0 0 0 0-3.712ZM19.513 8.199l-3.712-3.712-8.4 8.4a5.25 5.25 0 0 0-1.32 2.214l-.8 2.685a.75.75 0 0 0 .933.933l2.685-.8a5.25 5.25 0 0 0 2.214-1.32l8.4-8.4Z" />
                            <path d="M5.25 5.25a3 3 0 0 0-3 3v10.5a3 3 0 0 0 3 3h10.5a3 3 0 0 0 3-3V13.5a.75.75 0 0 0-1.5 0v5.25a1.5 1.5 0 0 1-1.5 1.5H5.25a1.5 1.5 0 0 1-1.5-1.5V8.25a1.5 1.5 0 0 1 1.5-1.5h5.25a.75.75 0 0 0 0-1.5H5.25Z" />
                          </svg>
                        </>
                      ) : (
                        "未上架"
                      )}
                    </div>
                  </div>
                  <div className="flex text-gray-400 text-sm mr-3">
                    <div>{nftItem.listed ? "上架时间:" : "下架时间:"}</div>
                    <div className="ml-2">
                      {nftItem.listed
                        ? nftItem.listedAt
                          ? new Date(nftItem.listedAt).toLocaleString("zh-CN", {
                              hour12: false,
                            })
                          : "-"
                        : nftItem.unlistedAt
                          ? new Date(nftItem.unlistedAt).toLocaleString(
                              "zh-CN",
                              {
                                hour12: false,
                              },
                            )
                          : "-"}
                    </div>
                  </div>
                  <div className="flex text-gray-400 text-sm mr-3 mt-3">
                    <div className="text-gray-400 text-sm mr-3 ">描述说明:</div>
                    <div className="text-gray-400 text-sm  line-clamp-2">
                      {nftItem.description || "暂无描述"}
                    </div>
                  </div>
                </div>

                {/* 价格和按钮区域 */}
                <div className="mt-1">
                  <div className="flex justify-end items-center">
                    {nftItem.listed ? (
                      <button
                        className="px-4 py-2 bg-gradient-to-r from-red-600 to-red-700 hover:from-red-700 hover:to-red-800 text-white text-sm font-medium rounded-lg transition-all duration-200 shadow-lg hover:shadow-red-500/25"
                        onClick={() =>
                          openUnlistModal(
                            nftItem.tokenId,
                            nftItem.price,
                            nftItem.name,
                            nftItem.description,
                            nftItem.image,
                            nftItem.listedAt,
                          )
                        } // 下架时调用 unlistItem 方法
                      >
                        下架
                      </button>
                    ) : (
                      <button
                        className="px-4 py-2 bg-gradient-to-r from-green-600 to-green-700 hover:from-green-700 hover:to-green-800 text-white text-sm font-medium rounded-lg transition-all duration-200 shadow-lg hover:shadow-green-500/25"
                        onClick={() => openListModal(nftItem.tokenId)} // 上架时调用 listItem 方法
                      >
                        上架
                      </button>
                    )}

                    <button
                      className="ml-5 px-4 py-2 bg-gradient-to-r from-blue-700 to-blue-800 hover:from-gray-600 hover:to-gray-700 text-white text-sm font-medium rounded-lg transition-all duration-200 shadow-lg"
                      onClick={() => openTransferModal(nftItem.tokenId)}
                    >
                      转账
                    </button>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : (
        // 没有 NFT 的状态
        <div className="bg-gradient-to-br from-[#1e1e2e] to-[#2d2d44] rounded-3xl p-12 text-center shadow-2xl">
          <div className="inline-block p-6 bg-gray-800/50 rounded-full mb-8">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-20 w-20 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1.5}
                d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
          </div>
          <div className="text-2xl font-bold mb-4">暂无 NFT 收藏</div>
          <div className="text-gray-400 text-lg max-w-md mx-auto mb-8">
            您的钱包中还没有 NFT，快去铸造或购买一些吧！
          </div>
          <div className="flex justify-center gap-4">
            <button
              onClick={handleRefresh}
              className="px-6 py-3 bg-gradient-to-r from-purple-600 to-purple-700 hover:from-purple-700 hover:to-purple-800 text-white font-medium rounded-lg transition-all duration-200 shadow-lg hover:shadow-purple-500/25"
            >
              刷新数据
            </button>
            <button className="px-6 py-3 bg-gradient-to-r from-gray-700 to-gray-800 hover:from-gray-600 hover:to-gray-700 text-white font-medium rounded-lg transition-all duration-200 shadow-lg">
              探索市场
            </button>
          </div>
        </div>
      )}
      {/* 分页组件 - 底部 */}
      {totalItems > 0 && nfts.length > 0 && (
        <div className="mt-8">
          <div className="bg-gradient-to-r from-[#1a1a2e] to-[#2d2d44] rounded-xl p-4 shadow-lg">
            <Pagination
              totalItems={totalItems}
              itemsPerPage={itemsPerPage}
              currentPage={currentPage}
              setCurrentPage={handlePageChange}
              onPageSizeChange={handlePageSizeChange}
            />
          </div>
        </div>
      )}
      {/* 转账模态框 */}
      <dialog id="transferModal" className="modal modal-bottom sm:modal-middle">
        <div className="modal-box bg-gradient-to-br from-[#1e1e2e] to-[#2d2d44] border border-gray-800">
          <h3 className="font-bold text-xl text-white mb-2">转账 NFT</h3>
          <p className="text-gray-400 mb-6">
            将 NFT #{selectedTokenId} 转账到其他地址
          </p>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                目标地址
              </label>
              <input
                type="text"
                value={targetAddress}
                onChange={(e) => setTargetAddress(e.target.value)}
                placeholder="0x..."
                className="w-full px-4 py-3 bg-gray-900/50 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              />
            </div>

            <div className="p-4 bg-gray-900/30 rounded-lg">
              <div className="text-sm text-gray-400 mb-1">转账详情</div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-300">NFT ID:</span>
                <span className="text-purple-400 font-medium">
                  #{selectedTokenId}
                </span>
              </div>
              <div className="flex justify-between text-sm mt-1">
                <span className="text-gray-300">从:</span>
                <span className="text-gray-300 truncate ml-2">
                  {truncateAddress(address)}
                </span>
              </div>
            </div>
          </div>

          <div className="modal-action">
            <form method="dialog">
              <button className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg mr-2 transition-colors duration-200">
                取消
              </button>
            </form>
            <button
              onClick={transferNFT}
              className="px-4 py-2 bg-gradient-to-r from-purple-600 to-purple-700 hover:from-purple-700 hover:to-purple-800 text-white rounded-lg transition-all duration-200"
              disabled={!targetAddress}
            >
              确认转账
            </button>
          </div>
        </div>
        <form method="dialog" className="modal-backdrop">
          <button>关闭</button>
        </form>
      </dialog>

      {/* 修改价格模态框 */}
      <dialog id="setPriceModal" className="modal modal-bottom sm:modal-middle">
        <div className="modal-box bg-gradient-to-br from-[#1e1e2e] to-[#2d2d44] border border-gray-800">
          <h3 className="font-bold text-xl text-white mb-2">
            修改 # {selectedTokenId} 售价
          </h3>

          <p className="text-gray-400 mb-6 mt-4 ">
            当前售价:
            <span className="text-purple-400 font-bold text-lg ml-2">
              {selectedNftPrice && selectedNftPrice !== "0"
                ? `${parseFloat(ethers.formatEther(selectedNftPrice)).toFixed(
                    2,
                  )} ETH`
                : "未上架"}
            </span>{" "}
          </p>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                设置新的售价 (ETH)
              </label>
              <input
                type="number"
                step="0.001"
                min="0"
                value={nftPriceWei}
                onChange={(e) => setNftPriceWei(e.target.value)}
                placeholder="0.05"
                className="w-full px-4 py-3 bg-gray-900/50 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              />
            </div>
          </div>

          <div className="modal-action">
            <form method="dialog">
              <button className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg mr-2 transition-colors duration-200">
                取消
              </button>
            </form>
            <button
              onClick={setPriceNFT}
              className="px-4 py-2 bg-gradient-to-r from-purple-600 to-purple-700 hover:from-purple-700 hover:to-purple-800 text-white rounded-lg transition-all duration-200"
              disabled={!nftPriceWei || parseFloat(nftPriceWei) <= 0}
            >
              确认修改
            </button>
          </div>
        </div>
        <form method="dialog" className="modal-backdrop">
          <button>关闭</button>
        </form>
      </dialog>

      {/* 上架模态框 */}
      <dialog id="listModal" className="modal modal-bottom sm:modal-middle">
        <div className="modal-box bg-gradient-to-br from-[#1e1e2e] to-[#2d2d44] border border-gray-800">
          <h3 className="font-bold text-xl text-white mb-2">上架 NFT</h3>
          <p className="text-gray-400 mb-6">
            设置 NFT #{selectedTokenId} 的出售价格
          </p>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                出售价格 (ETH)
              </label>
              <input
                type="number"
                step="0.001"
                min="0"
                value={nftPriceWei}
                onChange={(e) => setNftPriceWei(e.target.value)}
                placeholder="0.05"
                className="w-full px-4 py-3 bg-gray-900/50 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              />
              <div className="text-xs text-gray-500 mt-2">
                提示: 1 ETH = 10¹⁸ wei
              </div>
            </div>

            <div className="p-4 bg-gray-900/30 rounded-lg">
              <div className="text-sm text-gray-400 mb-1">上架详情</div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-300">NFT ID:</span>
                <span className="text-purple-400 font-medium">
                  #{selectedTokenId}
                </span>
              </div>
              <div className="flex justify-between text-sm mt-1">
                <span className="text-gray-300">当前所有者:</span>
                <span className="text-gray-300 truncate ml-2">
                  {truncateAddress(address)}
                </span>
              </div>
            </div>
          </div>

          <div className="modal-action">
            <form method="dialog">
              <button className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg mr-2 transition-colors duration-200">
                取消
              </button>
            </form>
            <button
              onClick={listNFT}
              className="px-4 py-2 bg-gradient-to-r from-purple-600 to-purple-700 hover:from-purple-700 hover:to-purple-800 text-white rounded-lg transition-all duration-200"
              disabled={!nftPriceWei || parseFloat(nftPriceWei) <= 0}
            >
              确认上架
            </button>
          </div>
        </div>
        <form method="dialog" className="modal-backdrop">
          <button>关闭</button>
        </form>
      </dialog>

      {/* 下架架模态框 */}
      <dialog id="unlistModal" className="modal modal-bottom sm:modal-middle">
        <div className="modal-box bg-gradient-to-br from-[#1e1e2e] to-[#2d2d44] border border-gray-800">
          <h3 className="font-bold text-xl text-white mb-2">下架 NFT</h3>

          <div className="space-y-4">
            <img
              className="w-full h-full object-cover  transition-transform duration-300 rounded-xl "
              src={selectedNftImage}
              onError={(e) => {
                e.target.src =
                  "https://images.unsplash.com/photo-1620641788421-7a1c342ea42e?ixlib=rb-1.2.1&auto=format&fit=crop&w=500&q=80";
              }}
            />

            <div className="p-4 bg-gray-900/30 rounded-lg">
              <div className="flex justify-between text-sm">
                <span className="text-gray-300">售价 </span>
                <span className="text-purple-400  font-bold">
                  {selectedNftPrice && selectedNftPrice !== "0"
                    ? `${parseFloat(
                        ethers.formatEther(selectedNftPrice),
                      ).toFixed(2)} ETH`
                    : "未上架"}
                </span>
              </div>
              <div className="flex justify-between text-sm mt-1">
                <span className="text-gray-300">NFT ID</span>
                <span className="text-gray-300 font-medium">
                  {selectedTokenId}
                </span>
              </div>
              <div className="flex justify-between text-sm mt-1">
                <span className="text-gray-300">名称</span>
                <span className="text-gray-300 font-medium">
                  {selectedNftName}
                </span>
              </div>
              <div className="flex justify-between text-sm mt-1">
                <span className="text-gray-300">说明</span>
                <span className="text-gray-300 font-medium">
                  {selectedNftDescription}
                </span>
              </div>
              <div className="flex justify-between text-sm mt-1">
                <span className="text-gray-300">上架时间</span>
                <span className="text-gray-300 truncate ml-2">
                  {selectedNftListedAt
                    ? new Date(selectedNftListedAt).toLocaleString("zh-CN", {
                        hour12: false,
                      })
                    : "-"}
                </span>
              </div>
            </div>
          </div>

          <div className="modal-action">
            <form method="dialog">
              <button className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg mr-2 transition-colors duration-200">
                取消
              </button>
            </form>
            <button
              onClick={unlistNFT}
              className="px-4 py-2 bg-gradient-to-r from-purple-600 to-purple-700 hover:from-purple-700 hover:to-purple-800 text-white rounded-lg transition-all duration-200"
            >
              确认下架
            </button>
          </div>
        </div>
        <form method="dialog" className="modal-backdrop">
          <button>关闭</button>
        </form>
      </dialog>
    </div>
  );
}

export default MyNft;
