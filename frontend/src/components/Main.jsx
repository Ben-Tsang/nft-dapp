import React, { useState, useEffect } from "react";

function Main() {
  // ==============================
  // 1. 变量定义（状态/普通变量）
  // ==============================
  const [count, setCount] = useState(0); // 可修改的状态变量
  const [name, setName] = useState("Alice");
  const [showPage, setShowPage] = useState(true); // 控制页面显示切换

  let tempVar = "我是普通变量"; // 不会触发渲染

  // ==============================
  // 2. 初始化方法 & 事件绑定
  // ==============================
  useEffect(() => {
    console.log("组件初始化，只执行一次");

    // 绑定窗口大小变化事件
    const handleResize = () => {
      console.log("窗口大小变化:", window.innerWidth);
    };
    window.addEventListener("resize", handleResize);

    // 模拟初始化数据
    setCount(5);
    setName("Bob");

    // 返回清理函数，组件卸载时触发
    return () => {
      console.log("组件卸载，解绑事件");
      window.removeEventListener("resize", handleResize);
    };
  }, []); // 空依赖 → 只执行一次

  // ==============================
  // 3. 每次状态变化触发
  // ==============================
  useEffect(() => {
    console.log("count 变化了:", count);
  }, [count]);

  // ==============================
  // 4. 事件绑定（按钮点击）
  // ==============================
  const handleIncrement = () => {
    setCount(count + 1);
  };

  const handleTogglePage = () => {
    setShowPage(!showPage);
  };

  // ==============================
  // 5. JSX 返回
  // ==============================
  return (
    <div className="w-full h-50 overflow-auto flex-col">
      {/* 搜索/铸造 */}
      <div className="w-full bg-[linear-gradient(to_right,#1d1d3b,#12121a)] pt-24">
        <div className="mx-auto max-w-7xl h-30  flex flex-col justify-center ">
          <div className="text-center text-5xl  font-bold bg-gradient-to-r from-[#6c5ce7] to-[#a29bfe] bg-clip-text text-transparent">探索独一无二的数字藏品</div>
          <div className="text-center text-[#636e72] text-xl mt-10">在NFTMarket发现、收集和出售非凡的数字艺术品。加入全球最大的NFT社区。</div>
          <div className="flex justify-center items-center mt-10">
            <input type="text" placeholder= "搜索艺术品、收藏或创作者" className="pl-5 w-2/5 p-3 border-0 rounded-tl-lg rounded-bl-lg bg-[#1e1e2e] text-[#f5f6fa]"/>
            <div className="text-white w-24 bg-[#6c5ce7] p-3 text-center rounded-tr-lg rounded-br-lg ">搜索</div>
          </div>
          <div className="flex justify-center mt-10 mb-20">
            <div className="bg-[#6c5ce7] mr-10 text-white rounded-lg flex justify-center items-center px-6 py-3 font-bold cursor-pointer">立即铸造</div>
            <div className=" border-2 text-white border-[#6c5ce7]  rounded-lg flex justify-center items-center px-6 py-3 text-sm font-bold cursor-pointer">探索市场</div>
          </div>
        </div>
      </div>
      
      {/* 热门NFT */}
      <div className="mx-auto mt-14 max-w-full">
        <div className="text-4xl font-bold text-white text-center mb-8">热门NFT</div>
        <div className="mx-auto max-w-7xl grid grid-cols-3 gap-8 ">
          {
            Array.from({length:3}).map((_,i)=>(
              <div className=" bg-[#1e1e2e] rounded-2xl  flex flex-col overflow-hidden hover:-translate-y-2 cursor-pointer transition-all duration-300 ease-in-out">
                <img className="h-72 w-full rounded-t-2xl object-cover"
                src="https://images.unsplash.com/photo-1620641788421-7a1c342ea42e?ixlib=rb-1.2.1&auto=format&fit=crop&w=500&q=80" alt="" />
                <div className="p-5">
                  <div className="text-white font-bold mt-3 mb-2 text-lg">数字宇宙</div>
                  <div className="text-[#636e72] mb-2 ">创作者: alice</div>
                  <div className="flex justify-between items-center">
                    <div className="text-[#6c5ce7] font-bold">0.45 ETH</div>
                    <div className="btn btn-primary" onClick={()=>document.getElementById('buyModal').showModal()} >购买</div>
                  </div>
                </div>
              </div>
            ))
          }
          
        </div>
      </div>
       {/* You can open the modal using document.getElementById('ID').showModal() method */}
      <dialog id="buyModal" className="modal">
        <div className="modal-box bg-[#1e1e2e] text-white">
          <form method="dialog">
            {/* if there is a button in form, it will close the modal */}
            <button className="absolute right-4 top-2 hover:font-bold">✕</button>
          </form>
          <div className="flex flex-col p-3">
            <img className="rounded-2xl object-cover w-full h-80"
            src="https://images.unsplash.com/photo-1620641788421-7a1c342ea42e?ixlib=rb-1.2.1&auto=format&fit=crop&w=500&q=80" alt="" />
            <div className="mt-5 mx-auto text-3xl text-[#6c5ce7] font-bold">0.45 ETH</div>
            <div className="text-white font-bold mt-3 text-2xl"s>数字宇宙</div>
            <div className="mt-5 text-gray-400">上架时间: 2025-10-25 12:03:21</div>
            <div className="text-sm mt-5 text-gray-400">Alice</div>
            <div className="text-sm mt-5 text-gray-400">这是一个描绘数字宇宙的独特艺术品，展现了虚拟世界中的无限可能性。</div>
            <div className="flex justify-around items-center mt-5">
              <div className="btn btn-primary">立即购买</div>
              <div className="btn btn-primary">出价</div>
            </div>
          </div>
        </div>
      </dialog>
    </div>
  );
}

export default Main;
