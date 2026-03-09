// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {ERC721} from "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {ERC721URIStorage} from "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";

contract NFT is ERC721, ERC721URIStorage, Ownable {
    uint256 private _nextTokenId;

    event NFTMinted(
        address indexed to,
        uint256 indexed tokenId,
        string name,
        string description,
        string tokenURI
    );

    constructor() ERC721("NFT", "MTK") Ownable(msg.sender) {}

    function safeMint(
        string calldata name,
        string calldata description,
        string calldata _tokenURI
    ) external returns (uint256) {
        require(bytes(_tokenURI).length > 0, "NFTMint: tokenURI cannot be empty");

        uint256 tokenId = _nextTokenId++;
        
        // 修复：先设置 URI，再 mint（避免某些版本的栈深度问题）
        _setTokenURI(tokenId, _tokenURI);
        _safeMint(msg.sender, tokenId);

        emit NFTMinted(msg.sender, tokenId, name, description, _tokenURI);
        return tokenId;
    }

    // 以下函数是必要的重写
    function tokenURI(uint256 tokenId) 
        public 
        view 
        override(ERC721, ERC721URIStorage) 
        returns (string memory) 
    {
        return super.tokenURI(tokenId);
    }

    function supportsInterface(bytes4 interfaceId) 
        public 
        view 
        override(ERC721, ERC721URIStorage) 
        returns (bool) 
    {
        return super.supportsInterface(interfaceId);
    }
}