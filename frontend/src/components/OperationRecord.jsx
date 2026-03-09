import React, { useState, useEffect, useCallback } from "react";
import { useWallet } from "@/context/WalletContext";
import { ethers } from "ethers";
import api from "@/js/api/api";
import Pagination from "./Pagination";
import {
  useReactTable,
  getCoreRowModel,
  createColumnHelper,
  flexRender,
  getSortedRowModel,
  getPaginationRowModel,
} from "@tanstack/react-table";
import {
  FaShoppingCart,
  FaPlusCircle,
  FaMinusCircle,
  FaExchangeAlt,
  FaTimesCircle,
  FaClock,
  FaSearch,
  FaSyncAlt,
} from "react-icons/fa";

const STATUS_MAP = {
  success: { label: "成功", color: "bg-green-500" },
  pending: { label: "处理中", color: "bg-yellow-500" },
  failed: { label: "失败", color: "bg-red-500" },
};

const formatTime = (timeStr) =>
  timeStr ? new Date(timeStr).toLocaleString("zh-CN") : "-";
const formatAmount = (amount) => {
  try {
    return amount && amount !== ""
      ? `${ethers.formatEther(amount)} ETH`
      : "0 ETH";
  } catch (e) {
    return "0 ETH";
  }
};
const formatHash = (hash) =>
  hash && hash !== "" ? `${hash.slice(0, 6)}...${hash.slice(-4)}` : "-";

const fetchOperationRecords = async (params) => {
  const {
    pageNo = 1,
    pageSize = 10,
    operateType = "",
    status = "",
    searchTerm = "",
  } = params;

  try {
    const data = await api.operateRecords(pageNo, pageSize, {
      operateType,
      status,
      searchTerm,
    });
    return data;
  } catch (error) {
    console.error("请求操作记录失败：", error);
    alert("获取操作记录失败：" + error.message);
    return { records: [], total: 0, pages: 0, size: 10, current: 1 };
  }
};

function OperationRecord() {
  const [loading, setLoading] = useState(false);
  const [tableData, setTableData] = useState([]);
  const [pagination, setPagination] = useState({
    total: 0,
    pages: 0,
    size: 10,
    current: 1,
  });

  const [searchTerm, setSearchTerm] = useState("");
  const [selectedType, setSelectedType] = useState("");
  const [amountRange, setAmountRange] = useState([0, 10]);
  const [selectedStatus, setSelectedStatus] = useState("");

  const [selectOptions, setSelectOptions] = useState({
    operateType: [],
    status: [],
  });
  const [selectLoading, setSelectLoading] = useState(false);

  const getSelects = useCallback(async () => {
    setSelectLoading(true);
    try {
      const res = await api.operateSelects({ params: { _t: Date.now() } });

      const operateTypeList = res.operateType.map((item) => ({
        value: item.Value,
        label: item.Label,
      }));
      const statusList = res.operateStatus.map((item) => ({
        value: item.Value,
        label: item.Label,
      }));

      setSelectOptions((prev) => ({
        ...prev,
        operateType: operateTypeList,
        status: statusList,
      }));
    } catch (error) {
      console.error("获取下拉列表失败：", error);
      alert("获取筛选选项失败：" + error.message);
      setSelectOptions((prev) => ({
        ...prev,
        operateType: [
          { value: "mint", label: "NFT铸造" },
          { value: "listed", label: "NFT上架" },
          { value: "unlisted", label: "NFT下架" },
          { value: "buy", label: "NFT购买" },
          { value: "transfer", label: "NFT转让" },
          { value: "set_price", label: "NFT改价" },
        ],
        status: [
          { value: "success", label: "成功" },
          { value: "failed", label: "失败" },
        ],
      }));
    } finally {
      setSelectLoading(false);
    }
  }, []);

  // 修复：loadData 可以手动传入筛选条件
  const loadData = useCallback(
    async (currentPage = 1, searchArgs = null) => {
      setLoading(true);

      const params = searchArgs || {
        pageNo: currentPage,
        pageSize: pagination.size,
        operateType: selectedType,
        status: selectedStatus,
        searchTerm: searchTerm,
      };

      const result = await fetchOperationRecords(params);

      const adaptedData = result.records.map((item) => ({
        id: item.id,
        type: item.operate_type,
        typeLabel: item.operate_type_label,
        nftName: `NFT #${item.token_id}`,
        nftId: item.token_id,
        amount: item.amount,
        txHash: item.tx_hash,
        status: item.status,
        createTime: item.operate_at || item.created_at,
        contractAddress: item.contract_address,
        userAddress: item.user_address,
      }));

      setTableData(adaptedData);
      setPagination({
        total: result.total,
        pages: result.pages,
        size: result.size,
        current: result.current,
      });
      setLoading(false);
    },
    [pagination.size, selectedType, selectedStatus, searchTerm],
  );

  const handleSearch = () => {
    loadData(1);
  };

  const handlePageChange = (newPage) => {
    loadData(newPage);
  };

  // 修复：一次点击就重置 + 立即请求
  const resetFilters = () => {
    // 先清空状态
    setSearchTerm("");
    setSelectedType("");
    setSelectedStatus("");
    setAmountRange([0, 10]);

    // 直接传清空后的参数，不依赖异步状态
    loadData(1, {
      pageNo: 1,
      pageSize: pagination.size,
      operateType: "",
      status: "",
      searchTerm: "",
    });
  };

  useEffect(() => {
    getSelects();
  }, [getSelects]);

  const columnHelper = createColumnHelper();
  const columns = [
    columnHelper.accessor("typeLabel", {
      header: "操作类型",
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          {row.getValue("typeLabel")}
        </div>
      ),
    }),
    columnHelper.accessor("nftName", {
      header: "NFT信息",
      cell: ({ row }) => {
        const { nftName, contractAddress, tokenId } = row.original;
        return (
          <div>
            <div>{nftName || "-"}</div>
            <div className="text-xs text-gray-400">
              {contractAddress
                ? `${contractAddress.slice(0, 6)}...${contractAddress.slice(-4)}`
                : ""}
              {tokenId ? ` · #${tokenId}` : ""}
            </div>
          </div>
        );
      },
    }),
    columnHelper.accessor("amount", {
      header: "金额",
      cell: ({ row }) => formatAmount(row.getValue("amount")),
    }),
    columnHelper.accessor("txHash", {
      header: "交易哈希",
      cell: ({ row }) => (
        <span className="font-mono text-sm">
          {formatHash(row.getValue("txHash"))}
        </span>
      ),
    }),
    columnHelper.accessor("status", {
      header: "状态",
      cell: ({ row }) => {
        const st = STATUS_MAP[row.getValue("status")] || STATUS_MAP.pending;
        return (
          <span
            className={`px-2 py-1 text-xs rounded-full ${st.color} text-white`}
          >
            {st.label}
          </span>
        );
      },
    }),
    columnHelper.accessor("createTime", {
      header: "操作时间",
      cell: ({ row }) => formatTime(row.getValue("createTime")),
    }),
  ];

  const table = useReactTable({
    data: tableData,
    columns: columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    pageCount: pagination.pages,
    state: {
      sorting: [{ id: "createTime", desc: true }],
      pagination: {
        pageIndex: pagination.current - 1,
        pageSize: pagination.size,
      },
    },
    manualPagination: true,
    manualSorting: true,
  });

  if (loading || selectLoading) {
    return (
      <div className="flex justify-center items-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-purple-500"></div>
        <span className="ml-4 text-lg">加载中...</span>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-6 text-white">
      <h1 className="text-3xl font-bold text-white mb-5">我的操作记录</h1>

      <div className="mb-6 flex flex-wrap items-center gap-4">
        <div className="relative flex-1 max-w-md">
          <input
            type="text"
            placeholder="搜索NFT名称/ID..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full p-2 pl-10 pr-8 rounded-md bg-gray-700 text-white"
          />
          <FaSearch className="absolute left-3 top-3 text-gray-400" />
          {searchTerm && (
            <button
              onClick={() => setSearchTerm("")}
              className="absolute right-3 top-2 text-gray-400 hover:text-white"
            >
              ✕
            </button>
          )}
        </div>

        <select
          value={selectedType}
          onChange={(e) => setSelectedType(e.target.value)}
          className="p-2 rounded-md bg-gray-700 text-white min-w-[120px]"
          disabled={selectLoading}
        >
          <option value="">所有操作类型</option>
          {selectOptions.operateType.map((item) => (
            <option key={item.value} value={item.value}>
              {item.label}
            </option>
          ))}
        </select>

        <select
          value={selectedStatus}
          onChange={(e) => setSelectedStatus(e.target.value)}
          className="p-2 rounded-md bg-gray-700 text-white min-w-[100px]"
        >
          <option value="">所有状态</option>
          {selectOptions.status.map((item) => (
            <option key={item.value} value={item.value}>
              {item.label}
            </option>
          ))}
        </select>

        <button
          onClick={resetFilters}
          className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-500"
          disabled={loading}
        >
          重置筛选
        </button>

        <button
          onClick={() => loadData(pagination.current)}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-500 flex items-center gap-2"
          disabled={loading}
        >
          <FaSyncAlt className={loading ? "animate-spin" : ""} />
          刷新数据
        </button>

        <button
          onClick={handleSearch}
          className="px-6 py-2 bg-purple-600 text-white rounded hover:bg-purple-500 flex items-center gap-2"
          disabled={loading}
        >
          <FaSearch />
          搜索
        </button>
      </div>

      {/* ========== 还原原有样式：bg-[#1d1d3b] ========== */}
      <div className="bg-[#1d1d3b] rounded-lg shadow-lg overflow-hidden">
        {tableData.length === 0 && (
          <div className="py-12 text-center text-gray-400">暂无操作记录</div>
        )}
        {tableData.length > 0 && (
          <table className="min-w-full divide-y divide-gray-700">
            <thead className="bg-gray-800/50">
              {table.getHeaderGroups().map((headerGroup) => (
                <tr key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <th
                      key={header.id}
                      className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase cursor-pointer"
                      onClick={() => table.toggleSorting(header.id)}
                    >
                      {flexRender(
                        header.column.columnDef.header,
                        header.getContext(),
                      )}
                      <span className="ml-1">
                        {header.isSorted
                          ? header.isSortedDesc
                            ? " 🔽"
                            : " 🔼"
                          : ""}
                      </span>
                    </th>
                  ))}
                </tr>
              ))}
            </thead>
            <tbody className="divide-y divide-gray-700">
              {table.getRowModel().rows.map((row) => (
                <tr key={row.id} className="hover:bg-gray-700/30">
                  {row.getVisibleCells().map((cell) => (
                    <td
                      key={cell.id}
                      className="px-6 py-4 whitespace-nowrap text-sm"
                    >
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext(),
                      )}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <div className="mt-4 flex justify-between items-center">
        <button
          onClick={() => handlePageChange(pagination.current - 1)}
          disabled={pagination.current <= 1 || loading}
          className="px-4 py-2 bg-purple-500 text-white rounded disabled:opacity-50"
        >
          上一页
        </button>
        <div>
          <span>
            共 {pagination.total} 条记录 | 第 {pagination.current} 页 /{" "}
            {pagination.pages} 页
          </span>
        </div>
        <button
          onClick={() => handlePageChange(pagination.current + 1)}
          disabled={pagination.current >= pagination.pages || loading}
          className="px-4 py-2 bg-purple-500 text-white rounded disabled:opacity-50"
        >
          下一页
        </button>
      </div>
    </div>
  );
}

export default OperationRecord;
