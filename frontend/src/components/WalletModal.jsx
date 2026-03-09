import React, { useState, useEffect } from "react";

function WalletModal() {
  // ==============================
  // 1. 变量定义（状态/普通变量）
  // ==============================
  const [count, setCount] = useState(0); // 可修改的状态变量

  // ==============================
  // 5. JSX 返回
  // ==============================
  return (
    <div className="mt-32 h-36 max-full p-5 bg-[#1d1d3b] shadow-[0_2px_10px_rgba(0,0,0,0.3)] flex flex-col justify-around items-center  ">
      <div role="tablist" className="tabs font-bold text-[#636e72]">
        <a role="tab" className="tab">
          关于我们
        </a>
        <a role="tab" className="tab tab-active">
          服务条款
        </a>
        <a role="tab" className="tab">
          隐私政策
        </a>
        <a role="tab" className="tab">
          帮助中心
        </a>
      </div>
      <div className="text-[#636e72]">© 2023 NFTMarket. 保留所有权利。</div>
    </div>
  );
}

export default WalletModal;
