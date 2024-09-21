// SPDX-License-Identifier: MIT
pragma solidity ~0.6.6;

import "hardhat/console.sol";
import "./protocols/uniswap/UniswapV2Library.sol";
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
    address private constant BUSD = 0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56;
    address private constant CAKE = 0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82;
    address private constant USDT = 0x55d398326f99059fF775485246999027B3197955;
    address private constant CROX = 0x2c094F5A7D1146BB93850f629501eB749f6Ed491;

    // Trade variables
    uint256 private deadline = block.timestamp + 1 days;
    uint256 private constant MAX_UINT = 115792089237316195423570985008687907853269984665640564039457584007913129639935;

    address public owner;

    constructor() public {
        owner = msg.sender;
    }

    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }

    function withdraw() public onlyOwner {
        payable(owner).transfer(address(this).balance);
    }

    // FUND SWAP CONTRACT
    // Provide a function to allow the contract to receive funds
    function fundFlashSwapContract(address _owner, address _token, uint256 _amount) public {
        IERC20(_token).safeTransferFrom(_owner, address(this), _amount);
    }

    // GET CONTRACT BALANCE
    // Get the balance of each token in the contract
    function getBalanceOfToken(address _address) public view returns (uint256) {
        return IERC20(_address).balanceOf(address(this));
    }

    // PLACE A TRADE
    // Execute placing a trade
    function placeTrade(address _fromToken, address _toToken, uint256 _amountIn) private returns (uint256) {
        address pair = IUniswapV2Factory(PANCAKE_FACTORY).getPair(_fromToken, _toToken);
        require(pair != address(0), "This pool does not exist");

        address[] memory path = new address[](2);
        path[0] = _fromToken;
        path[1] = _toToken;
        uint256 amountRequired = IUniswapV2Router01(PANCAKE_ROUTER).getAmountsOut(_amountIn, path)[1];
        console.log("Amount Required: %s", amountRequired);

        uint amountReceived = IUniswapV2Router02(PANCAKE_ROUTER).swapExactTokensForTokens(
            _amountIn, amountRequired, path, address(this), deadline
        )[1];
        console.log("Amount Received: %s", amountReceived);
        require(amountReceived > 0, "Aborted Tx: Trade returned zero");

        return amountReceived;
    }

    // INITIATE ARBITRAGE
    // Begins receiving loan to engage performing arbitrage trades
    function startArbitrage(address _tokenBorrow, uint256 _amount) public {
        IERC20(BUSD).safeApprove(address(PANCAKE_ROUTER), MAX_UINT);
        IERC20(CAKE).safeApprove(address(PANCAKE_ROUTER), MAX_UINT);
        IERC20(USDT).safeApprove(address(PANCAKE_ROUTER), MAX_UINT);
        IERC20(CROX).safeApprove(address(PANCAKE_ROUTER), MAX_UINT);
        console.log("Message sender: %s", msg.sender);

        // Return the Factory address for combined tokens
        address pair = IUniswapV2Factory(PANCAKE_FACTORY).getPair(_tokenBorrow, WBNB);
        require(pair != address(0), "This pool does not exist");
        // Figure out which token (0 or 1) has the amount and assign
        address token0 = IUniswapV2Pair(pair).token0();
        address token1 = IUniswapV2Pair(pair).token1();
        uint256 amount0Out = token0 == _tokenBorrow ? _amount : 0;
        uint256 amount1Out = token1 == _tokenBorrow ? _amount : 0;

        // Pass data as bytes so that the `swap` function knows it is a flashloan
        bytes memory data = abi.encode(_tokenBorrow, _amount, msg.sender);
        // Execute the initial swap to get the loan
        IUniswapV2Pair(pair).swap(amount0Out, amount1Out, address(this), data);
    }

    function pancakeCall(address _sender, uint256 _amount0, uint256 _amount1, bytes calldata _data) external {
        // Ensure this request comes from the contract
        address token0 = IUniswapV2Pair(msg.sender).token0();
        address token1 = IUniswapV2Pair(msg.sender).token1();
        address pair = IUniswapV2Factory(PANCAKE_FACTORY).getPair(token0, token1);
        require(msg.sender == pair, "The sender needs to match the pair");
        require(_sender == address(this), "Sender should match this contract");

        // Decode data for calculating the repayment
        (address tokenBorrow, uint256 amount, address msgSender) = abi.decode(_data, (address, uint256, address));
        // Calculate the amount to repay at the end
        uint256 fee = ((amount * 3) / 997) + 1;
        uint256 amountToRepay = amount + fee;

        // DO ARBITRAGE
        // !!!!!!!!!!!!!!!!!!!!!
        uint256 loanAmount = _amount0 > 0 ? _amount0 : _amount1;
        uint256 trade1AcquiredCoin = placeTrade(BUSD, CROX, loanAmount);
        uint256 trade2AcquiredCoin = placeTrade(CROX, CAKE, trade1AcquiredCoin);
        uint256 trade3AcquiredCoin = placeTrade(CAKE, BUSD, trade2AcquiredCoin);

        // CHECK PROFITABILITY
        // !!!!!!!!!!!!!!!!!!!!!
        bool isProfitable = checkProfitability(amountToRepay, trade3AcquiredCoin);
//        require(isProfitable, "Arbitrage is not profitable");
        if (isProfitable) {
            IERC20 otherToken = IERC20(BUSD);
            otherToken.transfer(msgSender, trade3AcquiredCoin - amountToRepay);
        }

        // PAY YOURSELF
        // Pay loan back
        IERC20(tokenBorrow).safeTransfer(pair, amountToRepay);
    }

    function checkProfitability(uint256 input, uint256 output) private pure returns (bool) {
        return input < output;
    }

    receive() external payable {}
}
