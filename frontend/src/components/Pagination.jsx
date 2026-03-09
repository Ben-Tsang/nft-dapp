import React, { useState, useEffect } from "react";

function Pagination({
  totalItems = 0,
  itemsPerPage = 10,
  currentPage = 1,
  setCurrentPage,
  onPageSizeChange,
}) {
  const [pageSize, setPageSize] = useState(itemsPerPage);
  const [inputPage, setInputPage] = useState(currentPage);

  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize));
  const maxVisiblePages = 5;

  // 同步外部传入的 currentPage
  useEffect(() => {
    setInputPage(currentPage);
  }, [currentPage]);

  // 同步外部传入的 itemsPerPage
  useEffect(() => {
    setPageSize(itemsPerPage);
  }, [itemsPerPage]);

  // 处理每页显示数量变化
  const handlePageSizeChange = (e) => {
    const newSize = parseInt(e.target.value);
    setPageSize(newSize);

    const newTotalPages = Math.max(1, Math.ceil(totalItems / newSize));
    const newCurrentPage = Math.min(currentPage, newTotalPages);

    if (onPageSizeChange) {
      onPageSizeChange(newSize);
    }

    setCurrentPage(newCurrentPage);
  };

  // 跳转到指定页
  const goToPage = (page) => {
    const validPage = Math.max(1, Math.min(page, totalPages));
    setCurrentPage(validPage);
  };

  // 处理输入框跳转
  const handleInputSubmit = (e) => {
    e.preventDefault();
    if (inputPage >= 1 && inputPage <= totalPages) {
      goToPage(inputPage);
    }
  };

  // 生成页码按钮数组
  const getPageNumbers = () => {
    const pages = [];

    if (totalPages <= maxVisiblePages) {
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      let startPage = Math.max(
        1,
        currentPage - Math.floor(maxVisiblePages / 2)
      );
      let endPage = startPage + maxVisiblePages - 1;

      if (endPage > totalPages) {
        endPage = totalPages;
        startPage = Math.max(1, endPage - maxVisiblePages + 1);
      }

      if (startPage > 1) {
        pages.push(1);
        if (startPage > 2) {
          pages.push("...");
        }
      }

      for (let i = startPage; i <= endPage; i++) {
        pages.push(i);
      }

      if (endPage < totalPages) {
        if (endPage < totalPages - 1) {
          pages.push("...");
        }
        pages.push(totalPages);
      }
    }

    return pages;
  };

  if (totalItems === 0) return null;

  return (
    <div className="flex flex-col md:flex-row items-center justify-between space-y-4 md:space-y-0">
      {/* 左侧：每页显示数量选择 */}
      <div className="flex items-center space-x-3">
        <span className="text-sm text-gray-300">每页显示:</span>
        <select
          value={pageSize}
          onChange={handlePageSizeChange}
          className="px-3 py-1.5 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
        >
          <option value="6">6 条</option>
          <option value="10">10 条</option>
          <option value="12">12 条</option>
          <option value="15">15 条</option>
          <option value="20">20 条</option>
          <option value="30">30 条</option>
        </select>
        <span className="text-sm text-gray-300">
          共 <span className="text-purple-400 font-semibold">{totalItems}</span>{" "}
          条记录
        </span>
      </div>

      {/* 中间：页码导航 */}
      <div className="flex flex-col items-center space-y-3 md:space-y-0 md:flex-row md:space-x-4">
        {/* 总页数信息 */}
        <div className="text-sm text-gray-300">
          第 <span className="font-bold text-purple-400">{currentPage}</span> 页
          / 共 <span className="font-bold text-purple-400">{totalPages}</span>{" "}
          页
        </div>

        {/* 页码按钮 */}
        <div className="flex items-center space-x-1">
          {/* 第一页按钮 */}
          <button
            onClick={() => goToPage(1)}
            disabled={currentPage === 1}
            className="px-3 py-1.5 rounded-lg bg-gray-800 text-gray-300 hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200 text-sm"
            title="第一页"
          >
            ««
          </button>

          {/* 上一页按钮 */}
          <button
            onClick={() => goToPage(currentPage - 1)}
            disabled={currentPage === 1}
            className="px-3 py-1.5 rounded-lg bg-gray-800 text-gray-300 hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200 text-sm"
            title="上一页"
          >
            «
          </button>

          {/* 页码数字按钮 */}
          {getPageNumbers().map((page, index) => (
            <React.Fragment key={index}>
              {page === "..." ? (
                <span className="px-2 py-1 text-gray-500 text-sm">...</span>
              ) : (
                <button
                  onClick={() => goToPage(page)}
                  className={`px-3 py-1.5 min-w-[40px] rounded-lg transition-colors duration-200 text-sm ${
                    currentPage === page
                      ? "bg-gradient-to-r from-purple-600 to-purple-700 text-white font-bold shadow-lg"
                      : "bg-gray-800 text-gray-300 hover:bg-gray-700"
                  }`}
                >
                  {page}
                </button>
              )}
            </React.Fragment>
          ))}

          {/* 下一页按钮 */}
          <button
            onClick={() => goToPage(currentPage + 1)}
            disabled={currentPage === totalPages}
            className="px-3 py-1.5 rounded-lg bg-gray-800 text-gray-300 hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200 text-sm"
            title="下一页"
          >
            »
          </button>

          {/* 最后一页按钮 */}
          <button
            onClick={() => goToPage(totalPages)}
            disabled={currentPage === totalPages}
            className="px-3 py-1.5 rounded-lg bg-gray-800 text-gray-300 hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200 text-sm"
            title="最后一页"
          >
            »»
          </button>
        </div>

        {/* 快速跳转 */}
        <form
          onSubmit={handleInputSubmit}
          className="flex items-center space-x-2"
        >
          <span className="text-sm text-gray-300">跳转到:</span>
          <input
            type="number"
            min="1"
            max={totalPages}
            value={inputPage}
            onChange={(e) => setInputPage(parseInt(e.target.value) || 1)}
            className="w-16 px-2 py-1.5 bg-gray-800 border border-gray-700 rounded text-white text-sm text-center focus:outline-none focus:ring-2 focus:ring-purple-500"
          />
          <span className="text-sm text-gray-300">页</span>
          <button
            type="submit"
            className="px-3 py-1.5 bg-gradient-to-r from-purple-600 to-purple-700 hover:from-purple-700 hover:to-purple-800 text-white rounded-lg transition-all duration-200 text-sm"
          >
            跳转
          </button>
        </form>
      </div>

      {/* 右侧：当前页记录范围 */}
      <div className="text-sm text-gray-300">
        显示{" "}
        <span className="text-purple-400 font-semibold">
          {Math.min((currentPage - 1) * pageSize + 1, totalItems)} -{" "}
          {Math.min(currentPage * pageSize, totalItems)}
        </span>{" "}
        条
      </div>
    </div>
  );
}

export default Pagination;
