
export const formatAddress = (address) => {
    if (!address) return '未连接';
    return `${address.slice(0, 6)}...${address.slice(-4)}`;
};

