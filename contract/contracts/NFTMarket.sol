// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.28;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {IERC721} from "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import {
    IERC721Metadata
} from "@openzeppelin/contracts/token/ERC721/extensions/IERC721Metadata.sol";

interface INFT is IERC721, IERC721Metadata { function safeMint(string memory uri) external returns (uint256); }

contract NFTMarket is Ownable {
    // ✅ 修复1：使用正确的地址格式（40个十六进制字符）
    INFT public immutable nft;

    uint256[] private _listed;
    mapping(uint256 => bool) public isListed;
    mapping(uint256 => uint256) public listPrice;
    mapping(uint256 => uint256) private _listIndex;

    // ✅ 修复2：通过构造函数传入 NFT 合约地址，而不是硬编码

    constructor(address nftContract) Ownable(msg.sender) {
        nft = INFT(nftContract);
    }

    // 上架事件
    // indexed关键字：让事件可被按参数索引查询（后端监听必备）
    event ItemListed(
        address indexed seller,
        address indexed nftContract,
        uint256 indexed tokenId,
        uint256 price,
        uint256 listedAt
    );

    // 下架事件
    event ItemUnlisted(
        address indexed seller,
        uint256 indexed tokenId,
        uint256 unlistedAt
    );

    // 转账事件
    event Transfer(
        address indexed seller,
        address indexed to,
        uint256 indexed tokenId,
        uint256 transferAt
    );

    // 修改价格事件
    event SetPrice(
        address indexed owner,
        uint256 indexed tokenId,
        uint256 price,
        uint256 setAt
    );

    // 购买事件
    event Buy(
        address indexed buyer,
        uint256 indexed tokenId,
        uint256 price,
        uint256 buyAt
    );

    // =================================================================================================================
    // 上架
    function list(uint256 tokenId, uint256 price) external {
        require(nft.ownerOf(tokenId) == msg.sender, "not owner");
        require(!isListed[tokenId], "already listed");
        require(price > 0, "invalid price");
        require(
            nft.getApproved(tokenId) == address(this) ||
                nft.isApprovedForAll(msg.sender, address(this)),
            "NFTMarket: approve this contract first (approve single NFT or setApprovalForAll)"
        );

        _listIndex[tokenId] = _listed.length;
        _listed.push(tokenId);
        isListed[tokenId] = true;
        listPrice[tokenId] = price;
        // 广播事件
        emit ItemListed(
            msg.sender,
            address(nft),
            tokenId,
            price,
            block.timestamp // 上架时间戳
        );
    }

    // =================================================================================================================
    // 下架
    function unlist(uint256 tokenId) external {
        require(isListed[tokenId], "not listed");
        require(
            nft.ownerOf(tokenId) == msg.sender || owner() == msg.sender,
            "no permission"
        );

        uint256 idx = _listIndex[tokenId];
        uint256 lastId = _listed[_listed.length - 1];

        // ✅ 修复3：检查是否最后一个元素
        if (lastId != tokenId) {
            _listed[idx] = lastId;
            _listIndex[lastId] = idx;
        }

        _listed.pop();
        delete _listIndex[tokenId];
        delete isListed[tokenId];
        delete listPrice[tokenId];
        // 发布事件
        emit ItemUnlisted(
            msg.sender,
            tokenId,
            block.timestamp
        );
    }

    // =================================================================================================================
    // 购买
    function buy(uint256 tokenId) external payable {
        require(isListed[tokenId], "not listed");
        uint256 price = listPrice[tokenId];
        require(msg.value == price, "bad price");

        address seller = nft.ownerOf(tokenId);
        require(seller != msg.sender, "self buy");

        // 转账 NFT
        nft.safeTransferFrom(seller, msg.sender, tokenId);

        // 转账 ETH
        (bool success, ) = payable(seller).call{value: msg.value}("");
        require(success, "pay fail");

        // 下架处理
        uint256 idx = _listIndex[tokenId];
        uint256 lastId = _listed[_listed.length - 1];

        if (lastId != tokenId) {
            _listed[idx] = lastId;
            _listIndex[lastId] = idx;
        }

        _listed.pop();
        delete _listIndex[tokenId];
        delete isListed[tokenId];
        delete listPrice[tokenId];
        // 发布购买事件
        emit Buy(
            msg.sender,
            tokenId,
            price,
            block.timestamp
        );
        // 发布下架事件
        emit ItemUnlisted(
            msg.sender,
            tokenId,
            block.timestamp
        );
    }
    // =================================================================================================================
    // 修改价格
    function setPrice(uint256 tokenId, uint256 price) external{
        require(nft.ownerOf(tokenId) == msg.sender, "not owner");
        require(isListed[tokenId], "not listed"); 
        require(price > 0, "invalid price");
        require(
            nft.getApproved(tokenId) == address(this) ||
                nft.isApprovedForAll(msg.sender, address(this)),
            "NFTMarket: approve this contract first (approve single NFT or setApprovalForAll)"
        );
        listPrice[tokenId] = price;
        // 发布修改价格事件
        emit SetPrice(
            msg.sender,
            tokenId,
            price,
            block.timestamp
        );

    }

    // =================================================================================================================
    // 分页获取发现页
    function listTokens(
        uint256 pageNum,
        uint256 pageSize
    )
        external
        view
        returns (
            uint256[] memory tokenIds,
            string[] memory tokenUris,
            uint256[] memory prices
        )
    {
        require(pageNum >= 1 && pageSize > 0, "bad paging");
        uint256 total = _listed.length;
        uint256 start = (pageNum - 1) * pageSize;
        if (start >= total) {
            return (new uint256[](0), new string[](0), new uint256[](0));
        }
        uint256 end = start + pageSize;
        if (end > total) end = total;
        uint256 len = end - start;

        tokenIds = new uint256[](len);
        tokenUris = new string[](len);
        prices = new uint256[](len);

        for (uint256 i = 0; i < len; i++) {
            uint256 tokenId = _listed[start + i];
            tokenIds[i] = tokenId;
            tokenUris[i] = nft.tokenURI(tokenId);
            prices[i] = listPrice[tokenId];
        }
    }

    // =================================================================================================================
    function totalListed() external view returns (uint256) {
        return _listed.length;
    }
}
