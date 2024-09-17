const { expect, assert } = require('chai');
const { ethers, waffle, network } = require('hardhat');
const { impersonateFundErc20 } = require('../utils/utilities');

const { abi } = require('../artifacts/contracts/interfaces/IERC20.sol/IERC20.json');
const provider = waffle.provider;

describe('FlashSwap contract', async () => {
  let FLASHSWAP, BORROW_AMOUNT, FUND_AMOUNT, initialFundHuman, txArbitrage, gasUsedUSD;
  const DECIMALS = 18;
  const BUSD_WHALE_ADDRESS = '0x489a8756c18c0b8b24ec2a2b9ff3d4d447f79bec';
  const WBNB_ADDRESS = '0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c';
  const BUSD_ADDRESS = '0xe9e7cea3dedca5984780bafc599bd69add087d56';
  const CAKE_ADDRESS = '0x0e09fabb73bd3ade0a17ecc321fd13a19e81ce82';
  const USDT_ADDRESS = '0x55d398326f99059ff775485246999027b3197955';
  const CROX_ADDRESS = '0x2c094f5a7d1146bb93850f629501eb749f6ed491';
  const BASE_TOKEN_ADDRESS = BUSD_ADDRESS;
  const baseToken = new ethers.Contract(BASE_TOKEN_ADDRESS, abi, provider);

  beforeEach(async () => {
    [owner] = await ethers.getSigners();

    // Ensure that the WHALE has a balance
    const whaleBalance = await provider.getBalance(BUSD_WHALE_ADDRESS);
    expect(whaleBalance).not.equal('0');

    // Deploy smart contract
    const FlashSwap = await ethers.getContractFactory('PancakeFlashSwap');
    FLASHSWAP = await FlashSwap.deploy();
    await FLASHSWAP.deployed();

    // Configure our Borrowing
    const borrowAmountHuman = '1';
    BORROW_AMOUNT = ethers.utils.parseUnits(borrowAmountHuman, DECIMALS);

    // Configure Funding - FOR TESTING ONLY
    initialFundHuman = '100';
    FUND_AMOUNT = ethers.utils.parseUnits(initialFundHuman, DECIMALS);

    // Fund our Contract - FOR TESTING ONLY
    await impersonateFundErc20(baseToken, BUSD_WHALE_ADDRESS, FLASHSWAP.address, initialFundHuman);
  });

  describe('Arbitrage Execution', () => {
    it('ensures the contract is funded', async () => {
      const flashSwapBalance = await FLASHSWAP.getBalanceOfToken(BASE_TOKEN_ADDRESS);
      const flashSwapBalanceHuman = ethers.utils.formatUnits(flashSwapBalance, DECIMALS);
      expect(Number(flashSwapBalanceHuman)).equal(Number(initialFundHuman));
    });

    it('executes the arbitrage', async () => {
      txArbitrage = await FLASHSWAP.startArbitrage(BASE_TOKEN_ADDRESS, BORROW_AMOUNT);
      assert(txArbitrage);

      const contractBalanceBUSD = await FLASHSWAP.getBalanceOfToken(BASE_TOKEN_ADDRESS);
      const contractBalanceBUSDHuman = ethers.utils.formatUnits(contractBalanceBUSD, DECIMALS);
      const contractBalanceCROX = await FLASHSWAP.getBalanceOfToken(CROX_ADDRESS);
      const contractBalanceCROXHuman = ethers.utils.formatUnits(contractBalanceCROX, DECIMALS);
      const contractBalanceCAKE = await FLASHSWAP.getBalanceOfToken(CAKE_ADDRESS);
      const contractBalanceCAKEHuman = ethers.utils.formatUnits(contractBalanceCAKE, DECIMALS);
      console.log('contractBalanceBUSDHuman', contractBalanceBUSDHuman);
      console.log('contractBalanceCROXHuman', contractBalanceCROXHuman);
      console.log('contractBalanceCAKEHuman', contractBalanceCAKEHuman);
    });
  });

  // it('general test', async () => {
  //   const whaleBalance = await provider.getBalance(BUSD_WHALE_ADDRESS);
  //   console.log(ethers.utils.formatUnits(whaleBalance, DECIMALS));
  //   console.log('test hehe');
  // });
});