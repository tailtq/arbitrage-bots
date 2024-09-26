const { expect } = require('chai');
const { ethers, waffle } = require('hardhat');
const { impersonateFundErc20 } = require('../utils/utilities');
const { abi } = require('../artifacts/contracts/interfaces/IERC20.sol/IERC20.json');

const provider = waffle.provider;

describe('ArbitrageExecutor', function () {
  let arbitrageExecutor;
  let owner;
  let addr1;
  let addr2;
  const DECIMALS = 18;

  // PancakeSwap addresses on BSC
  const PANCAKE_FACTORY = '0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73';
  const PANCAKE_ROUTER = '0x10ED43C718714eb63d5aA57B78B54704E256024E';
  const UNISWAP_V2_FACTORY = '0x0000000000000000000000000000000000000000';
  const UNISWAP_V2_ROUTER = '0x0000000000000000000000000000000000000000';

  // Some popular token addresses on BSC
  const WBNB = '0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c';
  const BUSD = '0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56';
  const CAKE = '0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82';
  const BUSD_WHALE_ADDRESS = '0x489a8756c18c0b8b24ec2a2b9ff3d4d447f79bec';
  const initialFundHuman = '100';
  const initialFundHumanDecimals = ethers.utils.parseUnits('100', DECIMALS);
  const busdToken = new ethers.Contract(BUSD, abi, provider);

  beforeEach(async function () {
    // Deploy ArbitrageExecutor
    const ArbitrageExecutor = await ethers.getContractFactory('ArbitrageExecutor');
    [owner, addr1, addr2] = await ethers.getSigners();
    arbitrageExecutor = await ArbitrageExecutor.deploy(
      PANCAKE_FACTORY, PANCAKE_ROUTER, UNISWAP_V2_FACTORY, UNISWAP_V2_ROUTER
    );
    await arbitrageExecutor.deployed();
    await impersonateFundErc20(busdToken, BUSD_WHALE_ADDRESS, arbitrageExecutor.address, initialFundHuman);
  });

  describe('Deployment', function () {
    it('Should set the right owner', async function () {
      expect(await arbitrageExecutor.owner()).to.equal(owner.address);
    });

    it('Should set the correct PancakeSwap addresses', async function () {
      expect(await arbitrageExecutor.PANCAKE_FACTORY()).to.equal(PANCAKE_FACTORY);
      expect(await arbitrageExecutor.PANCAKE_ROUTER()).to.equal(PANCAKE_ROUTER);
    });
  });

  describe('Ownership', function () {
    it('Should allow owner to transfer ownership', async function () {
      await arbitrageExecutor.transferOwnership(addr1.address);
      expect(await arbitrageExecutor.owner()).to.equal(addr1.address);
    });

    it('Should not allow non-owner to transfer ownership', async function () {
      await expect(arbitrageExecutor.connect(addr1).transferOwnership(addr2.address))
        .to.be.revertedWith('Only owner can call this function');
    });
  });

  describe('Token Management', function () {
    it('Should get correct token balance', async function () {
      const BUSDContract = new ethers.Contract(BUSD, abi, provider);
      const balance = await arbitrageExecutor.getBalanceOfToken(BUSD);
      expect(balance).to.equal(await BUSDContract.balanceOf(arbitrageExecutor.address));
    });
  });

  describe('Arbitrage Execution', function () {
    it('Should execute a non-profitable arbitrage', async function () {
      // Assuming we've set up an arbitrage opportunity between BUSD, WBNB, and CAKE
      const swapParams = [
        {
          protocol: 0, // PancakeSwap
          tokenIn: BUSD,
          tokenOut: WBNB,
          fee: 0 // Not used for PancakeSwap
        },
        {
          protocol: 0, // PancakeSwap
          tokenIn: WBNB,
          tokenOut: CAKE,
          fee: 0
        },
        {
          protocol: 0, // PancakeSwap
          tokenIn: CAKE,
          tokenOut: BUSD,
          fee: 0
        },
      ];
      const amountIn = ethers.utils.parseEther('1'); // 1 WBNB

      try {
        await arbitrageExecutor.swapIn(swapParams, amountIn, '0x2c094F5A7D1146BB93850f629501eB749f6Ed491')
      } catch (error) {
        expect(error.toString()).to.contains('Arbitrage not profitable');
      }
    });
  });
});