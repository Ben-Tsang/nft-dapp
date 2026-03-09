import React, { useState, useEffect, useRef } from "react";
import { useWallet } from "@/context/WalletContext";
import api from "@/js/api/api";

function Mint() {
  // --------------- State ---------------
  const [selectedFile, setSelectedFile] = useState(null);
  const [previewUrl, setPreviewUrl] = useState("");
  const [nftName, setNftName] = useState("");
  const [nftDescription, setNftDescription] = useState("");
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0); // 0~100
  const [nftTotal, setNftTotal] = useState(0);
  const [nftContract, setNftContract] = useState(null); // 统一为nftContract（带Signer）
  const [nftContractRO, setNftContractRO] = useState(null); // 只读NFT合约
  const [contractReady, setContractReady] = useState(false);
  const [minting, setMinting] = useState(false); // 铸造中状态
  const fileInputRef = useRef(null);

  // ✅ 从 Context 中获取：钱包信息 + 多合约入口
  const { walletData, getContract, getContractRO } = useWallet();
  const address = walletData?.address;
  const networkName = walletData?.networkName;

  // ========== 初始化合约实例 ==========
  useEffect(() => {
    // 重置合约状态
    setNftContract(null);
    setNftContractRO(null);
    setContractReady(false);

    // 钱包未连接/无合约方法时直接返回
    if (!address || !getContract || !getContractRO) {
      console.warn("❌ 钱包未连接或合约方法未初始化");
      return;
    }

    console.log("🔗 钱包已连接，初始化NFT合约...");
    try {
      // 1. 获取带签名权限的NFT合约（用于铸造：写操作）
      const nftSignerContract = getContract("nft");
      // 2. 获取只读NFT合约（用于查询：读操作）
      const nftProviderContract = getContractRO("nft");

      // 校验合约实例是否有效
      if (!nftSignerContract || !nftProviderContract) {
        throw new Error("合约实例获取失败");
      }

      setNftContract(nftSignerContract);
      setNftContractRO(nftProviderContract);
      setContractReady(true);
      console.log("✅ NFT合约初始化完成");
    } catch (err) {
      console.error("❌ 合约初始化失败:", err);
      alert(`合约初始化失败：${err.message}`);
      setContractReady(false);
    }
  }, [address, getContract, getContractRO, networkName]);

  // ========== 查询用户NFT数量（只读操作） ==========
  const getTokensByOwner = async (ownerAddress) => {
    if (!nftContractRO || !ownerAddress) return [];
    try {
      const balance = await nftContractRO.balanceOf(ownerAddress);
      const balanceNum = Number(balance.toString());
      if (balanceNum === 0) return [];

      const tokenIds = [];
      const maxCheck = 100; // 限制最大查询数量，避免卡死
      for (let tokenId = 1; tokenId <= maxCheck; tokenId++) {
        try {
          const owner = await nftContractRO.ownerOf(tokenId);
          if (owner.toLowerCase() === ownerAddress.toLowerCase()) {
            tokenIds.push(tokenId.toString());
            if (tokenIds.length >= balanceNum) break;
          }
        } catch (error) {
          // 该tokenId不存在，跳过
          continue;
        }
      }
      return tokenIds;
    } catch (error) {
      console.error("查询NFT失败:", error);
      alert(`查询NFT数量失败：${error.message}`);
      return [];
    }
  };

  // ========== 加载用户NFT数量 ==========
  useEffect(() => {
    const loadNFTs = async () => {
      if (!address || !nftContractRO) return;
      const myTokenIds = await getTokensByOwner(address);
      setNftTotal(myTokenIds.length);
    };
    loadNFTs();
  }, [address, nftContractRO]);

  // ========== Filebase配置（从环境变量读取更安全） ==========
  const FILEBASE_ACCESS_KEY = import.meta.env?.VITE_FILEBASE_ACCESS_KEY;
  const FILEBASE_BUCKET = import.meta.env?.VITE_FILEBASE_BUCKET;

  // ========== 文件处理逻辑 ==========
  const handleFileSelect = (event) => {
    const file = event.target.files?.[0];
    if (!file) return;

    // 校验文件类型和大小
    const validTypes = ["image/jpeg", "image/png", "image/gif", "video/mp4"];
    const maxSize = 30 * 1024 * 1024; // 30MB
    if (!validTypes.includes(file.type)) {
      alert("请选择 JPG/PNG/GIF/MP4 文件");
      return;
    }
    if (file.size > maxSize) {
      alert("文件不能超过 30MB");
      return;
    }

    setSelectedFile(file);
    setPreviewUrl(URL.createObjectURL(file));
  };

  const handleDropAreaClick = () => fileInputRef.current?.click();
  const handleDragOver = (e) => e.preventDefault();
  const handleDrop = (e) => {
    e.preventDefault();
    const file = e.dataTransfer.files?.[0];
    if (!file) return;
    handleFileSelect({ target: { files: [file] } });
  };

  // ========== 上传进度回调 ==========
  const onProgress = (percent) => {
    setUploadProgress(Math.max(0, Math.min(100, percent)));
  };

  // ========== Filebase上传核心逻辑 ==========
  async function uploadFileToFilebase(file) {
    if (!FILEBASE_ACCESS_KEY || !FILEBASE_BUCKET) {
      throw new Error(
        "Filebase 配置不完整（请检查VITE_FILEBASE_ACCESS_KEY和VITE_FILEBASE_BUCKET）",
      );
    }

    try {
      onProgress(0);
      const formData = new FormData();
      formData.append("file", file);

      const params = new URLSearchParams({
        bucket: FILEBASE_BUCKET,
        "wrap-with-directory": "true",
        "cid-version": "1",
        pin: "true",
      });

      return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();
        xhr.open(
          "POST",
          `https://rpc.filebase.io/api/v0/add?${params.toString()}`,
          true,
        );
        xhr.setRequestHeader("Authorization", `Bearer ${FILEBASE_ACCESS_KEY}`);

        // 上传进度监听
        xhr.upload.onprogress = (e) => {
          if (e.lengthComputable) {
            const percent = Math.round((e.loaded / e.total) * 100);
            onProgress(percent);
          }
        };

        // 上传完成处理
        xhr.onload = () => {
          if (xhr.status >= 200 && xhr.status < 300) {
            const responseText = xhr.responseText;
            const results = responseText
              .split("\n")
              .filter((line) => line.trim())
              .map((line) => JSON.parse(line));
            const mainResult = results[0];
            const cid = mainResult?.Hash;

            if (!cid) {
              reject(new Error("Filebase 返回空 CID"));
              return;
            }

            onProgress(100);
            resolve({
              cid,
              uri: `ipfs://${cid}`,
              url: `https://${cid}.ipfs.filebase.io/ipfs/${cid}`,
            });
          } else {
            reject(
              new Error(`Filebase 上传失败: ${xhr.status} - ${xhr.statusText}`),
            );
          }
        };

        // 网络错误处理
        xhr.onerror = () => {
          reject(new Error("网络错误：无法连接到 Filebase，请检查网络或配置"));
        };

        xhr.send(formData);
      });
    } catch (error) {
      console.error("Filebase 上传错误:", error);
      throw error;
    }
  }

  // ========== 文件哈希计算（防重复铸造） ==========
  async function calculateFileHash(file) {
    const cryptoObj = window.crypto || window.msCrypto;
    if (!cryptoObj || !cryptoObj.subtle) {
      console.warn("当前环境不支持 crypto.subtle，使用简易哈希");
      return `${file.name}-${file.size}-${Date.now()}`;
    }

    const fileReader = new FileReader();
    return new Promise((resolve, reject) => {
      fileReader.onloadend = async () => {
        try {
          const hashBuffer = await cryptoObj.subtle.digest(
            "SHA-256",
            fileReader.result,
          );
          const hashArray = Array.from(new Uint8Array(hashBuffer));
          const hashHex = hashArray
            .map((byte) => byte.toString(16).padStart(2, "0"))
            .join("");
          resolve(hashHex);
        } catch (err) {
          resolve(`${file.name}-${file.size}-${Date.now()}`);
        }
      };
      fileReader.onerror = () =>
        resolve(`${file.name}-${file.size}-${Date.now()}`);
      fileReader.readAsArrayBuffer(file);
    });
  }

  // ========== IPFS上传（文件+元数据） ==========
  async function uploadToIPFS(imageFile, name, description) {
    setIsUploading(true);
    setUploadProgress(0);
    try {
      // 1. 上传NFT文件（图片/视频）
      const imageUploadResult = await uploadFileToFilebase(imageFile);
      if (!imageUploadResult.uri) {
        throw new Error("NFT文件上传IPFS失败");
      }

      // 2. 构建NFT元数据
      const metadata = {
        name: name || "Untitled NFT",
        description: description || "No description",
        image: imageUploadResult.uri,
        timestamp: Date.now(),
        creator: address,
      };

      // 3. 上传元数据文件
      const metadataBlob = new Blob([JSON.stringify(metadata, null, 2)], {
        type: "application/json",
      });
      const metadataFile = new File([metadataBlob], "metadata.json", {
        type: "application/json",
      });
      const metadataUploadResult = await uploadFileToFilebase(metadataFile);

      return metadataUploadResult.uri;
    } finally {
      setIsUploading(false);
      setUploadProgress(0);
    }
  }

  // ========== 核心：NFT铸造逻辑 ==========
  const handleMintClick = async () => {
    // 1. 前置校验
    if (!address) {
      alert("请先连接钱包！");
      return;
    }
    if (!contractReady || !nftContract) {
      alert("NFT合约未就绪，请检查钱包连接或链选择是否正确！");
      return;
    }
    if (!selectedFile) {
      alert("请先选择要铸造的NFT文件（图片/视频）！");
      return;
    }
    if (!nftName.trim()) {
      alert("请输入NFT名称！");
      return;
    }
    if (minting) {
      alert("正在铸造中，请稍候...");
      return;
    }

    try {
      setMinting(true);

      // 2. 计算文件哈希，检查重复铸造
      const fileHash = await calculateFileHash(selectedFile);
      let isDuplicate = false;
      if (fileHash) {
        try {
          const checkResult = await api.CheckFileDuplicate(fileHash);
          isDuplicate = checkResult?.isDuplicate || false;
        } catch (err) {
          console.warn("重复检查失败，跳过检查:", err);
        }
      }
      if (isDuplicate) {
        alert("该文件已铸造过NFT，不允许重复铸造！");
        setMinting(false);
        return;
      }

      // 3. 上传到IPFS，获取元数据URI
      const tokenURI = await uploadToIPFS(
        selectedFile,
        nftName,
        nftDescription,
      );
      if (!tokenURI) {
        throw new Error("IPFS上传失败，无法获取元数据URI");
      }
      console.log("✅ IPFS上传完成，tokenURI:", tokenURI);

      // 4. 执行铸造（核心步骤）
      console.log("🔨 开始铸造NFT，参数：", {
        to: address,
        name: nftName,
        description: nftDescription,
        tokenURI: tokenURI,
      });

      // ⚠️ 关键：适配合约ABI的safeMint参数（以下提供2种常见写法，根据你的合约ABI选择）
      let tx;

      // 写法2：如果合约safeMint参数是 (name, description, tokenURI)（你的原始写法）
      tx = await nftContract.safeMint(nftName, nftDescription, tokenURI);

      console.log("📤 铸造交易已发送，等待链上确认...");
      const receipt = await tx.wait(); // 等待交易确认
      console.log("✅ NFT铸造成功，交易回执:", receipt);

      // 5. 提取铸造的TokenID
      const tokenIdEvent = receipt.events?.find(
        (e) => e.event === "Transfer" || e.event === "NFTMinted",
      );
      const tokenId = tokenIdEvent?.args?.tokenId?.toString() || "未知";

      // 6. 提示用户并刷新NFT数量
      alert(`🎉 NFT铸造成功！\nToken ID: ${tokenId}\nIPFS地址: ${tokenURI}`);

      // 7. 刷新用户NFT数量
      const myTokenIds = await getTokensByOwner(address);
      setNftTotal(myTokenIds.length);

      // 8. 清空表单
      setSelectedFile(null);
      setPreviewUrl("");
      setNftName("");
      setNftDescription("");
    } catch (err) {
      console.error("❌ 铸造失败:", err);
      alert(`铸造失败：${err.message || "未知错误，请检查控制台日志"}`);
    } finally {
      setMinting(false);
    }
  };

  // ========== 清理预览URL，避免内存泄漏 ==========
  useEffect(() => {
    return () => {
      if (previewUrl) URL.revokeObjectURL(previewUrl);
    };
  }, [previewUrl]);

  // ========== UI渲染 ==========
  return (
    <div className="max-w-7xl mx-auto flex flex-col p-4">
      {/* 标题区域 */}
      <div className="text-white w-full mx-auto text-4xl font-bold flex justify-center items-center mt-10">
        铸造你的 NFT
      </div>
      <div className="flex justify-center mt-5 text-[#bfe709]">
        你已拥有 {nftTotal} 个NFT
      </div>

      {/* 提示信息 */}
      {!address && (
        <div className="mt-5 text-center text-yellow-400">
          ⚠️ 请先连接钱包后再铸造NFT
        </div>
      )}
      {address && !contractReady && (
        <div className="mt-5 text-center text-red-400">
          ⚠️ NFT合约未就绪，请检查链选择或重新连接钱包
        </div>
      )}

      {/* 核心表单区域 */}
      <div className="mt-10 w-full p-10 bg-[#1d1d3b] rounded-2xl flex flex-col md:flex-row md:justify-between gap-8">
        {/* 左侧文件上传 */}
        <div className="w-full md:w-[48%] flex flex-col">
          <legend className="text-white mb-2 text-base font-thin w-full">
            上传 NFT 文件
          </legend>
          <input
            type="file"
            ref={fileInputRef}
            onChange={handleFileSelect}
            accept=".jpg,.jpeg,.png,.gif,.mp4"
            className="hidden"
          />
          <div
            className={`cursor-pointer h-64 bg-transparent border border-white border-dashed flex flex-col justify-center items-center text-white transition-colors ${
              selectedFile ? "border-green-500" : "hover:border-blue-500"
            }`}
            onClick={handleDropAreaClick}
            onDragOver={handleDragOver}
            onDrop={handleDrop}
          >
            {previewUrl ? (
              <div className="flex flex-col items-center">
                {selectedFile.type.startsWith("video/") ? (
                  <video
                    className="max-h-40 max-w-full mb-2"
                    src={previewUrl}
                    controls
                    muted
                  />
                ) : (
                  <img
                    src={previewUrl}
                    alt="NFT预览"
                    className="max-h-40 max-w-full object-contain mb-2"
                  />
                )}
                <span className="text-sm text-green-400">点击更换文件</span>
              </div>
            ) : (
              <div className="text-center p-4">
                <div className="text-lg mb-2">📁</div>
                <div>点击或拖拽文件到此区域上传</div>
                <div className="text-sm mt-2 text-gray-300">
                  支持 JPG / PNG / GIF / MP4，最大 30MB
                </div>
              </div>
            )}
          </div>

          {/* 已选文件信息 */}
          {selectedFile && (
            <div className="mt-4 text-white text-sm">
              <div>文件名：{selectedFile.name}</div>
              <div>
                文件大小：{(selectedFile.size / 1024 / 1024).toFixed(2)} MB
              </div>
              <div>文件类型：{selectedFile.type}</div>
            </div>
          )}

          {/* 上传进度条 */}
          {isUploading && (
            <div className="mt-4">
              <div className="w-full bg-gray-700 rounded h-2">
                <div
                  className="h-2 bg-green-500 rounded"
                  style={{ width: `${uploadProgress}%` }}
                />
              </div>
              <div className="text-white text-sm mt-1 text-right">
                上传进度：{uploadProgress}%
              </div>
            </div>
          )}
        </div>

        {/* 右侧表单 */}
        <div className="w-full md:w-[48%] flex flex-col">
          {/* NFT名称 */}
          <div className="w-full">
            <legend className="text-white mb-2 text-base font-thin w-full">
              NFT 名称
            </legend>
            <input
              type="text"
              className="input w-full bg-transparent border p-2 caret-white text-white"
              placeholder="为你的 NFT 起一个独特的名字"
              value={nftName}
              onChange={(e) => setNftName(e.target.value)}
              disabled={minting}
            />
          </div>

          {/* NFT描述 */}
          <div className="mt-5">
            <legend className="text-white mb-2 text-base font-thin">
              NFT 描述
            </legend>
            <textarea
              className="input w-full h-40 bg-transparent border p-2 caret-white text-white"
              placeholder="描述你的 NFT 作品（选填）"
              value={nftDescription}
              onChange={(e) => setNftDescription(e.target.value)}
              disabled={minting}
            />
          </div>

          {/* 铸造按钮 */}
          <button
            disabled={
              minting ||
              isUploading ||
              !contractReady ||
              !address ||
              !selectedFile ||
              !nftName.trim()
            }
            onClick={handleMintClick}
            className={`w-full p-5 rounded-md mt-8 text-center text-white font-bold transition-colors ${
              minting ||
              isUploading ||
              !contractReady ||
              !address ||
              !selectedFile ||
              !nftName.trim()
                ? "bg-gray-500 cursor-not-allowed"
                : "bg-[#00b894] hover:bg-[#00a085]"
            }`}
          >
            {minting ? "铸造中..." : isUploading ? "上传中..." : "开始铸造"}
          </button>

          {/* 状态提示 */}
          {(minting || isUploading) && (
            <div className="mt-4 text-white text-center">
              {minting
                ? "正在铸造NFT到链上，请确认钱包交易..."
                : "正在上传到 IPFS，请稍候…"}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default Mint;
