import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.jsx";
//import { initWalletListener } from "./js/utils/walletListener"; // 引入监听工具

// 初始化钱包监听
//initWalletListener();

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
