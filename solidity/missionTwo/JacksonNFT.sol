// SPDX-License-Identifier: MIT

pragma solidity ^0.8.21;
import "@openzeppelin/contracts/utils/Strings.sol";

// ### ✅ 作业2：在测试网上发行一个图文并茂的 NFT
// 任务目标
// 1. 使用 Solidity 编写一个符合 ERC721 标准的 NFT 合约。
// 2. 将图文数据上传到 IPFS，生成元数据链接。
// 3. 将合约部署到以太坊测试网（如 Goerli 或 Sepolia）。
// 4. 铸造 NFT 并在测试网环境中查看。
// 任务步骤
// 1. 编写 NFT 合约
//   - 使用 OpenZeppelin 的 ERC721 库编写一个 NFT 合约。
//   - 合约应包含以下功能：
//   - 构造函数：设置 NFT 的名称和符号。
//   - mintNFT 函数：允许用户铸造 NFT，并关联元数据链接（tokenURI）。
//   - 在 Remix IDE 中编译合约。
// 2. 准备图文数据
//   - 准备一张图片，并将其上传到 IPFS（可以使用 Pinata 或其他工具）。
//   - 创建一个 JSON 文件，描述 NFT 的属性（如名称、描述、图片链接等）。
//   - 将 JSON 文件上传到 IPFS，获取元数据链接。
//   - JSON文件参考 https://docs.opensea.io/docs/metadata-standards
// 3. 部署合约到测试网
//   - 在 Remix IDE 中连接 MetaMask，并确保 MetaMask 连接到 Goerli 或 Sepolia 测试网。
//   - 部署 NFT 合约到测试网，并记录合约地址。
// 4. 铸造 NFT
//   - 使用 mintNFT 函数铸造 NFT：
//   - 在 recipient 字段中输入你的钱包地址。
//   - 在 tokenURI 字段中输入元数据的 IPFS 链接。
//   - 在 MetaMask 中确认交易。
// 5. 查看 NFT
//   - 打开 OpenSea 测试网 或 Etherscan 测试网。
//   - 连接你的钱包，查看你铸造的 NFT。




contract JacksonNFT {
    using Strings for uint256;
    string public name;
    string public symbol;
    mapping(uint => address) private _owners;
    mapping(address => uint) private _balances;
    mapping(uint => address) private _tokenApprovals;
    mapping(address => mapping(address => bool)) private _operatorApprovals;

    event Transfer(address indexed from, address indexed to, uint256 indexed tokenId);
    event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId);
    event ApprovalForAll(address indexed owner, address indexed operator, bool approved);


    constructor(string memory name_, string memory symbol_)  {
        name = name_;
        symbol = symbol_;
    }

    function balanceOf(address owner) public view returns (uint256) {
        require(owner != address(0), "owner is zero address");
        return _balances[owner];
    }

    function ownerOf(uint256 tokenId) public view returns (address) {
        require(tokenId != 0, "tokenId is zero");
        address owner = _owners[tokenId];
        return owner;
    }

    function approve(address to, uint256 tokenId)  external {
        address owner = ownerOf(tokenId);
        require(msg.sender == owner || _operatorApprovals[owner][msg.sender], "not owner or approved for all");
        _approve(owner, to, tokenId);
    }

    function _approve( address owner, address to, uint tokenId ) private {
        _tokenApprovals[tokenId] = to;
        emit Approval(owner, to, tokenId);
    }

    function isApprovedForAll(address owner, address operator) external view returns (bool) {
        return _operatorApprovals[owner][operator];
    }

    function setApprovalForAll(address operator, bool approved) external {
        _operatorApprovals[msg.sender][operator] = approved;
        emit ApprovalForAll(msg.sender, operator, approved);
    }

    function getApproved(uint tokenId) external view returns (address) {
        require(_owners[tokenId] != address(0), "token does not exist");
        return _tokenApprovals[tokenId];
    }

    function _isApprovedOrOwner(address owner, address spender, uint256 tokenId) private view returns (bool) {
        return (spender == owner || _tokenApprovals[tokenId] == spender || _operatorApprovals[owner][spender]);
    }


    function mintNFT(address to, uint tokenId) public {
        require(to != address(0), "mint to zero address");
        require(_owners[tokenId] == address(0), "token already minted");
        _balances[to] += 1;
        _owners[tokenId] = to;
    }


    function _baseURI() internal pure virtual returns (string memory) {
        return "https://gateway.pinata.cloud/ipfs/bafkreib3txgdb5jptxcl6edvqfmn4xy7quzogz5cn67vtfpdhiq7ymna74";
    }
    
    function tokenURI(uint256 tokenId) public view virtual returns(string memory) {
        require(_owners[tokenId] != address(0), "token does not exist");
        string memory baseURI = _baseURI();
        return bytes(baseURI).length > 0 ? string(abi.encodePacked(baseURI, tokenId.toString())) : "";
    }

    function _transfer(address owner, address from, address to, uint256 tokenId) private  {
        require(from == owner, "not owner");
        require(to != address(0), "transfer to the zero address");
        _approve(owner, address(0), tokenId);
        _balances[from] -= 1;
        _balances[to] += 1;
        _owners[tokenId] = to;
        emit Transfer(from, to, tokenId);
    }

    function _safeTransfer(address owner, address from, address to, uint256 tokenId, bytes memory data) private {

    }

    function _checkOnERC721Received(address from, address to, uint256 tokenId, bytes memory data) private {
        if (to.code.length > 0) {
            try IERC721Receiver(to).onERC721Received(msg.sender, from, tokenId, data) returns(bytes4  val) {
                if (val != IERC721Receiver.onERC721Received.selector) {
                    revert("ERC721: transfer to non ERC721Receiver implementer");
                }
            } catch {
                revert("ERC721: transfer to non ERC721Receiver implementer");
            }
        }
    }

    function _burn(uint tokenId) internal virtual {
        address owner = ownerOf(tokenId);
        require(msg.sender == owner, "not owner of token");
        _balances[owner] -= 1;
        _owners[tokenId] = address(0);
        emit Transfer(owner, address(0), tokenId);
    }
    

}

interface IERC721Receiver {

    function onERC721Received(
        address operator,
        address from,
        uint256 tokenId,
        bytes calldata data
    ) external returns (bytes4);
}