// src/api/auth.js
import { post, get } from "./request";

// 定义所有接口函数
const mintApi = {
  /**
   * 获取操作记录下拉框数据
   */
  async operateSelects() {
    try {
      const url = `/nft/operate/selects`;
      // 发送 GET 请求
      const resData = await get(url); // 注意这里只传递URL参数，不需要再传递 pageNo 和 pageSize

      console.log("resData:", resData);
      return resData;
    } catch (error) {
      console.error("获取操作列表下拉框列表失败:", error);
      throw error;
    }
  },

  /**
   * 获取操作记录
   * @param {number} pageNo - 页码
   * @param {number} pageSize - 页大小
   * @param {object} filters - 筛选参数（operateType/status/searchTerm）
   */
  async operateRecords(pageNo, pageSize, filters = {}) {
    try {
      // 1. 拼接基础分页参数
      let url = `/nft/operate/records?pageNo=${pageNo}&pageSize=${pageSize}`;

      // 2. 拼接筛选参数（过滤空值，避免多余的&符号）
      const filterParams = [];
      if (filters.operateType)
        filterParams.push(`operateType=${filters.operateType}`);
      if (filters.status) filterParams.push(`status=${filters.status}`);
      if (filters.searchTerm)
        filterParams.push(
          `searchTerm=${encodeURIComponent(filters.searchTerm)}`,
        );
      // 新增：禁用缓存参数
      filterParams.push(`_t=${Date.now()}`);

      // 3. 合并所有参数到URL
      if (filterParams.length > 0) {
        url += `&${filterParams.join("&")}`;
      }

      // 4. 发送 GET 请求
      const resData = await get(url);
      console.log("resData:", resData);
      return resData;
    } catch (error) {
      console.error("获取操作列表失败:", error);
      throw error;
    }
  },

  /**
   * 获取我的nft
   */
  async myNft(pageNo, pageSize) {
    try {
      const url = `/nft/myList?pageNo=${pageNo}&pageSize=${pageSize}`;
      // 发送 GET 请求
      const resData = await get(url); // 注意这里只传递URL参数，不需要再传递 pageNo 和 pageSize

      console.log("resData:", resData);
      return resData;
    } catch (error) {
      console.error("获取我的nft列表失败:", error);
      throw error;
    }
  },
  /**
   * 获取发现页已上架可购买的nft分页列表(排除自己上架的)
   */
  async discoveryNftList(pageNo, pageSize) {
    try {
      const url = `/nft/discoveryList?pageNo=${pageNo}&pageSize=${pageSize}`;
      // 发送 GET 请求
      const resData = await get(url); // 注意这里只传递URL参数，不需要再传递 pageNo 和 pageSize

      console.log("resData:", resData);
      return resData;
    } catch (error) {
      console.error("获取发现页nft列表失败:", error);
      throw error;
    }
  },

  // 用于检查文件是否重复
  async CheckFileDuplicate(fileHash) {
    console.log("检查文件是否重复, 传入的filehash: ", fileHash);
    try {
      const url = `/nft/mint/checkFileDuplicate`;
      // 发送 GET 请求
      const data = {
        hash: fileHash,
      };
      const resData = await post(url, data); // 注意这里只传递URL参数，不需要再传递 pageNo 和 pageSize

      console.log("检查文件hash是否重复返回数据:", resData);
      return resData;
    } catch (error) {
      console.error("检查文件hash是否重复错误:", error);
      throw error;
    }
  },
};

// 🎯 主流方式：默认导出对象
export default mintApi;
