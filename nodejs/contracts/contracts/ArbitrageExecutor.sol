// SPDX-License-Identifier: MIT

pragma solidity >=0.8.18;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "./interfaces/IUniswapV2Factory.sol";
import "./interfaces/IUniswapV2Router01.sol";
import "./interfaces/IUniswapV2Pair.sol";


contract Ownable {
    address public owner;

    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }

    function transferOwnership(address _newOwner) onlyOwner public {
        require(_newOwner != address(0), "Invalid new owner address");
        owner = _newOwner;
    }

    function withdrawNativeToken() onlyOwner public {
        require(address(this).balance > 0, "Insufficient balance");
        payable(owner).transfer(address(this).balance);
    }
}

contract ArbitrageExecutor is Ownable {
    using SafeERC20 for IERC20;

    uint256 private constant MAX_UINT = 115792089237316195423570985008687907853269984665640564039457584007913129639935;
    address public immutable PANCAKE_FACTORY;
    address public immutable PANCAKE_ROUTER;
    address public immutable UNISWAP_V2_FACTORY;
    address public immutable UNISWAP_V2_ROUTER;

    struct SwapParams {
        uint8 protocol;  // 0: PancakeSwap, 1: UniswapV2, 2: UniswapV3
        address tokenIn;
        address tokenOut;
        uint24 fee;  // Only for UniswapV3
    }

    constructor(address _pancakeFactory, address _pancakeRouter, address _uniswapV2Factory, address _uniswapV2Router) {
        owner = msg.sender;
        PANCAKE_FACTORY = _pancakeFactory;
        PANCAKE_ROUTER = _pancakeRouter;
        UNISWAP_V2_FACTORY = _uniswapV2Factory;
        UNISWAP_V2_ROUTER = _uniswapV2Router;
    }

    function swapIn(SwapParams[] calldata paramsArray, uint256 amountIn, address flashloanToken1) public {
        // swapIn with Flashloan, remember to set allowance for the tokens
        require(paramsArray.length > 0, "Empty params array");
        SwapParams calldata swapParams = paramsArray[0];
        address factoryAddress = swapParams.protocol == 0 ? PANCAKE_FACTORY : UNISWAP_V2_FACTORY;

        if (swapParams.protocol == 0 || swapParams.protocol == 1) {
            address pair = IUniswapV2Factory(factoryAddress).getPair(swapParams.tokenIn, flashloanToken1);
            require(pair != address(0), "This pool does not exist");
            (uint256 amount0Out, uint256 amount1Out) = IUniswapV2Pair(pair).token0() == swapParams.tokenIn
                ? (amountIn, uint256(0))
                : (uint256(0), amountIn);
            // Execute the initial swap to get the loan
            IUniswapV2Pair(pair).swap(
                amount0Out,
                amount1Out,
                address(this),
                abi.encode(paramsArray, amountIn)
            );
        }
    }

    function pancakeCall(address _sender, uint256 _amount0, uint256 _amount1, bytes calldata _data) external {
        // Ensure this request comes from the pair pool contract
        address token0 = IUniswapV2Pair(msg.sender).token0();
        address token1 = IUniswapV2Pair(msg.sender).token1();
        address pair = IUniswapV2Factory(PANCAKE_FACTORY).getPair(token0, token1);
        require(msg.sender == pair, "The sender needs to match the pair");
        require(_sender == address(this), "Sender should match this contract");

        (SwapParams[] memory paramsArray, uint256 amountIn) = abi.decode(_data, (SwapParams[], uint256));
        // Calculate the amount to repay at the end
        uint256 fee = ((amountIn * 3) / 997) + 1;
        uint256 amountToRepay = amountIn + fee;
        uint256 loanAmount = _amount0 > 0 ? _amount0 : _amount1;
        uint256 amountOut = placeTradeUniswapV2(paramsArray, loanAmount, PANCAKE_ROUTER);
        require(amountOut > amountToRepay, "Arbitrage not profitable");
        // Pay loan back
        IERC20(paramsArray[0].tokenIn).safeTransfer(pair, amountToRepay);
    }

    function uniswapV2Call(address _sender, uint256 _amount0, uint256 _amount1, bytes calldata _data) external {
        // Ensure this request comes from the pair pool contract
        address token0 = IUniswapV2Pair(msg.sender).token0();
        address token1 = IUniswapV2Pair(msg.sender).token1();
        address pair = IUniswapV2Factory(UNISWAP_V2_FACTORY).getPair(token0, token1);
        require(msg.sender == pair, "The sender needs to match the pair");
        require(_sender == address(this), "Sender should match this contract");

        (SwapParams[] memory paramsArray, uint256 amountIn) = abi.decode(_data, (SwapParams[], uint256));
        // Calculate the amount to repay at the end
        uint256 fee = ((amountIn * 3) / 997) + 1;
        uint256 amountToRepay = amountIn + fee;
        uint256 loanAmount = _amount0 > 0 ? _amount0 : _amount1;
        uint256 amountOut = placeTradeUniswapV2(paramsArray, loanAmount, UNISWAP_V2_ROUTER);
        require(amountOut > amountToRepay, "Arbitrage not profitable");
        // Pay loan back
        IERC20(paramsArray[0].tokenIn).safeTransfer(pair, amountToRepay);
    }

    function placeTradeUniswapV2(
        SwapParams[] memory paramsArray,
        uint256 _amountIn,
        address _router
    ) private returns (uint256) {
        // Form the path to trade
        // USDT -> BTC, BTC -> ETH, ETH -> USDT => Path is USDT -> BTC -> ETH -> USDT
        address[] memory path = new address[](paramsArray.length + 1);
        path[0] = paramsArray[0].tokenIn;

        for (uint8 i = 0; i < paramsArray.length;) {
            path[i + 1] = paramsArray[i].tokenOut;
            unchecked {
                i++;
            }
        }
        checkAndSetAllowances(path, _amountIn, _router);

        IUniswapV2Router01 router1 = IUniswapV2Router01(_router);
        uint256 amountOutMin = router1.getAmountsOut(_amountIn, path)[1];
        uint256 deadline = block.timestamp + 30 minutes;
        uint amountReceived = router1.swapExactTokensForTokens(_amountIn, amountOutMin, path, address(this), deadline)[1];
        require(amountReceived > 0, "Aborted Tx: Trade returned zero");

        return amountReceived;
    }

    function checkAndSetAllowances(address[] memory _tokens, uint256 _amountIn, address _router) internal {
        for (uint8 i = 0; i < _tokens.length;) {
            if (IERC20(_tokens[i]).allowance(address(this), _router) < _amountIn) {
                IERC20(_tokens[i]).safeIncreaseAllowance(_router, MAX_UINT);
            }
            unchecked {
                i++;
            }
        }
    }

    function getBalanceOfToken(address _token) public view returns (uint256) {
        return IERC20(_token).balanceOf(address(this));
    }

    function withdrawNonNativeToken(address _token) onlyOwner public {
        IERC20 token = IERC20(_token);
        require(token.balanceOf(address(this)) > 0, "Insufficient balance");
        token.safeTransfer(owner, IERC20(_token).balanceOf(address(this)));
    }
}
