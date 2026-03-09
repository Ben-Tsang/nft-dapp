import React, { useState, useEffect, useCallback } from "react";
import { useWallet } from "@/context/WalletContext";
import { ethers } from "ethers";
import api from "@/js/api/api";
import Pagination from "./Pagination";
function Discovery() {
  // ==============================
  // 1. 变量定义（状态/普通变量）
  // 保留原有核心状态，补充分页等对齐我的NFT的状态
  // ==============================
  const [loading, setLoading] = useState(false); // 加载状态
  const [count, setCount] = useState(0); // 可修改的状态变量
  const [name, setName] = useState("Alice");
  const [showPage, setShowPage] = useState(true); // 控制页面显示切换
  const [nfts, setNfts] = useState([]); // 存储 NFT 数据
  const [selectedTokenId, setSelectedTokenId] = useState(""); // 记录选中的nft的tokenId
  const [selectedTokenImage, setSelectedTokenImage] = useState("");
  const [selectedTokenName, setSelectedTokenName] = useState("");
  const [selectedTokenDesc, setSelectedTokenDesc] = useState("");
  const [selectedTokenPriceWei, setSelectedTokenPriceWei] = useState(0);

  // 新增：分页相关状态（对齐我的NFT页面）
  const [itemsPerPage, setItemsPerPage] = useState(12);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalItems, setTotalItems] = useState(0);

  // ====== 只改这一块：从 Context 拿钱包与合约 ======
  const { walletData, getContract, getContractRO } = useWallet();
  const address = walletData?.address;

  // 市场合约：只读用于列表，签名用于购买
  const marketRO = getContractRO?.("market");
  const market = getContract?.("market");

  // ==============================
  // 工具函数：地址截断（对齐我的NFT）
  // ==============================
  const truncateAddress = (addr) => {
    if (!addr) return "Unknown";
    if (addr.length > 10) {
      return `${addr.slice(0, 6)}...${addr.slice(-4)}`;
    }
    return addr;
  };

  // ==============================
  // 2. 核心函数：获取NFT列表
  // ==============================
  const queryDiscoveryList = useCallback(
    async (page = 1) => {
      if (!address) {
        console.log("未连接钱包，跳过查询");
        return;
      }

      setLoading(true);
      console.log(
        `开始查询 NFT，地址: ${address}, 页码: ${page}, 每页: ${itemsPerPage}`
      );

      try {
        // 优先从 API 获取数据（添加随机参数，禁用缓存）
        const data = await api.discoveryNftList(page, itemsPerPage, {
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
                tokenUri = tokenUri.replace(
                  "ipfs://",
                  "https://gateway.lighthouse.storage/ipfs/"
                );
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
                img = img.replace(
                  "ipfs://",
                  "https://gateway.lighthouse.storage/ipfs/"
                );
              }

              // 处理 nft_uri 的图片链接
              let image = item.nft_uri || "";
              if (image && image.startsWith("ipfs://")) {
                image = image.replace(
                  "ipfs://",
                  "https://gateway.lighthouse.storage/ipfs/"
                );
              }

              return {
                tokenId: item.nft_id || item.tokenId || "",
                name: meta.name || item.nft_name || `NFT #${item.nft_id || ""}`,
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
            })
          );

          // 获取总数
          totalFromApi = data.total || data.count || data.records.length;
          console.log(
            "API 返回总数:",
            totalFromApi,
            "格式化 NFTs:",
            formattedNfts.length
          );
        }
        setNfts(formattedNfts);
        setTotalItems(totalFromApi);
      } catch (apiError) {
        console.error("API查询失败: ", apiError);
      } finally {
        setLoading(false);
      }
    },
    [address, itemsPerPage] //
  );

  const queryMyNft = useCallback(
    async (page = 1) => {
      if (!marketRO) return;

      setLoading(true);
      try {
        // 保留原有合约调用逻辑
        const [tokenIds, tokenUris, tokenPrice] = await marketRO.listTokens(
          page,
          itemsPerPage
        );
        setTotalItems(tokenIds.length); // 设置总数

        const nftPromises = tokenIds.map(async (tokenId, index) => {
          try {
            let tokenUri = tokenUris[index];
            let tokenPriceWei = tokenPrice[index];

            if (
              typeof tokenUri === "string" &&
              tokenUri.startsWith("ipfs://")
            ) {
              tokenUri = tokenUri.replace(
                "ipfs://",
                "https://gateway.lighthouse.storage/ipfs/"
              );
            }

            const response = await fetch(tokenUri);
            const metadata = await response.json();

            let imageUrl = metadata.image;
            if (
              typeof imageUrl === "string" &&
              imageUrl.startsWith("ipfs://")
            ) {
              imageUrl = imageUrl.replace(
                "ipfs://",
                "https://gateway.lighthouse.storage/ipfs/"
              );
            }

            return {
              tokenId: tokenId.toString(),
              name: metadata.name || `NFT #${tokenId}`,
              description: metadata.description || "",
              image:
                imageUrl ||
                "https://images.unsplash.com/photo-1620641788421-7a1c342ea42e?ixlib=rb-1.2.1&auto=format&fit=crop&w=500&q=80",
              attributes: metadata.attributes || [],
              creator: metadata.creator || "Unknown",
              price: tokenPriceWei,
            };
          } catch (error) {
            console.error(
              `Error fetching metadata for token ${tokenId}:`,
              error
            );
            return {
              tokenId: tokenId.toString(),
              name: `NFT #${tokenId}`,
              description: "Metadata not available",
              image:
                "https://images.unsplash.com/photo-1620641788421-7a1c342ea42e?ixlib=rb-1.2.1&auto=format&fit=crop&w=500&q=80",
              attributes: [],
              creator: "Unknown",
              price: 0,
            };
          }
        });

        const nftData = await Promise.all(nftPromises);
        setNfts(nftData);
      } catch (error) {
        console.error("Error fetching NFTs:", error);
        setNfts([]);
        setTotalItems(0);
      } finally {
        setLoading(false);
      }
    },
    [marketRO, itemsPerPage]
  );

  // ==============================
  // 3. 分页处理函数（对齐我的NFT）
  // ==============================
  const handlePageChange = useCallback((page) => {
    setCurrentPage(page);
  }, []);

  const handlePageSizeChange = useCallback((newSize) => {
    setItemsPerPage(newSize);
    setCurrentPage(1);
  }, []);

  // ==============================
  // 4. 购买相关函数（保留原有逻辑）
  // ==============================
  // 打开购买模态框
  const openBuyModal = (tokenId, name, image, price) => {
    setSelectedTokenId(tokenId);
    setSelectedTokenName(name);
    setSelectedTokenImage(image);
    setSelectedTokenPriceWei(price);
    document.getElementById("buyModal").showModal();
  };

  // 小助手：从市场合约里挑一个可用的“购买方法”
  const pickBuyFunction = (contract) => {
    if (!contract?.interface) return null;
    const funcs =
      contract.interface.fragments?.filter((f) => f.type === "function") || [];
    const candidates = [
      { name: "buy", argc: 1 }, // buy(listingId|tokenId)
      { name: "buyNow", argc: 1 }, // buyNow(listingId)
      { name: "purchase", argc: 1 }, // purchase(listingId)
      { name: "executeTrade", argc: 1 }, // executeTrade(listingId)
      { name: "buy", argc: 2 }, // buy(nft, tokenId)
    ];
    for (const c of candidates) {
      const f = funcs.find(
        (ff) => ff.name === c.name && (ff.inputs?.length || 0) === c.argc
      );
      if (f) return f;
    }
    return null;
  };

  // 购买 NFT 函数（保留原有逻辑）
  const buyNFT = async () => {
    if (!market) {
      alert("请先连接钱包（市场合约未就绪）");
      return;
    }
    if (
      !window.confirm(
        `确认以 ${ethers.formatEther(
          selectedTokenPriceWei || 0n
        )} ETH 购买 NFT #${selectedTokenId}？`
      )
    ) {
      return;
    }

    const buyFrag = pickBuyFunction(market);
    if (!buyFrag) {
      const all =
        market?.interface?.fragments
          ?.filter((f) => f.type === "function")
          ?.map(
            (f) => `${f.name}(${(f.inputs || []).map((i) => i.type).join(",")})`
          ) || [];
      alert("市场合约中未找到可用的购买方法。可用函数有：\n" + all.join("\n"));
      return;
    }

    try {
      // 兼容不同签名：优先 1 参数（listingId/tokenId），其次 2 参数（nft, tokenId）
      let tx;
      if (buyFrag.inputs.length === 1) {
        tx = await market[buyFrag.name](selectedTokenId, {
          value: selectedTokenPriceWei,
        });
      } else {
        // 需要 nft 地址 + tokenId 的签名，这里从合约读取或让后续你补齐
        alert(
          "你的市场合约 buy(...) 需要两个参数（nft, tokenId），请补充 NFT 合约地址后再调用。"
        );
        return;
      }
      await tx.wait();
      alert("购买完成！");
      // 关闭模态框
      document.getElementById("buyModal").close();
      // 重置选中状态
      setSelectedTokenId("");
      setSelectedTokenName("");
      setSelectedTokenImage("");
      setSelectedTokenPriceWei(0);
      // 购买成功后刷新数据
      queryDiscoveryList(currentPage);
    } catch (error) {
      console.error("购买NFT失败:", error);
      alert("购买失败: " + error.message);
    }
  };

  // ==============================
  // 5. 生命周期：页面加载触发查询
  // ==============================
  useEffect(() => {
    if (address && showPage) {
      queryDiscoveryList(currentPage);
    }
  }, [address, currentPage, itemsPerPage, showPage, queryDiscoveryList]);

  // ==============================
  // 6. 加载状态组件（对齐我的NFT）
  // ==============================
  const LoadingSkeleton = () => (
    <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-6">
      <div className="w-full text-white mb-8 flex justify-between items-center">
        <div className="font-bold text-2xl">发现 NFT</div>
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
                  <div className="h-8 bg-gray-700 rounded w-16"></div>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );

  // ==============================
  // 7. 未连接钱包状态（对齐我的NFT）
  // ==============================
  if (!address) {
    return (
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-6">
        <div className="w-full text-white mb-8">
          <div className="font-bold text-2xl">发现 NFT</div>
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
            连接您的钱包以浏览和购买精彩的 NFT 藏品
          </div>
          <div className="text-sm text-gray-500">
            支持 MetaMask, WalletConnect 等钱包
          </div>
        </div>
      </div>
    );
  }

  // ==============================
  // 8. 无NFT状态（对齐我的NFT）
  // ==============================
  if (nfts.length === 0 && !loading) {
    return (
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-6">
        <div className="w-full text-white mb-8">
          <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
            <div>
              <div className="font-bold text-2xl">发现 NFT 市场</div>
              <div className="text-gray-400 text-sm mt-1">
                探索全网优质的 NFT 藏品
              </div>
            </div>
          </div>
        </div>

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
          <div className="text-2xl font-bold mb-4">暂无可发现的 NFT</div>
          <div className="text-gray-400 text-lg max-w-md mx-auto mb-8">
            市场中暂无上架的 NFT，敬请期待！
          </div>
          <button
            onClick={() => queryDiscoveryList(currentPage)}
            className="px-6 py-3 bg-gradient-to-r from-purple-600 to-purple-700 hover:from-purple-700 hover:to-purple-800 text-white font-medium rounded-lg transition-all duration-200 shadow-lg hover:shadow-purple-500/25"
          >
            刷新数据
          </button>
        </div>
      </div>
    );
  }

  // ==============================
  // 9. 加载中状态
  // ==============================
  if (loading) {
    return <LoadingSkeleton />;
  }

  // ==============================
  // 10. 主页面渲染（核心：对齐我的NFT风格，修复白线问题）
  // ==============================
  return (
    <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-6">
      {/* 顶部标题栏（对齐我的NFT） */}
      <div className="w-full text-white mb-8">
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
          <div>
            <div className="font-bold text-2xl">发现 NFT 市场</div>
            <div className="text-gray-400 text-sm mt-1">
              探索全网优质的 NFT 藏品
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

            <button
              onClick={() => queryDiscoveryList(currentPage)}
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

      {/* NFT卡片网格 */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        {nfts.map((nft, index) => (
          <div
            key={`${nft.tokenId}-${index}`}
            className="group bg-gradient-to-br from-[#1e1e2e] to-[#2a2a3e] rounded-2xl flex flex-col overflow-hidden hover:-translate-y-2 cursor-pointer transition-all duration-300 ease-in-out shadow-xl hover:shadow-2xl border border-gray-800 hover:border-purple-500/30"
          >
            {/* NFT图片区域（修复白线：增加transform-gpu origin-center，遮罩bottom:-1px） */}
            <div className="relative h-64 overflow-hidden">
              <img
                className="w-full h-full object-cover  transition-transform duration-300 transform-gpu origin-center"
                src={nft.image}
                alt={nft.name}
                onError={(e) => {
                  e.target.src =
                    "https://images.unsplash.com/photo-1620641788421-7a1c342ea42e?ixlib=rb-1.2.1&auto=format&fit=crop&w=500&q=80";
                }}
              />
              <div className="absolute top-3 right-3 bg-black/80 text-white text-xs px-3 py-1 rounded-full">
                ID: {nft.tokenId}
              </div>
              <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/80 to-transparent p-4">
                <div className="text-white font-bold text-lg truncate">
                  {nft.name}
                </div>
              </div>
            </div>

            {/* NFT信息区域（对齐我的NFT） */}
            <div className="p-5 flex-1 flex flex-col">
              <div className="flex-1">
                <div className="space-y-2 mb-4">
                  <div className="flex items-center text-sm">
                    <span className="text-gray-500 mr-2">创作者:</span>
                    <span
                      className="text-gray-300 truncate"
                      title={nft.creator}
                    >
                      {truncateAddress(nft.creator)}
                    </span>
                  </div>
                  <div className="flex items-center text-sm">
                    <span className="text-gray-500 mr-2">上架时间:</span>
                    <span className="text-gray-300">
                      {new Date(nft.listedAt).toLocaleString("zh-CN", {
                        hour12: false,
                      })}
                    </span>
                  </div>
                  <div className="flex items-center text-sm">
                    <span className="text-gray-500 mr-2">描述说明:</span>
                    <span className="text-gray-300">{nft.description}</span>
                  </div>
                </div>
              </div>

              {/* 价格和购买按钮（替换我的NFT的转账/上架按钮） */}
              <div className="pt-4 border-t border-gray-800">
                <div className="flex justify-between items-center">
                  <div>
                    <div className="text-gray-400 text-xs">售价</div>
                    <div className="text-purple-400 font-bold text-lg">
                      {nft.price &&
                      parseFloat(ethers.formatEther(nft.price)) > 0
                        ? `${parseFloat(ethers.formatEther(nft.price)).toFixed(
                            2
                          )} ETH`
                        : "未定价"}
                    </div>
                  </div>
                  <button
                    className="px-4 py-2 bg-gradient-to-r from-purple-600 to-purple-700 hover:from-purple-700 hover:to-purple-800 text-white text-sm font-medium rounded-lg transition-all duration-200 shadow-lg hover:shadow-purple-500/25"
                    onClick={() =>
                      openBuyModal(nft.tokenId, nft.name, nft.image, nft.price)
                    }
                  >
                    购买
                  </button>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* 分页组件（如果需要请补充Pagination组件，这里预留位置） */}
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

      {/* 购买模态框（对齐我的NFT风格） */}
      <dialog id="buyModal" className="modal modal-bottom sm:modal-middle">
        <div className="modal-box bg-gradient-to-br from-[#1e1e2e] to-[#2d2d44] border border-gray-800">
          <h3 className="font-bold text-xl text-white mb-2">购买 NFT</h3>
          <p className="text-gray-400 mb-6">确认购买 NFT #{selectedTokenId}</p>

          {/* NFT预览（同样修复白线问题） */}
          <div className="relative h-64 overflow-hidden rounded-xl mb-6">
            <img
              className="w-full h-full object-cover transform-gpu origin-center"
              src={selectedTokenImage}
              alt={selectedTokenName}
              onError={(e) => {
                e.target.src =
                  "https://images.unsplash.com/photo-1620641788421-7a1c342ea42e?ixlib=rb-1.2.1&auto=format&fit=crop&w=500&q=80";
              }}
            />
          </div>

          {/* NFT信息 */}
          <div className="space-y-4 mb-6">
            <div className="flex justify-between text-sm">
              <span className="text-gray-300">NFT 名称:</span>
              <span className="text-white font-medium truncate ml-2">
                {selectedTokenName}
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-300">Token ID:</span>
              <span className="text-purple-400 font-medium">
                #{selectedTokenId}
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-300">购买价格:</span>
              <span className="text-purple-400 font-bold text-lg">
                {selectedTokenPriceWei
                  ? `${parseFloat(
                      ethers.formatEther(selectedTokenPriceWei)
                    ).toFixed(2)} ETH`
                  : "0 ETH"}
              </span>
            </div>
          </div>

          {/* 操作按钮（对齐我的NFT） */}
          <div className="modal-action">
            <form method="dialog">
              <button className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg mr-2 transition-colors duration-200">
                取消
              </button>
            </form>
            <button
              onClick={() => buyNFT()}
              className="px-4 py-2 bg-gradient-to-r from-purple-600 to-purple-700 hover:from-purple-700 hover:to-purple-800 text-white rounded-lg transition-all duration-200"
              disabled={
                !selectedTokenPriceWei ||
                parseFloat(ethers.formatEther(selectedTokenPriceWei)) <= 0
              }
            >
              确认购买
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

export default Discovery;
