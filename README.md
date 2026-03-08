
## 项目介绍

本项目是一个基于Web3技术栈开发的NFT发行与交易全栈平台，支持NFT铸造、上架、下架、改价、购买等核心功能，实现链上合约与链下服务的数据同步。

项目定位：Web3全栈DApp，覆盖智能合约开发、后端服务、前端交互、工程化部署全流程，适配MetaMask钱包，支持IPFS存储。

## 技术栈

### 核心技术栈

- **智能合约**：Solidity 0.8+、Hardhat、OpenZeppelin（ERC721标准）

- **链下服务**：Go、Gin、Gorm、ethclient（链上事件监听）、PostgreSQL（数据存储）、Redis（缓存优化）

- **前端交互**：React、Ethers.js、MetaMask集成

- **存储与部署**：IPFS（Lighthouse）、Docker、Docker Compose、云服务部署

## 核心功能

1.  **合约功能**：基于ERC721实现独立NFT合约与Market合约，支持铸造、上架、下架、改价、购买逻辑。

2.  **链下服务**：基于Go+Gin开发后端接口，通过ethclient实时监听链上自定义事件，解析参数后同步至PostgreSQL，支持定时同步，保障链上链下数据一致性。

3.  **前端交互**：React+Ethers.js开发DApp，支持MetaMask钱包连接、链切换、签名授权，实现NFT列表、详情、交易记录可视化展示，支持IPFS文件存储与预览。

4.  **工程化部署**：提供Dockerfile与docker-compose.yml配置，实现前端、后端、数据库一键容器化部署，保障环境一致性，支持快速上线与维护。

## 快速开始

### 前提条件

- 安装MetaMask钱包，并切换至Sepolia测试网

- 安装Docker、Docker Compose（用于本地部署）

- 获取Sepolia测试网ETH（用于合约交互、NFT铸造）或本地hardhat测试链测试


## 项目结构

/contracts       # 智能合约源码（Solidity）
/hardhat         # Hardhat配置、测试用例
/backend         # Go+Gin后端服务
/frontend        # React前端DApp
/docker          # Docker配置文件
/docker-compose.yml  # 容器化部署配置


