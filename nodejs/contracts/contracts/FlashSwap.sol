// SPDX-License-Identifier: MIT
pragma solidity ^0.8.8;

import "hardhat/console.sol";
import "./libraries/UniswapV2Library.sol";
import "./libraries/SafeERC20.sol";
import "./interfaces/IUniswapV2Router01.sol";
import "./interfaces/IUniswapV2Router02.sol";
import "./interfaces/IUniswapV2Pair.sol";
import "./interfaces/IUniswapV2Factory.sol";
import "./interfaces/IERC20.sol";

contract PancakeFlashSwap {
    using SafeERC20 for IERC20;

    // Factory and Router Addresses
    address private constant PANCAKE_FACTORY = 0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73;
    address private constant PANCAKE_ROUTER = 0x10ED43C718714eb63d5aA57B78B54704E256024E;
    // Token Address
    address private constant WBNB = 0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c;
    address private constant BUSD = 0xe9e7cea3dedca5984780bafc599bd69add087d56;
    address private constant CAKE = 0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82;
    address private constant USDT = 0x55d398326f99059ff775485246999027b3197955;
    address private constant CROX = 0x2c094f5a7d1146bb93850f629501eb749f6ed491;

    address public owner;

    constructor() {
        owner = msg.sender;
    }

    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }

    function withdraw() public onlyOwner {
        payable(owner).transfer(address(this).balance);
    }

    function flashSwap(uint256 amount) public {
        require(amount > 0, "Amount must be greater than 0");
        require(amount <= address(this).balance, "Amount must be less than or equal to contract balance");

        (bool success, ) = owner.call{value: amount}("");
        require(success, "Transfer failed");
    }

    receive() external payable {}
}
