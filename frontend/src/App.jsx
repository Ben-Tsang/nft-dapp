import { BrowserRouter, Routes, Route } from "react-router-dom";
import { useState } from "react";
import reactLogo from "./assets/react.svg";
import viteLogo from "/vite.svg";
import "./App.css";
import "./index.css"; // 关键：引入 Tailwind 样式
import Main from "@/components/Main";
import Header from "@/components/Header";
import Footer from "@/components/Footer";
import Mint from "@/components/Mint";
import { WalletProvider } from "@/context/walletContext"; // 全局变量
import MyNft from "./components/MyNft";
import Discovery from "./components/Discovery";
import OperationRecord from "./components/OperationRecord";
function App() {
  const [count, setCount] = useState(0);

  return (
    <WalletProvider>
      <BrowserRouter>
        <div className="w-full min-h-screen flex flex-col bg-[#12121a]">
          {/* 在header里修改路径 */}
          <Header></Header>
          <div className="pt-28  flex-1">
            <Routes>
              <Route path="/" element={<Main />} />
              <Route path="/discovery" element={<Discovery />} />
              <Route path="/mint" element={<Mint />} />
              <Route path="/myNft" element={<MyNft />} />
              <Route path="operationRecord" element={<OperationRecord />} />
            </Routes>
          </div>
          <Footer></Footer>
        </div>
      </BrowserRouter>
    </WalletProvider>
  );
}

export default App;
