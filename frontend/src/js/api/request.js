import axios from "axios";

// 🌟 从环境变量读取baseURL（优先级：环境变量 > 兜底值）
const BASE_URL = import.meta.env.VITE_BASE_URL || "http://localhost:8080";
console.log("当前配置的BASE_URL：", BASE_URL); // 新增：打印验证配置是否生效

// 白名单, 不用通过请求拦截器
const NO_AUTH_WHITELIST = ["/auth/login", "/auth/nonce", "/auth/captcha"];

// 消息提示函数（替换 antd message）
const showMessage = (text, type = "error") => {
  if (type === "error") {
    alert(text);
  }
};

// 错误处理器：新增过滤逻辑，排除地址切换导致的预期错误
const handleError = (message, type = "error", skipAlert = false) => {
  if (skipAlert) {
    console.log(`[忽略提示] ${message}`);
    return;
  }
  showMessage(message, type);
};

// 🎯 创建 axios 实例
const service = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
});

// ================ 🌟 请求拦截器 🌟 ================
service.interceptors.request.use((config) => {
  // 检查 URL 是否在白名单中
  const isPublic = NO_AUTH_WHITELIST.some((url) => config.url?.includes(url));

  // ✅ 核心修复：只有非白名单接口才做token/地址验证
  if (!isPublic) {
    const token = localStorage.getItem("token");
    const storedAddress = localStorage.getItem("connectedAddress");
    const currentWalletAddress = window.ethereum?.selectedAddress;

    // 地址验证逻辑（只在非白名单接口执行）
    if (token && storedAddress && currentWalletAddress) {
      const storedAddrLower = storedAddress.toLowerCase();
      const currentAddrLower = currentWalletAddress.toLowerCase();

      if (storedAddrLower !== currentAddrLower) {
        localStorage.removeItem("token");
        localStorage.removeItem("connectedAddress");
        console.log(
          `地址不匹配：存储地址(${storedAddrLower}) ≠ 当前地址(${currentAddrLower})，已清空旧token`,
        );
        // 地址不匹配时，直接拒绝请求，不执行后续逻辑
        return Promise.reject(new Error("地址切换导致登录失效"));
      }
    }

    // 无token则跳转登录（只在非白名单接口执行）
    if (!token) {
      console.log("无token，跳转到登录页");
      window.location.href = "/login";
      return Promise.reject(new Error("未登录（地址切换导致）"));
    }

    // 携带token（只在非白名单接口执行）
    config.headers.Authorization = `Bearer ${token}`;
  }

  // ✅ 白名单接口直接返回config，不做任何拦截
  return config;
});

// ================ 🌟 响应拦截器 🌟 ================
service.interceptors.response.use(
  (response) => {
    const res = response.data;
    if (res.code === 0) {
      return res.data;
    } else {
      if (res.code === 401) {
        handleError("登录已过期，请重新连接钱包");
        localStorage.removeItem("token");
        localStorage.removeItem("connectedAddress");
        window.location.href =
          "/login?redirect=" + encodeURIComponent(window.location.pathname);
      } else {
        handleError(res.message || "请求失败");
      }
      return Promise.reject(new Error(res.message || "Error"));
    }
  },
  (error) => {
    const isAddressSwitchError = error.message.includes("地址切换导致");

    if (error.response) {
      const status = error.response.status;
      switch (status) {
        case 401:
          if (isAddressSwitchError) {
            handleError("地址切换导致登录失效", "error", true);
          } else {
            handleError("未授权，请重新连接钱包");
          }
          localStorage.removeItem("token");
          localStorage.removeItem("connectedAddress");
          window.location.href = "/login";
          break;
        case 404:
          handleError("请求地址不存在");
          break;
        case 500:
          handleError("服务器内部错误");
          break;
        default:
          handleError(`请求错误: ${status}`);
      }
    } else if (error.request) {
      if (isAddressSwitchError) {
        handleError("地址切换导致请求中断", "error", true);
      } else {
        handleError("网络异常，请检查网络连接");
      }
    } else {
      if (isAddressSwitchError || error.message.includes("未登录")) {
        handleError(`请求配置错误（预期内）: ${error.message}`, "error", true);
      } else {
        handleError("请求配置错误");
      }
    }

    return Promise.reject(error);
  },
);

// 🎯 导出常用的请求方法
export const get = (url, params, config = {}) =>
  service.get(url, { params, ...config });

export const post = (url, data, config = {}) => service.post(url, data, config);

export const put = (url, data, config = {}) => service.put(url, data, config);

export const del = (url, params, config = {}) =>
  service.delete(url, { params, ...config });

export default service;
